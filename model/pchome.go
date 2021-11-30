package model

import (
	"encoding/json"
	"fmt"

	"github.com/gocolly/colly"
)

type Pchome struct {
	Keyword string
}

type pchomeData struct {
	TotalRows int    `json:"totalRows"`
	TotalPage int    `json:"totalPage"`
	Prods     []Item `json:"prods"`
}

type Item struct {
	Id    string `json:"Id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	PicS  string `json:"picS"`
}

// Parse the response depends on different websites. (json)
func (w *Pchome) Parse(collyWorker *colly.Collector, products chan Product) {
	collyWorker.OnResponse(func(r *colly.Response) {
		fmt.Println("On response")
		var data pchomeData
		if err := json.Unmarshal(r.Body, &data); err != nil {
			fmt.Println(err)
		}

		for _, item := range data.Prods {
			productURL := fmt.Sprintf("https://24h.pchome.com.tw/prod/%s", item.Id)
			imageURL := fmt.Sprintf("https://e.ecimg.tw%s", item.PicS)
			products <- Product{Name: item.Name, Price: item.Price, ImageURL: imageURL, ProductURL: productURL}
		}
	})
}

// Create the jobs depends on the URL of different websites.
// Each page contains 20 items.
func (w *Pchome) CreateJobs(jobs chan string) {
	for i := 1; i <= 5; i++ {
		jobs <- fmt.Sprintf("https://ecshweb.pchome.com.tw/search/v3.3/all/results?page=%v&q=%s&sort=rnk", i, w.Keyword)
	}
}
