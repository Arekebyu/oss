package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"oss/internal/models"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Client struct {
	es *elasticsearch.Client
}

func NewClient(address string) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{address},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		es: es,
	}, nil
}

func (c *Client) InitIndex(ctx context.Context, schemaJson []byte) error {
	res, err := c.es.Indices.Exists([]string{"pages"}, c.es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		return nil // already exists
	}

	res, err = c.es.Indices.Create(
		"pages",
		c.es.Indices.Create.WithBody(bytes.NewReader(schemaJson)),
		c.es.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}
	return nil
}

func (c *Client) SavePage(ctx context.Context, p models.ScrapedPage) error {
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

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      "pages",
		DocumentID: p.URL,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing: %s", res.String())
	}

	return nil
}
