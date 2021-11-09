package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Product struct {
	Name       string
	Price      int
	ImageURL   string
	ProductURL string
}

func main() {
	jobs := make(chan string, 5)
	finishJobs := make(chan int, 5)
	products := make(chan Product, 100)

	c := colly.NewCollector()

	c.OnHTML(".BaseGridItem__grid___2wuJ7", func(e *colly.HTMLElement) {
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

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		finishJobs <- 1
		if len(finishJobs) == 5 {
			time.Sleep(time.Second)
			close(products)
		}
	})

	// Assign the jobs to the workers
	worker := func(id int, jobs <-chan string) {
		for url := range jobs {
			c.Visit(url)
			time.Sleep(1 * time.Second)
		}
	}
	for w := 1; w <= 3; w++ {
		go worker(w, jobs)
	}

	// Create the jobs
	for i := 0; i < 5; i++ {
		jobs <- fmt.Sprintf("https://tw.buy.yahoo.com/search/product?p=iphone&pg=%v", i)
	}
	close(jobs)

	// Result
	for product := range products {
		fmt.Println(product)
	}
	fmt.Println("Done")
}
