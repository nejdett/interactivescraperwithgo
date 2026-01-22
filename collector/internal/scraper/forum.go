package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/proxy"
)

type ForumScraper struct {
	client *http.Client
}

func NewForumScraper() *ForumScraper {
	client := &http.Client{
		Timeout: 90 * time.Second,
	}

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

	return &ForumScraper{client: client}
}

func (f *ForumScraper) ScrapeForumDeep(source Source) ([]ScrapedContent, error) {
	results := []ScrapedContent{}

	resp, err := f.client.Get(source.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	links := f.extractLinks(doc, source.URL)

	// separate thread links from other links
	var threads, others []string
	for _, link := range links {
		linkLower := strings.ToLower(link)
		isThread := strings.Contains(linkLower, "thread") || 
		            strings.Contains(linkLower, "topic") || 
		            strings.Contains(linkLower, "post") || 
		            strings.Contains(linkLower, "discussion")
		
		if isThread {
			threads = append(threads, link)
		} else {
			others = append(others, link)
		}
	}
	
	links = append(threads, others...)
	
	// limit to 100 links max
	if len(links) > 100 {
		links = links[:100]
	}

	for idx, link := range links {
		if idx > 0 {
			time.Sleep(2 * time.Second)
		}

		content, err := f.scrapePage(link, source.Name)
		if err == nil && content.Content != "" {
			results = append(results, content)
		}
	}

	return results, nil
}

func (f *ForumScraper) extractLinks(n *html.Node, baseURL string) []string {
	links := []string{}
	seen := make(map[string]bool)

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if link := f.normalizeURL(a.Val, baseURL); link != "" && !seen[link] && f.isValidForumLink(link, baseURL) {
						seen[link] = true
						links = append(links, link)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return links
}

func (f *ForumScraper) normalizeURL(href, baseURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	u, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(u)
	return resolved.String()
}

func (f *ForumScraper) isValidForumLink(link, baseURL string) bool {
	linkURL, err := url.Parse(link)
	if err != nil {
		return false
	}
	
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	
	// must be same host
	if linkURL.Host != baseURLParsed.Host {
		return false
	}

	linkLower := strings.ToLower(link)

	// exclude common non-content pages
	excludePatterns := []string{
		"login", "register", "logout", "search", "profile",
		"memberlist", "ucp", "faq", "terms", "privacy",
		"user-", "member.php", "user.php", "member/",
		".css", ".js", ".png", ".jpg", ".gif", ".ico",
	}
	
	for _, pattern := range excludePatterns {
		if strings.Contains(linkLower, pattern) {
			return false
		}
	}

	// check for thread patterns
	threadPatterns := []string{"thread", "topic", "post", "discussion", "tid=", "topic="}
	for _, pattern := range threadPatterns {
		if strings.Contains(linkLower, pattern) {
			return true
		}
	}

	// check for forum patterns
	forumPatterns := []string{"forum", "board", "category", "fid="}
	for _, pattern := range forumPatterns {
		if strings.Contains(linkLower, pattern) {
			return true
		}
	}

	return false
}

func (f *ForumScraper) scrapePage(pageURL, sourceName string) (ScrapedContent, error) {
	resp, err := f.client.Get(pageURL)
	if err != nil {
		return ScrapedContent{}, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return ScrapedContent{}, err
	}

	return ScrapedContent{
		Source:      Source{Name: sourceName, URL: pageURL},
		Content:     ExtractText(doc),
		PublishedAt: time.Now(),
	}, nil
}

