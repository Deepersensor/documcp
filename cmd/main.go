package main

import (
	"flag"
	"fmt"
	"os"
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
		crawlCmd.Parse(os.Args[2:])
		if *url == "" {
			fmt.Println("Please provide a seed URL with -url")
			os.Exit(1)
		}
		fmt.Printf("Crawling: %s (stub)\n", *url)
		// TODO: Call crawler module
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

// Deprecated: CLI entrypoint moved to project root main.go.
// Please use `go run .` or `go build` from the root directory.
