package models

type PageSection struct {
	Type     string `json:"type"` // code or text
	Content  string `json:"content"`
	Language string `json:"language,omitempty"` // e.g. go, python, c
}

type ScrapedPage struct {
	URL       string        `json:"url"`
	Title     string        `json:"title"`
	Sections  []PageSection `json:"sections"`
	CrawledAt string        `json:"crawled_at"`
}
