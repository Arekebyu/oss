package main

import (
	"context"
	"log"
	"os"
	"oss/internal/config"
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
	cfg := config.LoadConfig()
	// elasticsearch shenanigans
	es, _ := search.NewClient(cfg.ElasticsearchURL)
	schema, _ := os.ReadFile("internal/search/schema.json")
	es.InitIndex(context.Background(), schema)

	// postgres db init
	db, err := storage.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	defer db.Close()

	saver := DualSaver{
		PG: db,
		ES: es,
	}

	crawler := crawler.NewCrawler(&saver)

	domains := []string{"crates.io", "docs.rs", "docs.rust.lang.org", "rust-lang.org"}
	startURLs := []string{
		"https://go.dev/doc/tutorial/getting-started",
	}

	crawler.Collector.AllowedDomains = domains

	crawler.Crawl(domains, startURLs)
}
