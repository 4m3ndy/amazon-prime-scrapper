//Package cmd is responsible for initializing grpc and http servers.
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/service"
)

func serve(ctx context.Context) (err error) {
	rtr := mux.NewRouter()

	// Health Route
	rtr.HandleFunc("/health", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "OK")
		},
	))

	// Amazon Movie Route
	rtr.HandleFunc("/movie/amazon/{amazon_id:[A-Za-z0-9]+}", AmazonMovieHandler).Methods("GET")

	serverPort, exists := os.LookupEnv("AMAZON_SCRAPPER_SVC_HTTP_PORT")
	if !exists {
		logger.Log().Panicf("Failed to run, please define AMAZON_SCRAPPER_SVC_HTTP_PORT env variable")
	}

	srv := &http.Server{
		Addr:    ":" + serverPort,
		Handler: rtr,
	}

	go func() {
		// Get service listening port
		addrHTTP := ":" + os.Getenv("AMAZON_SCRAPPER_SVC_HTTP_PORT")
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log().Fatalf("Failed to listen: %v", err)
		} else {
			logger.Log().Warnf("HTTP Service will stop listening on %s", addrHTTP)
		}
	}()

	<-ctx.Done()

	logger.Log().Warnf("Server Stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		logger.Log().Fatalf("Server Shutdown Failed:%+s", err)
	} else {
		logger.Log().Infof("Server Exited Properly")
	}

	if err == http.ErrServerClosed {
		err = nil
	}

	return
}

func RunServer() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		logger.Log().Infof("System Call %#v received", oscall)
		cancel()
	}()

	if err := serve(ctx); err != nil {
		logger.Log().Errorf("Failed to serve: %#v", err)
	}
}

// Amazon Movie Route Handler
func AmazonMovieHandler(w http.ResponseWriter, r *http.Request) {
	requestVars := mux.Vars(r)

	movie, err := service.ScrapeMovie(requestVars["amazon_id"])
	if err != nil {
		logger.Log().Errorf("Error scraping the requested movie: %#v", err)
		return
	}

	js, err := json.Marshal(movie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
