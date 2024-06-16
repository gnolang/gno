package doctest

import (
	"fmt"
	"os"
)

// ReadMarkdownFile reads a markdown file and returns its content
func ReadMarkdownFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}
