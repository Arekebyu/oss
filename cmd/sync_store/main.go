package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"oss/internal/config"
	"oss/internal/models"
	"oss/internal/storage"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

func main() {
	cfg := config.LoadConfig()
	// allow option to reset index (--reset)
	resetIndex := flag.Bool("reset", false, "Delete and recreate elasticsearch index")
	flag.Parse()

	log.Println("Connecting to services")

	pgConnString := cfg.DatabaseURL
	db, err := storage.NewDB(pgConnString)
	if err != nil {
		log.Fatalf("DB Error: %v", err)
	}
	defer db.Close()

	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.ElasticsearchURL},
	}

	es, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		log.Fatalf("ES Error: %v", err)
	}

	if *resetIndex {
		log.Println("Deleting existing index 'pages'...")
		es.Indices.Delete([]string{"pages"})
		schema, _ := os.ReadFile("internal/search/schema.json")
		res, _ := es.Indices.Create("pages", es.Indices.Create.WithBody(bytes.NewReader(schema)))
		res.Body.Close()
	}

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         "pages",
		Client:        es,
		NumWorkers:    4,
		FlushBytes:    5 * 1024 * 1024,
		FlushInterval: 1 * time.Second,
	})
	if err != nil {
		log.Fatalf("Error creating bulk indexer: %v", err)
	}

	var count uint64
	start := time.Now()
	log.Println("Syncing db with es...")

	err = db.IteratePages(context.Background(), func(p models.ScrapedPage) error {
		var codeBuilder strings.Builder
		var textBuilder strings.Builder
		for _, sec := range p.Sections {
			if sec.Type == "code" {
				codeBuilder.WriteString(sec.Content + "\n")
			} else {
				textBuilder.WriteString(sec.Content + "\n")
			}
		}

		doc := map[string]interface{}{
			"url":           p.URL,
			"title":         p.Title,
			"content":       textBuilder.String(),
			"code_snippets": codeBuilder.String(),
			"crawled_at":    p.CrawledAt,
		}

		data, _ := json.Marshal(doc)

		err := bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: p.URL, 
				Body:       bytes.NewReader(data),
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR: %s", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&count, 1)
				},
			},
		)
		return err
	})

	if err != nil {
		log.Fatalf("Iteration failed: %v", err)
	}

	if err := bi.Close(context.Background()); err != nil {
		log.Fatalf("Unexpected error closing bulk indexer: %s", err)
	}

	stats := bi.Stats()
	elapsed := time.Since(start)
	rate := float64(count) / elapsed.Seconds()

	log.Printf("Sync complete")
	log.Printf("Indexed: %d documents", count)
	log.Printf("Time:    %s", elapsed)
	log.Printf("Rate:    %.0f docs/sec", rate)
	log.Printf("Errors:  %d", stats.NumFailed)
}
