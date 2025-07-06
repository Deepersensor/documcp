package crawler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/net/html"
)

// CrawlResult holds the crawled URL and its extracted text (stub for now).
type CrawlResult struct {
	URL  string
	Text string
}

// queueItem holds a URL and its crawl depth.
type queueItem struct {
	url   string
	depth int
}

// Crawler is the main struct for managing crawl state.
type Crawler struct {
	Visited    map[string]struct{}
	Queue      chan queueItem
	Results    chan CrawlResult
	Host       string
	mu         sync.Mutex
	wg         sync.WaitGroup
	ProcessDir string // Directory for this crawl process
	ProcessID  string // Unique process ID
}

// NewCrawler creates a new Crawler for the given seed URL and process directory.
func NewCrawler(seed, processDir string) (*Crawler, error) {
	u, err := url.Parse(seed)
	if err != nil {
		return nil, err
	}
	c := &Crawler{
		Visited:    make(map[string]struct{}),
		Queue:      make(chan queueItem, 100),
		Results:    make(chan CrawlResult, 100),
		Host:       u.Host,
		ProcessDir: processDir,
		ProcessID:  filepath.Base(processDir),
	}
	return c, nil
}

// Start begins crawling and persists results to processDir/results.json.
func (c *Crawler) Start(seed string, maxDepth, maxPages, concurrency int) ([]CrawlResult, error) {
	// Start the crawl with the seed URL
	c.wg.Add(1)
	go c.enqueue(seed, 0)

	// Start worker goroutines
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

	// Wait for all URLs to be processed
	c.wg.Wait()
	close(c.Queue)
	// Wait for results to be collected or for maxPages to be reached
	<-done
	// After results are collected, close Results channel to terminate any consumers
	close(c.Results)

	// Persist results to disk
	if err := c.saveResults(results); err != nil {
		return results, err
	}
	return results, nil
}

func (c *Crawler) worker(maxDepth, maxPages int) {
	for item := range c.Queue {
		c.crawl(item.url, item.depth, maxDepth, maxPages)
	}
}

func (c *Crawler) enqueue(u string, depth int) {
	defer c.wg.Done()
	c.mu.Lock()
	// If already visited or over maxPages, skip
	if _, ok := c.Visited[u]; ok {
		c.mu.Unlock()
		return
	}
	c.Visited[u] = struct{}{}
	c.mu.Unlock()
	// Add to queue for processing
	c.Queue <- queueItem{url: u, depth: depth}
}

func (c *Crawler) crawl(u string, depth, maxDepth, maxPages int) {
	if depth > maxDepth {
		return
	}
	c.mu.Lock()
	if len(c.Visited) > maxPages {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	fmt.Printf("[CRAWL] Depth %d: %s\n", depth, u)
	resp, err := http.Get(u)
	if err != nil {
		fmt.Printf("[ERROR] Failed to GET %s: %v\n", u, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[ERROR] Non-OK status for %s: %d\n", u, resp.StatusCode)
		return
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Printf("[ERROR] Failed to parse HTML for %s: %v\n", u, err)
		return
	}
	text := extractText(doc)
	c.Results <- CrawlResult{URL: u, Text: text}

	links := extractLinks(doc, c.Host)
	fmt.Printf("[LINKS] Found %d links on %s\n", len(links), u)
	for _, link := range links {
		c.mu.Lock()
		if _, ok := c.Visited[link]; !ok && len(c.Visited) < maxPages {
			c.Visited[link] = struct{}{}
			c.mu.Unlock()
			fmt.Printf("[ENQUEUE] %s (depth %d)\n", link, depth+1)
			c.wg.Add(1)
			go c.enqueue(link, depth+1)
		} else {
			c.mu.Unlock()
		}
	}
}

// saveResults persists crawl results to processDir/results.json.
func (c *Crawler) saveResults(results []CrawlResult) error {
	if err := os.MkdirAll(c.ProcessDir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(c.ProcessDir, "results.json"))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
