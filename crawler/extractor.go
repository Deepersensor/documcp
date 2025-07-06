package crawler

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

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
