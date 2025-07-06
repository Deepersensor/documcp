package crawler

import (
	"encoding/json"
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

// Crawler is the main struct for managing crawl state.
type Crawler struct {
	Visited    map[string]struct{}
	Queue      chan string
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
		Queue:      make(chan string, 100),
		Results:    make(chan CrawlResult, 100),
		Host:       u.Host,
		ProcessDir: processDir,
		ProcessID:  filepath.Base(processDir),
	}
	return c, nil
}

// Start begins crawling and persists results to processDir/results.json.
func (c *Crawler) Start(seed string, maxDepth, maxPages, concurrency int) ([]CrawlResult, error) {
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
	close(c.Queue)
	close(c.Results)
	<-done

	// Persist results to disk
	if err := c.saveResults(results); err != nil {
		return results, err
	}
	return results, nil
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
				c.Queue <- l
			}(link, depth+1)
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
