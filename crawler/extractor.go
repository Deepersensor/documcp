package crawler

import (
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// extractLinks finds all internal links in the HTML document.
func extractLinks(n *html.Node, baseHost string) []string {
	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link := attr.Val
					// Ignore empty, fragment, or mailto links
					if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "mailto:") {
						continue
					}
					// Resolve relative links
					if strings.HasPrefix(link, "/") {
						link = "https://" + baseHost + link
					}
					u, err := url.Parse(link)
					if err != nil {
						continue
					}
					// Normalize host (strip www.)
					host := strings.TrimPrefix(u.Host, "www.")
					base := strings.TrimPrefix(baseHost, "www.")
					if host == "" {
						u.Host = base
						u.Scheme = "https"
						link = u.String()
						host = base
					}
					// Only follow links within the same host
					if host == base && (u.Scheme == "http" || u.Scheme == "https") {
						// Remove fragment
						u.Fragment = ""
						// Normalize path (remove duplicate slashes)
						u.Path = path.Clean(u.Path)
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
