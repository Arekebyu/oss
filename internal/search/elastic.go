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

func (c *Client) Search(ctx context.Context, query string) ([]models.ScrapedPage, error) {
	searchQuery := map[string]interface{}{
		"size": 50,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     query,
				"fields":    []string{"title^3", "code_snippets^2", "content"},
				"fuzziness": "AUTO",
			},
		},
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(searchQuery)
	if err != nil {
		return nil, err
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex("pages"),
		c.es.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error %s", res)
	}

	var r map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	var results []models.ScrapedPage
	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		page := models.ScrapedPage{
			URL:   source["url"].(string),
			Title: source["title"].(string),
			Sections: []models.PageSection{
				{Content: fmt.Sprintf("%v", source["content"])[:200] + "..."},
			},
		}
		results = append(results, page)
	}
	return results, nil
}