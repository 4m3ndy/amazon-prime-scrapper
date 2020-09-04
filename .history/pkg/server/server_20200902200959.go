package cmd

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/4m3ndy/amazon-prime-scrapper/logger"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	shutdownChannel = make(chan bool, 1)
)

// ServerConfig the config for the grpc and http server
type ServerConfig struct {
	ServerEnvConfig
}

// ServerEnvConfig the server config that can be configured with env variables
type ServerEnvConfig struct {
	ProjectID               string `env:"GOOGLE_PROJECT_ID,required"`
	EnvironmentName         string `env:"ENV_NAME,required"`
	EnableTracingAndMetrics bool   `env:"ENABLE_TRACING_AND_METRICS" envDefault:"true"`
}

// RegisterGrpc ...
// Callback type for registering grpc endpoints
type RegisterGrpc func(*grpc.Server)

// RegisterHTTP ...
// Callback type for registering http endpoints
type RegisterHTTP func(*http.ServeMux)

func init() {
	logger.CreateLogger()
}

// Server ...
// Server structure that serves the health endpoints and
// grpc endpoints
type Server struct {
	GrpcPort     string
	HTTPPort     string
	grpcServer   *grpc.Server
	httpServer   *http.Server
	grpcListener net.Listener
	httpListener net.Listener
}

// RunServer ...
// start a server and initialize the grpc endpoints via the given
// callback
func (srv *Server) RunServer(registerGrpcServer RegisterGrpc, registerHTTPServer RegisterHTTP, config ServerConfig) {
	srv.runServerInternal(registerGrpcServer, registerHTTPServer, config, true)
}

// RunServerWithoutRequestAuthentication ...
// start a server and initialize the grpc endpoints via the given
// callback. The individual endpoints don't validated the passed token.
// This is needed for services that provide endpoints that don't enforce authentication (e.g. the auth or tracking-service)
func (srv *Server) RunServerWithoutRequestAuthentication(registerGrpcServer RegisterGrpc, registerHTTPServer RegisterHTTP, config ServerConfig) {
	srv.runServerInternal(registerGrpcServer, registerHTTPServer, config, false)
}

func (srv *Server) runServerInternal(registerGrpcServer RegisterGrpc, registerHTTPServer RegisterHTTP, config ServerConfig, enforceRequestAuth bool) {
	// Setup the server
	setupServer(srv, registerGrpcServer, registerHTTPServer, config.EcdsaPublicKey, enforceRequestAuth)

	if config.EnableTracingAndMetrics {
		flush := setupTracingAndMetrics(config.ProjectID, config.EnvironmentName)
		defer flush()
	}

	// Listening for shutdown signal
	listenToShutdownSignal(srv)

	// Start the listening on each protocol
	go serveHTTP(srv)
	serveGrpc(srv)
}

func listenToShutdownSignal(srv *Server) {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		logger.Log().Warningf("shutdown signal has been received: %s", sig)
		logger.Log().Warningln("gracefully shutting down has been initiated!")

		// Wait for a signal to shutdown all servers
		<-shutdownChannel
		gracefulShutdown(srv, 30*time.Second)
	}()

	shutdownChannel <- true
}

func gracefulShutdown(srv *Server, timeout time.Duration) {
	// Instantiate background context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	// Gracefully shutdown http server
	logger.Log().Warningf("shutdown http server with timeout: %s", timeout)

	if err := srv.httpServer.Shutdown(ctx); err != nil {
		logger.Log().Fatalf("http server couldn't gracefully shutdown %v", err)
	} else {
		logger.Log().Warningln("http server has been gracefully shutdown!")
	}

	// Gracefully shutdown gRPC server
	srv.grpcServer.GracefulStop()
	logger.Log().Warningln("gRPC server has been gracefully shutdown!")
}

func setupServer(srv *Server, registerGrpc RegisterGrpc, registerHTTP RegisterHTTP, ecdsaPublicKey *ecdsa.PublicKey, enforceAuth bool) {
	// Get service listening port
	addrGrpc := ":" + srv.GrpcPort
	addrHTTP := ":" + srv.HTTPPort

	// Setup a listener for gRPC port
	grpcL, err := net.Listen("tcp", addrGrpc)
	if err != nil {
		logger.Log().Fatalf("failed to listen on %s: %v", addrGrpc, err)
	} else {
		logger.Log().Infof("grpc service starting on %s", addrGrpc)
	}

	// Setup a listener for HTTP port
	httpL, err := net.Listen("tcp", addrHTTP)
	if err != nil {
		logger.Log().Fatalf("failed to listen on %s: %v", addrHTTP, err)
	} else {
		logger.Log().Infof("http service starting on %s", addrHTTP)
	}

	// Instantiate gRPC server
	grpcS := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc.UnaryInterceptor(createInterceptor(ecdsaPublicKey, enforceAuth)))

	// Register gRPC service
	registerGrpc(grpcS)

	// Instantiate regular HTTP server
	h := http.NewServeMux()

	httpS := &http.Server{
		Handler: h,
	}

	// Handel health check endpoint over HTTP
	h.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "OK")
	})

	registerHTTP(h)

	srv.grpcListener = grpcL
	srv.grpcServer = grpcS
	srv.httpServer = httpS
	srv.httpListener = httpL
}

func setupTracingAndMetrics(projectID string, environmentName string) func() {
	// TODO: configure trace sampling once in production
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	exporter, err := metrics.CreateStackDriverExporter(projectID, environmentName)
	if err != nil {
		logger.Log().WithError(err).Fatal("error initializing stats stackDriverExporter")
	}
	trace.RegisterExporter(exporter)
	if err := exporter.StartMetricsExporter(); err != nil {
		logger.Log().WithError(err).Fatal("error starting metric exporter")
	}
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		log.Fatal(err)
	}

	// Flush must be called before main() exits to ensure metrics are recorded.
	return exporter.Flush
}

func serveGrpc(srv *Server) {
	if err := srv.grpcServer.Serve(srv.grpcListener); err != nil {
		logger.Log().Fatalf("error while serving gRPC request: %#v", err)
		shutdownChannel <- true
	}
}

func serveHTTP(srv *Server) {
	if err := srv.httpServer.Serve(srv.httpListener); !strings.Contains(err.Error(), "closed") {
		logger.Log().Fatalf("error while serving http request: %#v", err)
		shutdownChannel <- true
	}
}
