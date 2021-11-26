package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type Yahoo struct {
	c       *colly.Collector
	jobs    chan string
	keyword string
	webName string
}

// Parse the html depends on the website.
func (w *Yahoo) Parse(products chan Product) {

	w.c.OnHTML(".BaseGridItem__grid___2wuJ7", func(e *colly.HTMLElement) {
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

// Create the jobs depend on the format of url.
func (w *Yahoo) CreateJobs() {
	for i := 0; i < 5; i++ {
		w.jobs <- fmt.Sprintf("https://tw.buy.yahoo.com/search/product?p=%s&pg=%v", w.keyword, i)
	}
	close(w.jobs)
}
