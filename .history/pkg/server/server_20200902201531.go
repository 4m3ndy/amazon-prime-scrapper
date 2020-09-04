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
	HTTPPort     string
	httpServer   *http.Server
	httpListener net.Listener
}

// RunServer ...
// start a server and initialize the grpc endpoints via the given
// callback
func (srv *Server) RunServer(registerHTTPServer RegisterHTTP) {
	srv.runServerInternal(registerHTTPServer, config, true)
}

// RunServerWithoutRequestAuthentication ...
// start a server and initialize the grpc endpoints via the given
// callback. The individual endpoints don't validated the passed token.
// This is needed for services that provide endpoints that don't enforce authentication (e.g. the auth or tracking-service)
func (srv *Server) RunServerWithoutRequestAuthentication(registerGrpcServer RegisterGrpc, registerHTTPServer RegisterHTTP, config ServerConfig) {
	srv.runServerInternal(registerGrpcServer, registerHTTPServer, config, false)
}

func (srv *Server) runServerInternal(registerHTTPServer RegisterHTTP, config ServerConfig) {
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
}

func setupServer(srv *Server, registerHTTP RegisterHTTP) {
	// Get service listening port
	addrHTTP := ":" + srv.HTTPPort

	// Setup a listener for HTTP port
	httpL, err := net.Listen("tcp", addrHTTP)
	if err != nil {
		logger.Log().Fatalf("failed to listen on %s: %v", addrHTTP, err)
	} else {
		logger.Log().Infof("http service starting on %s", addrHTTP)
	}

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

	srv.httpServer = httpS
	srv.httpListener = httpL
}

func serveHTTP(srv *Server) {
	if err := srv.httpServer.Serve(srv.httpListener); !strings.Contains(err.Error(), "closed") {
		logger.Log().Fatalf("error while serving http request: %#v", err)
		shutdownChannel <- true
	}
}
