//Package cmd is responsible for initializing grpc and http servers.
package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"strings"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/gocolly/colly/v2"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/service"
)

type Movie struct {
	Title   	string `json:"title"`
	ReleaseYear string `json:"release_year"`
	Actors		[]string `json:"actors"`
	Poster		string `json:"poster"`
	SimilarIds 	[]string `json:"similar_ids"`
}

func serve(ctx context.Context) (err error) {

	rtr := mux.NewRouter()
	rtr.HandleFunc("/health", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "OK")
		},
	))

	rtr.HandleFunc("/movie/amazon/{amazon_id:[A-Za-z0-9]+}", AmazonMovieHandler).Methods("GET")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: rtr,
	}

	go func() {
		// Get service listening port
		addrHTTP := ":" + "8080"
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log().Fatalf("failed to listen: %v", err)
		} else {
			logger.Log().Infof("http service starting on %s", addrHTTP)
		}	
	}()

	<-ctx.Done()

	log.Printf("Server Stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Server Shutdown Failed:%+s", err)
	}

	log.Printf("Server Exited Properly")

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
		log.Printf("system call:%+v", oscall)
		cancel()
	}()

	if err := serve(ctx); err != nil {
		log.Printf("failed to serve:+%v\n", err)
	}
}


func AmazonMovieHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	m = service.ScrapeMovie(vars["amazon_id"])

	js, err := json.Marshal(movie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	service.ScrapeMovie(vars["amazon_id"])
	
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "Amazon ID: %v\n", vars["amazon_id"])
	w.Write(js)

}
