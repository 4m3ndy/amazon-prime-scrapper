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
