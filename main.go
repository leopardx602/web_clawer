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

type Yahoo struct {
	c       *colly.Collector
	jobs    chan string
	keyword string
}

func (y *Yahoo) Parse(products chan Product) {
	y.c.OnHTML(".BaseGridItem__grid___2wuJ7", func(e *colly.HTMLElement) {
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

func (y *Yahoo) CreateJobs() {
	// Create the jobs
	for i := 0; i < 5; i++ {
		y.jobs <- fmt.Sprintf("https://tw.buy.yahoo.com/search/product?p=%s&pg=%v", y.keyword, i)
	}
	close(y.jobs)
}

func Clawer(yahoo *Yahoo, products chan Product, wg *sync.WaitGroup) {
	finishJobs := make(chan int, 5)

	yahoo.Parse(products) // different
	yahoo.c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	yahoo.c.OnScraped(func(r *colly.Response) {
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
			yahoo.c.Visit(url)
			time.Sleep(1 * time.Second)
		}
	}
	for w := 1; w <= 3; w++ {
		go worker(w, yahoo.jobs)
	}

	// Create the jobs
	yahoo.CreateJobs() // different
}

// type Pchome struct {
// 	c       *colly.Collector
// 	jobs    chan string
// 	keyword string
// }

// func (y *Pchome) Parse(products chan Product) {
// 	fmt.Println("pchome parse")
// }

// func (y *Pchome) CreateJobs() {
// 	fmt.Println("create jobs")
// }

func WebClawer(keyword string) {
	products := make(chan Product, 100)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	yahoo := Yahoo{c: colly.NewCollector(), keyword: keyword, jobs: make(chan string, 5)}
	go Clawer(&yahoo, products, wg)

	// Output
	go func() {
		for product := range products {
			fmt.Println(product)
		}
	}()

	wg.Wait()
	close(products)
	fmt.Println("Done")
}

func main() {
	WebClawer("iphone")
}
