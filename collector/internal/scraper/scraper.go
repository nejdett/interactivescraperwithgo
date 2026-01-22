package scraper

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/proxy"
)

type Source struct {
	Name string
	URL  string
}

type ScrapedContent struct {
	Source      Source
	Content     string
	PublishedAt time.Time
}

type WebScraper struct {
	client  *http.Client
	sources []Source
}

func NewWebScraper(sources []Source) *WebScraper {
	timeout := 60 * time.Second
	client := &http.Client{Timeout: timeout}

	useTor := os.Getenv("USE_TOR")
	if useTor == "true" {
		torProxy := os.Getenv("TOR_PROXY")
		if torProxy == "" {
			torProxy = "tor:9050"
		}

		dialer, err := proxy.SOCKS5("tcp", torProxy, nil, proxy.Direct)
		if err == nil {
			client.Transport = &http.Transport{Dial: dialer.Dial}
		}
	}

	return &WebScraper{
		client:  client,
		sources: sources,
	}
}

func (s *WebScraper) ScrapeAll() ([]ScrapedContent, error) {
	var results []ScrapedContent

	for _, source := range s.sources {
		content, err := s.ScrapeSource(source)
		if err != nil {
			continue
		}
		results = append(results, content)
	}

	return results, nil
}

func (s *WebScraper) ScrapeSource(source Source) (ScrapedContent, error) {
	resp, err := s.client.Get(source.URL)
	if err != nil {
		return ScrapedContent{}, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return ScrapedContent{}, err
	}

	return ScrapedContent{
		Source:      source,
		Content:     ExtractText(doc),
		PublishedAt: time.Now(),
	}, nil
}

func ExtractText(n *html.Node) string {
	var buf strings.Builder
	skip := map[string]bool{
		"script": true, "style": true, "noscript": true,
		"iframe": true, "svg": true,
	}

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && skip[node.Data] {
			return
		}
		if node.Type == html.TextNode {
			txt := strings.TrimSpace(node.Data)
			if txt != "" {
				buf.WriteString(txt)
				buf.WriteByte(' ')
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)

	result := strings.Join(strings.Fields(buf.String()), " ")
	
	// limit content size
	maxLen := 5000
	if len(result) > maxLen {
		result = result[:maxLen]
	}
	return result
}
