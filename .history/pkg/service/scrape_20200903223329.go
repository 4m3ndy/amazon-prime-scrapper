package service

import (
	"fmt"
	"net/http"
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

// ScrapeMovie returns an access and refresh token upon a successful login ofr the lr connect app
func ScrapeMovie(s string) (string, error) {
	fmt.Println("ScrapeMovie")
	return "Yes", nil
}
