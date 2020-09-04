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
	//"github.com/gocolly/colly/v2/debug"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/service"
)

func serve(ctx context.Context) (err error) {

	rtr := mux.NewRouter()
	rtr.HandleFunc("/health", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "OK")
		},
	))

	rtr.HandleFunc("/movie/amazon/{amazon_id:[A-Za-z0-9]+}", AmazonPrimeScrapperHandler).Methods("GET")

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

	log.Printf("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server Shutdown Failed:%+s", err)
	}

	log.Printf("server exited properly")

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


func AmazonPrimeScrapperHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	movie := new(Movie)
	c := colly.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	// On every a element which has href attribute call callback
	c.OnHTML("span[data-automation-id=release-year-badge]", func(e *colly.HTMLElement) {
		movie.ReleaseYear = e.Text
	})

	// On every a element which has href attribute call callback
	c.OnHTML("h1[data-automation-id=title]", func(e *colly.HTMLElement) {
		movie.Title = e.Text
	})

	// On every a element which has href attribute call callback
	c.OnHTML("div.dv-fallback-packshot-image", func(e *colly.HTMLElement) {
		posters := strings.Split(e.ChildAttr("img", "srcset"), ",")
		movie.Poster = posters[0]
	})

	// On every a element which has href attribute call callback
	c.OnHTML("div[data-automation-id=meta-info]", func(e *colly.HTMLElement) {
		actors := strings.Split(e.ChildText("div dl:nth-of-type(2) dd"), ",")
		movie.Actors = actors
	})

	// On every a element which has href attribute call callback
	c.OnHTML("div.DVWebNode-detail-btf-wrapper", func(e *colly.HTMLElement) {
		var similarIDs []string		
		e.ForEach("ul li a", func(_ int, elem *colly.HTMLElement) {
			id := strings.Split(elem.Attr("href"), "/")[4]
			similarIDs = append(similarIDs, id)
		})

		movie.SimilarIds = similarIDs
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something Went Wrong:", err)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on https://hackerspaces.org
	c.Visit("http://www.amazon.de/gp/product/" + vars["amazon_id"])

	js, err := json.Marshal(movie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}
