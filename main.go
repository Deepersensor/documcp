package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/deepersensor/documcp/crawler"
)

const version = "0.1.0"

func main() {
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
		c, err := crawler.NewCrawler(*url)
		if err != nil {
			fmt.Println("Invalid URL:", err)
			os.Exit(1)
		}
		fmt.Printf("Crawling: %s (depth=%d, max=%d, concurrency=%d)\n", *url, *depth, *maxPages, *concurrency)
		results := c.Start(*url, *depth, *maxPages, *concurrency)
		for _, res := range results {
			fmt.Printf("URL: %s\n", res.URL)
			fmt.Printf("Text: %.100s\n", res.Text)
			fmt.Println("-----")
		}
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		port := serveCmd.String("port", "8080", "Port to run API server on")
		serveCmd.Parse(os.Args[2:])
		fmt.Printf("Starting API server on port %s (stub)\n", *port)
		// TODO: Call API server module
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
	fmt.Println("  crawl   -url <seed_url>    Crawl a documentation site (stub)")
	fmt.Println("  serve   [-port <port>]     Start the API server (stub)")
	fmt.Println("  version                     Show version")
}
