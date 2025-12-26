package main

import (
	"log"
	"oss/internal/crawler"
	"oss/internal/storage"
)

func main() {
	// testing enviroment but will probably change to os.env
	connString := "postgres://admin:secretpassword@localhost:5432/search_engine"
	db, err := storage.NewDB(connString)
	if err != nil {
		log.Fatal("Error connecting to database: %v", err)
	}
	defer db.Close()

	crawler := crawler.NewCrawler(db)

	domains := []string{"pytorch.org", "go.dev"}
	startURLs := []string{
		"https://pytorch.org/docs/stable/tensors.html",
		"https://go.dev/doc/tutorial/getting-started",
	}

	crawler.Collector.AllowedDomains = domains

	crawler.Crawl(domains, startURLs)
}
