package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"encoding/json"

	"github.com/gocolly/colly/v2"
	//"github.com/gocolly/colly/v2/debug"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
)

type Movie struct {
	Title   	string `json:"title"`
	ReleaseYear string `json:"release_year"`
	Actors		[]string `json:"actors"`
	Poster		string `json:"poster"`
	SimilarIds 	[]string `json:"similar_ids"`
}

// ScrapeMovie returns an access and refresh token upon a successful login ofr the lr connect app
func ScrapeMovie(s *string) (*Movie, error) {
	fmt.Println("Visiting", r.URL.String())
	return nil, nil
}
