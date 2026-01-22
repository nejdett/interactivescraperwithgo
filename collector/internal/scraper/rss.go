package scraper

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Content     string `xml:"encoded"`
}

type RSSFetcher struct {
	client *http.Client
}

func NewRSSFetcher() *RSSFetcher {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	if useTor := os.Getenv("USE_TOR"); useTor == "true" {
		torProxy := os.Getenv("TOR_PROXY")
		if torProxy == "" {
			torProxy = "tor:9050"
		}

		dialer, err := proxy.SOCKS5("tcp", torProxy, nil, proxy.Direct)
		if err != nil {
		} else {
			transport := &http.Transport{
				Dial: dialer.Dial,
			}
			client.Transport = transport
		}
	}

	return &RSSFetcher{
		client: client,
	}
}

func (f *RSSFetcher) FetchFeed(source Source) ([]ScrapedContent, error) {
	resp, err := f.client.Get(source.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, err
	}

	results := make([]ScrapedContent, 0, len(rss.Channel.Items))
	for _, item := range rss.Channel.Items {
		txt := item.Content
		if txt == "" {
			txt = item.Description
		}

		results = append(results, ScrapedContent{
			Source: Source{
				Name: source.Name,
				URL:  item.Link,
			},
			Content:     stripHTML(txt),
			PublishedAt: parseRSSDate(item.PubDate),
		})
	}

	return results, nil
}

func stripHTML(s string) string {
	// simple html tag removal
	for {
		start := strings.Index(s, "<")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], ">")
		if end == -1 {
			break
		}
		s = s[:start] + " " + s[start+end+1:]
	}
	
	s = strings.Join(strings.Fields(s), " ")
	
	// truncate if too long
	if len(s) > 5000 {
		s = s[:5000]
	}
	return s
}

func parseRSSDate(dateStr string) time.Time {
	// try common RSS date formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
		time.RFC3339,
	}

	for _, fmt := range formats {
		t, err := time.Parse(fmt, dateStr)
		if err == nil {
			return t
		}
	}

	// fallback to current time
	return time.Now()
}
