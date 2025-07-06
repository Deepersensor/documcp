package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/deepersensor/documcp/config"
	"github.com/deepersensor/documcp/scheduler"
)

const version = "0.1.0"

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
		for _, res := range results {
			fmt.Printf("URL: %s\n", res.URL)
			fmt.Printf("Text: %.100s\n", res.Text)
			fmt.Println("-----")
		}
		fmt.Printf("Results saved to: %s\n", job.ProcessDir)
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		port := serveCmd.String("port", "8080", "Port to run API server on")
		serveCmd.Parse(os.Args[2:])
		fmt.Printf("Starting API server on port %s (stub)\n", *port)
		// TODO: Call API server module
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
	fmt.Println("  crawl   -url <seed_url>    Crawl a documentation site (stub)")
	fmt.Println("  serve   [-port <port>]     Start the API server (stub)")
	fmt.Println("  config  [-dir <dir>]       Show config from specified directory")
	fmt.Println("  version                     Show version")
}
