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
	"fmt"

	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
)

var (
	shutdownChannel = make(chan bool, 1)
)

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
type server struct {
	config       ServerConfig
	httpServer   *http.Server
	httpListener net.Listener
	httpShutdown Shutdownable
}

type ServerConfig struct {
	HttpPort           string
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
	// Setup the server
	setupServer(srv)
	fmt.Println("setupServer")

	// Listening for shutdown signal
	listenToShutdownSignal(srv)
	fmt.Println("listenToShutdownSignal")

	// Start the listening on each protocol
	serveHTTP(srv)
	fmt.Println("serveHTTP")
}

func listenToShutdownSignal(srv *server) {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		logger.Log().Warningf("Shutdown Signal has been received: %s", sig)
		logger.Log().Warningln("Gracefully Shutting Down has been initiated!")

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
	logger.Log().Debugf("Shutdown HTTP server with timeout: %s", timeout)

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
}

func setupServer(srv *server) {

	// Get service listening port
	addrHttp := ":" + srv.config.HttpPort

	// Setup a listener for HTTP port
	httpL, err := net.Listen("tcp", addrHttp)
	if err != nil {
		logger.Log().Fatalf("Failed to listen on %s: %v", addrHttp, err)
	} else {
		logger.Log().Infof("HTTP service starting on %s", addrHttp)
	}
	srv.httpListener = httpL
}

func serveHTTP(srv *server) {
	fmt.Println("serveHTTP 2")
	if err := srv.httpServer.Serve(srv.httpListener); !strings.Contains(err.Error(), "closed") {
		logger.Log().Fatalf("error while serving http request: %#v", err)
		shutdownChannel <- true
	}
}
