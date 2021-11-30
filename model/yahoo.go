package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type Yahoo struct {
	Keyword string
}

// Parse the response depends on different websites. (html)
func (w *Yahoo) Parse(collyWorker *colly.Collector, products chan Product) {
	collyWorker.OnHTML(".BaseGridItem__grid___2wuJ7", func(e *colly.HTMLElement) {
		var product Product
		product.Name = e.DOM.Find(".BaseGridItem__title___2HWui").Text()
		product.ImageURL, _ = e.DOM.Find(".SquareImg_img_2gAcq").Attr("src")
		product.ProductURL, _ = e.DOM.Find("a").Attr("href")

		priceInfo := e.DOM.Find(".BaseGridItem__price___31jkj")
		var priceStr string
		if len(priceInfo.Find("em").Text()) > 0 {
			priceStr = priceInfo.Find("em").Text()
		} else {
			priceStr = priceInfo.Text()
		}

		priceStr = strings.ReplaceAll(priceStr, ",", "")
		priceStr = strings.ReplaceAll(priceStr, "$", "")
		price, err := strconv.Atoi(priceStr)
		if err != nil {
			fmt.Println(err)
		}
		product.Price = price
		products <- product
	})
}

// Create the jobs depends on the URL of different websites.
// Each page contains 6 items.
func (w *Yahoo) CreateJobs(jobs chan string) {
	for i := 0; i < 5; i++ {
		jobs <- fmt.Sprintf("https://tw.buy.yahoo.com/search/product?p=%s&pg=%v", w.Keyword, i)
	}
}
