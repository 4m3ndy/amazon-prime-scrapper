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

	"github.com/gorilla/mux"
	"github.com/gocolly/colly/v2"
	//"github.com/gocolly/colly/v2/debug"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
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

	c := colly.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	// On every a element which has href attribute call callback
	c.OnHTML("span[data-automation-id=release-year-badge]", func(e *colly.HTMLElement) {
		fmt.Printf("Release found: %q\n", e.Text)
	})

	// On every a element which has href attribute call callback
	c.OnHTML("h1[data-automation-id=title]", func(e *colly.HTMLElement) {
		fmt.Printf("Title found: %q\n", e.Text)
	})

	// On every a element which has href attribute call callback
	c.OnHTML("div.dv-fallback-packshot-image", func(e *colly.HTMLElement) {
		posters := strings.Split(e.ChildAttr("img", "srcset"), ",")
		fmt.Printf("Poster: %v\n", posters[0])
		
	})

	// On every a element which has href attribute call callback
	c.OnHTML("div[data-automation-id=meta-info]", func(e *colly.HTMLElement) {
		actors := e.ChildText("div dl:nth-of-type(2) dd")
		fmt.Printf("Actors: %v\n", actors)
	})

	// On every a element which has href attribute call callback
	c.OnHTML("div.DVWebNode-detail-btf-wrapper", func(e *colly.HTMLElement) {
		var similarIDs []string		
		e.ForEach("ul li a", func(_ int, elem *colly.HTMLElement) {
			id = strings.Split(elem.Attr("href"), "/")
			similarIDs = append(s, id[4])
		})
		
		fmt.Printf("Similar IDs: %v\n", similarIDs)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on https://hackerspaces.org
	c.Visit("http://www.amazon.de/gp/product/" + vars["amazon_id"])
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Amazon ID: %v\n", vars["amazon_id"])

}
