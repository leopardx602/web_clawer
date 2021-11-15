package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

type Product struct {
	Name       string
	Price      int
	ImageURL   string
	ProductURL string
}

type Web struct {
	c       *colly.Collector
	jobs    chan string
	keyword string
	webName string
}

// Parse the html depends on the website.
func (w *Web) Parse(products chan Product) {
	switch w.webName {
	case "yahoo":
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
		// case "pchome":
	}
}

// Create the jobs depend on the format of url.
func (w *Web) CreateJobs() {
	switch w.webName {
	case "yahoo":
		for i := 0; i < 5; i++ {
			w.jobs <- fmt.Sprintf("https://tw.buy.yahoo.com/search/product?p=%s&pg=%v", w.keyword, i)
		}
		close(w.jobs)
	}
	// case "pchome":
}

// Build a clawer on each web site, and use worker(gorutine) to get each page.
func Clawer(web *Web, products chan Product, wg *sync.WaitGroup) {
	finishJobs := make(chan int, 5)

	web.Parse(products) // different
	web.c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	web.c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		finishJobs <- 1
		if len(finishJobs) == 5 {
			time.Sleep(time.Second)
			wg.Done()
		}
	})

	// Assign the jobs to the workers
	worker := func(id int, jobs <-chan string) {
		for url := range jobs {
			web.c.Visit(url)
			time.Sleep(1 * time.Second)
		}
	}
	for w := 1; w <= 3; w++ {
		go worker(w, web.jobs)
	}

	// Create the jobs
	web.CreateJobs() // different
}

// Looking for on different web sites. Continuous data output when found.
func WebClawer(keyword string) {
	products := make(chan Product, 100)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	web := Web{c: colly.NewCollector(), keyword: keyword, jobs: make(chan string, 5), webName: "yahoo"}
	go Clawer(&web, products, wg)

	// web := Web{c: colly.NewCollector(), keyword: keyword, jobs: make(chan string, 5), webName: "pchome"}
	// go Clawer(&web, products, wg)

	// Output
	go func() {
		for product := range products {
			fmt.Println(product)
		}
	}()

	wg.Wait()
	close(products)
}

func main() {
	WebClawer("iphone")
	fmt.Println("Done")
}
