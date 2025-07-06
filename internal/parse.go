package internal

import (
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// SplitTextToSentences splits plain text into sentences (simple heuristic).
func SplitTextToSentences(text string) []string {
	re := regexp.MustCompile(`(?m)([^.!?]+[.!?])`)
	matches := re.FindAllString(text, -1)
	for i, s := range matches {
		matches[i] = strings.TrimSpace(s)
	}
	return matches
}

// ExtractCodeSnippets extracts code snippets from HTML nodes.
func ExtractCodeSnippets(n *html.Node) []string {
	var snippets []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "pre" || n.Data == "code") {
			snippets = append(snippets, getNodeText(n))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return snippets
}

// ExtractHeadings extracts headings (h1-h6) from HTML nodes.
func ExtractHeadings(n *html.Node) []string {
	var headings []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.HasPrefix(n.Data, "h") && len(n.Data) == 2 && n.Data[1] >= '1' && n.Data[1] <= '6' {
			headings = append(headings, getNodeText(n))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return headings
}

// getNodeText returns the concatenated text of a node and its children.
func getNodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getNodeText(c))
	}
	return sb.String()
}

// ParseHTMLFromURL fetches and parses HTML from a URL.
func ParseHTMLFromURL(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return html.Parse(resp.Body)
}
