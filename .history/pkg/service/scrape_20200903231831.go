package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly/v2"
	// "github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
)

type Movie struct {
	Title   	string `json:"title"`
	ReleaseYear string `json:"release_year"`
	Actors		[]string `json:"actors"`
	Poster		string `json:"poster"`
	SimilarIds 	[]string `json:"similar_ids"`
}

// ScrapeMovie returns an access and refresh token upon a successful login ofr the lr connect app
func ScrapeMovie(s string) (*Movie, error) {
	movie := new(Movie)
	c := colly.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	// Scrape Movie Release Year
	c.OnHTML("span[data-automation-id=release-year-badge]", func(e *colly.HTMLElement) {
		movie.ReleaseYear = e.Text
	})

	// Scrape Movie Title
	c.OnHTML("h1[data-automation-id=title]", func(e *colly.HTMLElement) {
		movie.Title = e.Text
	})

	// Scrape Movie Poster
	c.OnHTML("div.dv-fallback-packshot-image", func(e *colly.HTMLElement) {
		posters := strings.Split(e.ChildAttr("img", "srcset"), ",")
		movie.Poster = posters[0]
	})

	// Scrape Movie Actors
	c.OnHTML("div[data-automation-id=meta-info]", func(e *colly.HTMLElement) {
		actors := strings.Split(e.ChildText("div dl:nth-of-type(2) dd"), ",")
		movie.Actors = actors
	})

	// Scrape Movie Similar IDs
	c.OnHTML("div.DVWebNode-detail-btf-wrapper", func(e *colly.HTMLElement) {
		var similarIDs []string		
		e.ForEach("ul li a", func(_ int, elem *colly.HTMLElement) {
			id := strings.Split(elem.Attr("href"), "/")[4]
			similarIDs = append(similarIDs, id)
		})

		movie.SimilarIds = similarIDs
	})

	c.OnError(func(_ *colly.Response, err error) {
		logger.Log().Errorf("Something Went Wrong: %#v", err)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		logger.Log().Infof("Visiting: %#v", r.URL.String())
	})

	// Start scraping on https://www.amazon.de/gp/product/MovieID
	c.Visit("https://www.amazon.de/gp/product/" + s)

	return movie, nil
}
