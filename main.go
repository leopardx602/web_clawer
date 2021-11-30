package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/leopardx602/web_crawler/model"
)

type Web interface {
	Parse(collyWorker *colly.Collector, products chan model.Product)
	CreateJobs(jobs chan string)
}

// Build a crawler for a web site.
// Create workers to get specified pages.
func Crawler(ctx context.Context, web Web, products chan model.Product, wg *sync.WaitGroup) {
	finishJobs := make(chan int, 5)

	// Set crawler
	collyWorker := colly.NewCollector()

	web.Parse(collyWorker, products)
	collyWorker.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
		finishJobs <- 1
	})

	collyWorker.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		finishJobs <- 1
	})

	// Assign the jobs to the workers
	worker := func(id int, jobs <-chan string) {
		for {
			select {
			case url := <-jobs:
				fmt.Println("go to: ", url)
				collyWorker.Visit(url)
				time.Sleep(1 * time.Second)
			case <-ctx.Done():
				return
			}
		}
	}

	// Create the workers
	var jobs = make(chan string, 5)
	for w := 1; w <= 3; w++ {
		go worker(w, jobs)
	}

	// Create the jobs
	web.CreateJobs(jobs)

	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			if len(finishJobs) == 5 {
				time.Sleep(time.Second)
				wg.Done()
				return
			}
			time.Sleep(time.Second)
		}
	}
}

// Looking for on different web sites. Continuous data output when found.
func WebCrawler(ctx context.Context, products chan model.Product, keyword string) {

	// Set the websites to search.
	var webs []Web
	webs = append(webs, &model.Yahoo{Keyword: keyword})
	webs = append(webs, &model.Pchome{Keyword: keyword})
	wg := new(sync.WaitGroup)
	wg.Add(len(webs))
	for _, web := range webs {
		go Crawler(ctx, web, products, wg)
	}

	wg.Wait()
	close(products)
}

func main() {
	// Store all collected products.
	products := make(chan model.Product, 100)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go WebCrawler(ctx, products, "iphone")
	for product := range products {
		fmt.Println(product)
	}
	fmt.Println("Done")
}
