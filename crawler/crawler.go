package crawler

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// CrawlResult holds the crawled URL and its extracted text (stub for now).
type CrawlResult struct {
	URL  string
	Text string
}

// Crawler is the main struct for managing crawl state.
type Crawler struct {
	Visited map[string]struct{}
	Queue   chan string
	Results chan CrawlResult
	Host    string
	mu      sync.Mutex
	wg      sync.WaitGroup
}

// NewCrawler creates a new Crawler for the given seed URL.
func NewCrawler(seed string) (*Crawler, error) {
	u, err := url.Parse(seed)
	if err != nil {
		return nil, err
	}
	c := &Crawler{
		Visited: make(map[string]struct{}),
		Queue:   make(chan string, 100),
		Results: make(chan CrawlResult, 100),
		Host:    u.Host,
	}
	return c, nil
}

// Start begins crawling from the seed URL, up to maxDepth and maxPages.
func (c *Crawler) Start(seed string, maxDepth, maxPages, concurrency int) []CrawlResult {
	c.wg.Add(1)
	go c.enqueue(seed)

	for i := 0; i < concurrency; i++ {
		go c.worker(maxDepth, maxPages)
	}

	var results []CrawlResult
	done := make(chan struct{})
	go func() {
		for res := range c.Results {
			results = append(results, res)
			if len(results) >= maxPages {
				break
			}
		}
		close(done)
	}()

	c.wg.Wait()
	close(c.Queue)   // Only close after all enqueue/crawl goroutines are done
	close(c.Results) // Close results so the collector goroutine can finish
	<-done
	return results
}

func (c *Crawler) enqueue(u string) {
	defer c.wg.Done()
	c.mu.Lock()
	if _, ok := c.Visited[u]; ok {
		c.mu.Unlock()
		return
	}
	c.Visited[u] = struct{}{}
	c.mu.Unlock()
	c.Queue <- u
}

func (c *Crawler) worker(maxDepth, maxPages int) {
	for u := range c.Queue {
		c.crawl(u, 0, maxDepth, maxPages)
	}
}

func (c *Crawler) crawl(u string, depth, maxDepth, maxPages int) {
	if depth > maxDepth || len(c.Visited) > maxPages {
		return
	}
	resp, err := http.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return
	}
	text := extractText(doc)
	c.Results <- CrawlResult{URL: u, Text: text}

	links := extractLinks(doc, c.Host)
	for _, link := range links {
		c.mu.Lock()
		if _, ok := c.Visited[link]; !ok && len(c.Visited) < maxPages {
			c.Visited[link] = struct{}{}
			c.mu.Unlock()
			c.wg.Add(1)
			go func(l string, d int) {
				defer c.wg.Done()
				// Only send to Queue if it's not closed
				c.Queue <- l
			}(link, depth+1)
		} else {
			c.mu.Unlock()
		}
	}
}

// extractLinks finds all internal links in the HTML document.
func extractLinks(n *html.Node, host string) []string {
	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link := attr.Val
					if strings.HasPrefix(link, "/") {
						link = "http://" + host + link
					}
					u, err := url.Parse(link)
					if err == nil && u.Host == host && (u.Scheme == "http" || u.Scheme == "https") {
						links = append(links, u.String())
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return links
}

// extractText is a stub for extracting visible text from HTML.
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(extractText(c))
	}
	return sb.String()
}
