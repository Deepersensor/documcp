package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/deepersensor/documcp/docstore"
	"github.com/deepersensor/documcp/index"
)

var (
	docStore map[string]*docstore.Document
	idx      *index.InvertedIndex
)

func SetGlobalStores(ds map[string]*docstore.Document, i *index.InvertedIndex) {
	docStore = ds
	idx = i
}

// StartServer starts the API server on the given address.
func StartServer(addr string) error {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/mcp", mcpHandler)
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/document/", documentHandler)

	fmt.Printf("API server listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// mcpHandler is a placeholder for Model Context Protocol integration.
func mcpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "MCP endpoint not implemented yet",
	})
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}
	results := idx.Search(q)
	type apiResult struct {
		ID           string   `json:"id"`
		URL          string   `json:"url"`
		Text         string   `json:"text"`
		Headings     []string `json:"headings,omitempty"`
		CodeSnippets []string `json:"code_snippets,omitempty"`
	}
	var out []apiResult
	for _, doc := range results {
		d := docStore[doc.ID]
		out = append(out, apiResult{
			ID:           d.ID,
			URL:          d.URL,
			Text:         d.Text,
			Headings:     d.Headings,
			CodeSnippets: d.CodeSnippets,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func documentHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/document/")
	if id == "" {
		http.Error(w, "Missing document ID", http.StatusBadRequest)
		return
	}
	d, ok := docStore[id]
	if !ok {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}
