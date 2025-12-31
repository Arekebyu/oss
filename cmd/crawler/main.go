package main

import (
	"context"
	"log"
	"os"
	"oss/internal/crawler"
	"oss/internal/models"
	"oss/internal/search"
	"oss/internal/storage"
)

type DualSaver struct {
	PG *storage.DB
	ES *search.Client
}

func (ds *DualSaver) SavePage(ctx context.Context, p models.ScrapedPage) error {
	if err := ds.PG.SavePage(ctx, p); err != nil {
		return err
	}
	if err := ds.ES.SavePage(ctx, p); err != nil {
		log.Printf("Warning: Failed to index page %s: %v", p.URL, err)
	}
	return nil
}

func main() {
	// testing enviroment should probably change to os.env

	// elasticsearch shenanigans
	es, _ := search.NewClient("http://localhost:9200")
	schema, _ := os.ReadFile("internal/search/schema.json")
	es.InitIndex(context.Background(), schema)

	// postgres db init
	connString := "postgres://admin:secretpassword@localhost:5432/search_engine"
	db, err := storage.NewDB(connString)
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	defer db.Close()

	saver := DualSaver{
		PG: db,
		ES: es,
	}

	crawler := crawler.NewCrawler(&saver)

	domains := []string{"pytorch.org", "go.dev"}
	startURLs := []string{
		"https://pytorch.org/docs/stable/tensors.html",
		"https://go.dev/doc/tutorial/getting-started",
	}

	crawler.Collector.AllowedDomains = domains

	crawler.Crawl(domains, startURLs)
}
