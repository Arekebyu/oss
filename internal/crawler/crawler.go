package crawler

import (
	"context"
	"fmt"
	"log"
	"oss/internal/models"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

type Saver interface {
	SavePage(ctx context.Context, doc models.ScrapedPage) error
}

type Crawler struct {
	Collector *colly.Collector
	saver     Saver
}

func NewCrawler(saver Saver) *Crawler {

	col := colly.NewCollector(
		colly.Async(true),
	)

	// todo! set up useragent
	// col.UserAgent = "OpenSourceSearchBot/1.0 (+http://your-website.com/bot)"

	extensions.RandomUserAgent(col)

	return &Crawler{
		Collector: col,
		saver:     saver,
	}
}

func (crawler *Crawler) Crawl(domains []string, startURLs []string) {

	crawler.Collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 4,
		Delay:       1 * time.Second,
	})

	crawler.Collector.OnError(func(r *colly.Response, err error) {
		log.Printf("error visiting %s: %v \n", r.Request.URL, err)
	})

	// heuristic searching for article, main, or generic divs
	crawler.Collector.OnHTML("article, main, div[role='main'], .documentation", func(e *colly.HTMLElement) {
		page := models.ScrapedPage{

			URL:       e.Request.URL.String(),
			Title:     e.ChildText("h1"),
			CrawledAt: time.Now().Format(time.RFC3339),
		}

		if page.Title == "" {
			page.Title = e.DOM.Find("title").Text()
		}

		e.ForEach("p, pre, h2, h3", func(_ int, el *colly.HTMLElement) {
			tagName := el.Name
			text := strings.TrimSpace(el.Text)

			if text == "" {
				return
			}

			switch tagName {
			case "pre":
				// likely a code block
				page.Sections = append(page.Sections, models.PageSection{
					Type:     "code",
					Content:  text,
					Language: "detected",
				})
			case "h2", "h3":
				page.Sections = append(page.Sections, models.PageSection{
					Type:    "text",
					Content: "## " + text, // header in markdown
				})
			default:
				page.Sections = append(page.Sections, models.PageSection{
					Type:    "text",
					Content: text,
				})
			}
		})

		if len(page.Sections) > 0 {
			crawler.savePage(page)
		}
	})

	crawler.Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// todo! remove login or signups
		if isDocsLink(link) {
			e.Request.Visit(e.Request.AbsoluteURL(link))
		}
	})

	for _, url := range startURLs {
		crawler.Collector.Visit(url)
	}

	crawler.Collector.Wait()
}

func (crawler *Crawler) savePage(p models.ScrapedPage) {
	ctx := context.Background()
	err := crawler.saver.SavePage(ctx, p)
	if err != nil {
		log.Printf("failed to save page to DB :%v\n", err)
		return
	}
	fmt.Printf("sections saved to db: %s, (%d, sections)\n", p.Title, len(p.Sections))
}

// heuristic to remove signin pages
func isDocsLink(link string) bool {
	if strings.HasPrefix(link, "#") || strings.Contains(link, "signin") {
		return false
	}
	return true
}
