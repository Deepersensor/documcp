package crawler

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// NewProcessDir creates a unique process directory under the given baseDir.
func NewProcessDir(baseDir string) (string, error) {
	processID := fmt.Sprintf("%d", time.Now().UnixNano())
	processDir := filepath.Join(baseDir, processID)
	if err := os.MkdirAll(processDir, 0o755); err != nil {
		return "", err
	}
	return processDir, nil
}
