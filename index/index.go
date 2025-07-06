package index

import (
	"strconv"
	"strings"
	"sync"
)

// Document represents a crawled document to be indexed.
type Document struct {
	ID    string
	URL   string
	Title string
	Text  string
}

// InvertedIndex is a simple in-memory full-text index.
type InvertedIndex struct {
	mu        sync.RWMutex
	Docs      map[string]Document
	Index     map[string]map[string]struct{} // term -> set of doc IDs
	nextDocID int
}

// NewInvertedIndex creates a new empty index.
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		Docs:  make(map[string]Document),
		Index: make(map[string]map[string]struct{}),
	}
}

// AddDocument indexes a new document.
func (idx *InvertedIndex) AddDocument(url, title, text string) string {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	docID := idx.generateDocID()
	doc := Document{
		ID:    docID,
		URL:   url,
		Title: title,
		Text:  text,
	}
	idx.Docs[docID] = doc
	for _, term := range tokenize(text) {
		if idx.Index[term] == nil {
			idx.Index[term] = make(map[string]struct{})
		}
		idx.Index[term][docID] = struct{}{}
	}
	return docID
}

// Search returns document IDs matching all terms.
func (idx *InvertedIndex) Search(query string) []Document {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	terms := tokenize(query)
	if len(terms) == 0 {
		return nil
	}
	var resultIDs map[string]struct{}
	for i, term := range terms {
		docSet, ok := idx.Index[term]
		if !ok {
			return nil
		}
		if i == 0 {
			resultIDs = make(map[string]struct{}, len(docSet))
			for id := range docSet {
				resultIDs[id] = struct{}{}
			}
		} else {
			for id := range resultIDs {
				if _, ok := docSet[id]; !ok {
					delete(resultIDs, id)
				}
			}
		}
	}
	var results []Document
	for id := range resultIDs {
		results = append(results, idx.Docs[id])
	}
	return results
}

// GetDocument returns a document by its ID.
func (idx *InvertedIndex) GetDocument(id string) (Document, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	doc, ok := idx.Docs[id]
	return doc, ok
}

// SearchSentences returns all sentences from indexed documents that match the query.
func (idx *InvertedIndex) SearchSentences(query string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	terms := tokenize(query)
	if len(terms) == 0 {
		return nil
	}
	var resultIDs map[string]struct{}
	for i, term := range terms {
		docSet, ok := idx.Index[term]
		if !ok {
			return nil
		}
		if i == 0 {
			resultIDs = make(map[string]struct{}, len(docSet))
			for id := range docSet {
				resultIDs[id] = struct{}{}
			}
		} else {
			for id := range resultIDs {
				if _, ok := docSet[id]; !ok {
					delete(resultIDs, id)
				}
			}
		}
	}
	var sentences []string
	for id := range resultIDs {
		doc := idx.Docs[id]
		// Split text into sentences (simple heuristic)
		for _, s := range tokenizeSentences(doc.Text) {
			if containsAllTerms(s, terms) {
				sentences = append(sentences, s)
			}
		}
	}
	return sentences
}

// generateDocID returns a new unique document ID.
func (idx *InvertedIndex) generateDocID() string {
	idx.nextDocID++
	return "doc" + strconv.Itoa(idx.nextDocID)
}

// tokenize splits text into lowercase words.
func tokenize(text string) []string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9')
	})
	for i, w := range words {
		words[i] = strings.ToLower(w)
	}
	return words
}

// Helper to split text into sentences (simple heuristic).
func tokenizeSentences(text string) []string {
	// Simple split by period, exclamation, or question mark.
	// You may want to use a more robust method.
	sentences := []string{}
	start := 0
	for i, r := range text {
		if r == '.' || r == '!' || r == '?' {
			s := text[start : i+1]
			s = strings.TrimSpace(s)
			if s != "" {
				sentences = append(sentences, s)
			}
			start = i + 1
		}
	}
	// Add any trailing text
	if start < len(text) {
		s := strings.TrimSpace(text[start:])
		if s != "" {
			sentences = append(sentences, s)
		}
	}
	return sentences
}

// Helper to check if all terms are in the sentence.
func containsAllTerms(s string, terms []string) bool {
	s = strings.ToLower(s)
	for _, t := range terms {
		if !strings.Contains(s, t) {
			return false
		}
	}
	return true
}
