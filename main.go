package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/deepersensor/documcp/api"
	"github.com/deepersensor/documcp/config"
	"github.com/deepersensor/documcp/docstore"
	"github.com/deepersensor/documcp/index"
	"github.com/deepersensor/documcp/internal"
	"github.com/deepersensor/documcp/scheduler"
)

const version = "0.1.0"

// Global in-memory stores for API and CLI
var (
	globalDocStore = make(map[string]*docstore.Document)
	globalIndex    = index.NewInvertedIndex()
)

func main() {
	// Load config at startup
	configDir, err := config.GetDefaultConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get config dir: %v\n", err)
		os.Exit(1)
	}
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	_ = cfg // suppress unused variable warning

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "crawl":
		crawlCmd := flag.NewFlagSet("crawl", flag.ExitOnError)
		url := crawlCmd.String("url", "", "Seed URL to crawl")
		depth := crawlCmd.Int("depth", 2, "Max crawl depth")
		maxPages := crawlCmd.Int("max", 20, "Max pages to crawl")
		concurrency := crawlCmd.Int("concurrency", 4, "Number of concurrent workers")
		crawlCmd.Parse(os.Args[2:])
		if *url == "" {
			fmt.Println("Please provide a seed URL with -url")
			os.Exit(1)
		}
		s := scheduler.NewScheduler(configDir)
		job, results, err := s.StartCrawlJob(*url, *depth, *maxPages, *concurrency)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Crawl failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Crawling: %s (depth=%d, max=%d, concurrency=%d)\n", *url, *depth, *maxPages, *concurrency)
		fmt.Printf("Process ID: %s\nProcess Dir: %s\n", job.ProcessID, job.ProcessDir)
		// Indexing and docstore population
		for _, res := range results {
			// Parse sentences, code, headings
			sentences := internal.SplitTextToSentences(res.Text)
			// For code/headings, parse HTML again (not optimal, but works for now)
			doc, _ := internal.ParseHTMLFromURL(res.URL)
			var codeSnippets, headings []string
			if doc != nil {
				codeSnippets = internal.ExtractCodeSnippets(doc)
				headings = internal.ExtractHeadings(doc)
			}
			docID := globalIndex.AddDocument(res.URL, "", res.Text)
			d := docstore.NewDocument(docID, res.URL, "", res.Text, headings, codeSnippets, nil, 1)
			globalDocStore[docID] = d
			// Optionally, index sentences and code snippets as well
			for _, s := range sentences {
				globalIndex.AddDocument(res.URL, "", s)
			}
			for _, code := range codeSnippets {
				globalIndex.AddDocument(res.URL, "", code)
			}
		}
		fmt.Printf("Indexed %d documents.\n", len(results))
		fmt.Printf("Results saved to: %s\n", job.ProcessDir)
	case "query":
		queryCmd := flag.NewFlagSet("query", flag.ExitOnError)
		queryStr := queryCmd.String("s", "", "Query string")
		queryCmd.Parse(os.Args[2:])
		if *queryStr == "" {
			fmt.Println("Please provide a query string with -s")
			os.Exit(1)
		}
		results := globalIndex.Search(*queryStr)
		fmt.Printf("Found %d results for query: %q\n", len(results), *queryStr)
		for _, doc := range results {
			d := globalDocStore[doc.ID]
			fmt.Printf("URL: %s\n", d.URL)
			fmt.Printf("Text: %.200s\n", d.Text)
			if len(d.Headings) > 0 {
				fmt.Printf("Headings: %v\n", d.Headings)
			}
			if len(d.CodeSnippets) > 0 {
				fmt.Printf("Code Snippets: %v\n", d.CodeSnippets)
			}
			fmt.Println("-----")
		}
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		port := serveCmd.String("port", "8080", "Port to run API server on")
		serveCmd.Parse(os.Args[2:])
		fmt.Printf("Starting API server on port %s\n", *port)
		// Pass references to globalDocStore and globalIndex to the API
		api.SetGlobalStores(globalDocStore, globalIndex)
		if err := api.StartServer(":" + *port); err != nil {
			fmt.Fprintf(os.Stderr, "API server failed: %v\n", err)
			os.Exit(1)
		}
	case "config":
		configCmd := flag.NewFlagSet("config", flag.ExitOnError)
		dir := configCmd.String("dir", "", "Config directory to use")
		configCmd.Parse(os.Args[2:])
		var configDir string
		var err error
		if *dir != "" {
			configDir = *dir
		} else {
			configDir, err = config.GetDefaultConfigDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get default config dir: %v\n", err)
				os.Exit(1)
			}
		}
		cfg, err := config.LoadConfig(configDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config from %s: %v\n", configDir, err)
			os.Exit(1)
		}
		fmt.Printf("Config loaded from %s:\n", configDir)
		fmt.Printf("  AppName: %s\n", cfg.AppName)
		fmt.Printf("  Version: %s\n", cfg.Version)
	case "version":
		fmt.Println("documcp version", version)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: documcp <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  crawl   -url <seed_url>    Crawl a documentation site")
	fmt.Println("  query   -s <string>        Query indexed content")
	fmt.Println("  serve   [-port <port>]     Start the API server")
	fmt.Println("  config  [-dir <dir>]       Show config from specified directory")
	fmt.Println("  version                     Show version")
}
