package docstore

import "time"

// Document represents a structured crawled document.
type Document struct {
	ID           string            // Unique document ID
	URL          string            // Source URL
	Title        string            // Extracted title (if available)
	Text         string            // Main extracted text
	Headings     []string          // Section headings (h1-h6)
	CodeSnippets []string          // Extracted code snippets
	Metadata     map[string]string // Arbitrary metadata (e.g., last-modified)
	Version      int               // Version number for changed documents
	LastUpdated  time.Time         // Last update timestamp
}

// NewDocument creates a new Document with the given fields.
func NewDocument(id, url, title, text string, headings, codeSnippets []string, metadata map[string]string, version int) *Document {
	return &Document{
		ID:           id,
		URL:          url,
		Title:        title,
		Text:         text,
		Headings:     headings,
		CodeSnippets: codeSnippets,
		Metadata:     metadata,
		Version:      version,
		LastUpdated:  time.Now(),
	}
}
