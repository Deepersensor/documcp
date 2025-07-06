package scheduler

import (
	"path/filepath"

	"github.com/deepersensor/documcp/config"
	"github.com/deepersensor/documcp/crawler"
)

// CrawlJob represents a crawl job request.
type CrawlJob struct {
	SeedURL     string
	MaxDepth    int
	MaxPages    int
	Concurrency int
	ProcessID   string
	ProcessDir  string
}

// Scheduler manages crawl jobs and process directories.
type Scheduler struct {
	ConfigDir string
}

// NewScheduler creates a new Scheduler.
func NewScheduler(configDir string) *Scheduler {
	return &Scheduler{ConfigDir: configDir}
}

// StartCrawlJob creates a process dir, starts a crawl, and returns the job info and results.
func (s *Scheduler) StartCrawlJob(seedURL string, maxDepth, maxPages, concurrency int) (*CrawlJob, []crawler.CrawlResult, error) {
	processesDir := config.GetProcessesDir(s.ConfigDir)
	processDir, err := crawler.NewProcessDir(processesDir)
	if err != nil {
		return nil, nil, err
	}
	c, err := crawler.NewCrawler(seedURL, processDir)
	if err != nil {
		return nil, nil, err
	}
	results, err := c.Start(seedURL, maxDepth, maxPages, concurrency)
	job := &CrawlJob{
		SeedURL:     seedURL,
		MaxDepth:    maxDepth,
		MaxPages:    maxPages,
		Concurrency: concurrency,
		ProcessID:   filepath.Base(processDir),
		ProcessDir:  processDir,
	}
	return job, results, err
}
