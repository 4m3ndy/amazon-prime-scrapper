//
// Server implementation shared between all microservices.
// If this file is changed it will affect _all_ microservices in the monorepo (and this
// is deliberately so).
//
package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
)

var (
	shutdownChannel = make(chan bool, 1)
)

// RegisterHTTP ...
// Callback type for registering http endpoints
type RegisterHTTP func(*http.ServeMux) Shutdownable

type Shutdownable interface {
	Shutdown()
}

type CancelFunc context.CancelFunc

func (cf CancelFunc) Shutdown() {
	cf()
}

type ShutdownAll []Shutdownable

func (s ShutdownAll) Shutdown() {
	for _, shutdownable := range s {
		shutdownable.Shutdown()
	}
}

func init() {
	logger.CreateLogger()
}

// Server ...
// Server structure that serves the health endpoints and
// grpc endpoints
type server struct {
	config       ServerConfig
	httpServer   *http.Server
	httpListener net.Listener
	httpShutdown Shutdownable
}

type ServerConfig struct {
	//if set to true, the http server will be automatically created as a proxy to grpc
	AutoHttpGrpcProxy  bool
	GrpcPort           string
	HttpPort           string
	RegisterGrpcServer RegisterGrpc
	GrpcServerOptions  []grpc.ServerOption
	RegisterHTTPServer RegisterHTTP
	// If AutoHttpGrpcProxy is set to true, only RegisterAutoHttpGrpcProxy has to be supplied, if it false RegisterHTTP can be supplied
	RegisterAutoHttpGrpcProxy RegisterAutoHttpGrpcProxy
}

// NewServer(config).RunServer()
func RunConfig(config ServerConfig) {
	NewServer(config).RunServer()
}

func NewServer(config ServerConfig) *server {
	return &server{config: config}
}

// RunServer ...
// start a server and initialize it with the given config
func (srv *server) RunServer() {

	if srv.config.RegisterGrpcServer == nil && srv.config.RegisterHTTPServer == nil {
		panic("either GRPC or HTTP server register have to be defined")
	}

	// Setup the server
	setupServer(srv)

	// Listening for shutdown signal
	listenToShutdownSignal(srv)

	// Start the listening on each protocol
	if srv.config.RegisterGrpcServer != nil && srv.config.RegisterHTTPServer != nil || srv.config.AutoHttpGrpcProxy {
		go serveHTTP(srv)
		serveGrpc(srv)
	} else if srv.config.RegisterGrpcServer == nil {
		serveHTTP(srv)
	} else if srv.config.RegisterHTTPServer == nil {
		serveGrpc(srv)
	}
}

func listenToShutdownSignal(srv *server) {
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

func gracefulShutdown(srv *server, timeout time.Duration) {
	// Instantiate background context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	// Gracefully shutdown http server
	logger.Log().Debugf("shutdown http server with timeout: %s", timeout)

	if srv.httpServer != nil {
		if err := srv.httpServer.Shutdown(ctx); err != nil {
			logger.Log().Fatalf("http server couldn't gracefully shutdown %v", err)
		} else {
			logger.Log().Debugf("http server has been gracefully shutdown!")
		}

		if srv.httpShutdown != nil {
			srv.httpShutdown.Shutdown()
		}
	}

	if srv.grpcServer != nil {
		logger.Log().Debugf("gRPC server shutting down gracefully ...")
		// Gracefully shutdown gRPC server
		go srv.grpcServer.GracefulStop()
		logger.Log().Debugf("gRPC server has been gracefully shutdown!")
		if srv.grpcShutdown != nil {
			srv.grpcShutdown.Shutdown()
		}

	}
}

func setupServer(srv *server) {

	if srv.config.RegisterHTTPServer != nil && srv.config.RegisterAutoHttpGrpcProxy != nil {
		panic("auto http grpc proxy and normal http cannot be defined together")
	}

	// Get service listening port
	addrGrpc := ":" + srv.config.GrpcPort
	addrHttp := ":" + srv.config.HttpPort

	if srv.config.RegisterGrpcServer != nil {
		// Setup a listener for gRPC port
		grpcL, err := net.Listen("tcp", addrGrpc)
		if err != nil {
			logger.Log().Fatalf("failed to listen on %s: %v", addrGrpc, err)
		} else {
			logger.Log().Infof("GRPC service starting on %s", addrGrpc)
		}
		srv.grpcListener = grpcL
	}

	if srv.config.RegisterHTTPServer != nil || srv.config.AutoHttpGrpcProxy {
		// Setup a listener for HTTP port
		httpL, err := net.Listen("tcp", addrHttp)
		if err != nil {
			logger.Log().Fatalf("failed to listen on %s: %v", addrHttp, err)
		} else {
			logger.Log().Infof("HTTP service starting on %s", addrHttp)
		}
		srv.httpListener = httpL
	}

	if srv.config.RegisterGrpcServer != nil {
		var grpcS *grpc.Server

		// Instantiate gRPC server
		if srv.config.GrpcServerOptions != nil {
			grpcS = grpc.NewServer(srv.config.GrpcServerOptions...)

		} else {
			grpcS = grpc.NewServer()

		}
		// Register gRPC service
		srv.grpcShutdown = srv.config.RegisterGrpcServer(grpcS)

		srv.grpcServer = grpcS
	}

	if !srv.config.AutoHttpGrpcProxy {
		// Instantiate regular HTTP server
		if srv.config.RegisterHTTPServer != nil {
			mux := http.NewServeMux()
			srv.httpShutdown = srv.config.RegisterHTTPServer(mux)
			httpS := &http.Server{
				Handler: mux,
			}
			srv.httpServer = httpS
		}
	} else {
		logger.Log().Debugf("Creating Http -> GRPC Proxy")
		mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(runtime.DefaultHeaderMatcher))
		srv.config.RegisterAutoHttpGrpcProxy(mux)
		httpS := &http.Server{
			Handler: mux,
		}
		srv.httpServer = httpS
	}
}

func serveGrpc(srv *server) {
	if err := srv.grpcServer.Serve(srv.grpcListener); err != nil {
		logger.Log().Fatalf("error while serving gRPC request: %#v", err)
		shutdownChannel <- true
	}
}

func serveHTTP(srv *server) {
	if err := srv.httpServer.Serve(srv.httpListener); !strings.Contains(err.Error(), "closed") {
		logger.Log().Fatalf("error while serving http request: %#v", err)
		shutdownChannel <- true
	}
}

// ToGrpcResponseErr converts an error to a proper grpc error with a
// proper error code
func ToGrpcResponseErr(msg string, err error) error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	code := grpcCodes.Internal
	if ok {
		code = s.Code()
	}
	return status.Errorf(code, msg, err)
}
