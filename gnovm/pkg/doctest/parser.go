package doctest

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type CodeBlock struct {
	Content string
	Start   int
	End     int
	T       string
	Index   int
}

// ReadMarkdownFile reads a markdown file and returns its content
func ReadMarkdownFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// getCodeBlocks extracts code blocks from the markdown file content
func getCodeBlocks(body string) []CodeBlock {
	blocksRegex := regexp.MustCompile("```\\w*[^`]+```*")
	matches := blocksRegex.FindAllStringIndex(body, -1)

	return mapWithIndex(extractCodeBlock, matches, body)
}

// extractCodeBlock extracts a single code block from the markdown content
func extractCodeBlock(match []int, index int, body string) CodeBlock {
	if len(match) < 2 {
		return CodeBlock{}
	}

	codeStr := body[match[0]:match[1]]
	// Remove the backticks from the code block content
	codeStr = strings.TrimPrefix(codeStr, "```")
	codeStr = strings.TrimSuffix(codeStr, "```")

	result := CodeBlock{
		Content: codeStr,
		Start:   match[0],
		End:     match[1],
		Index:   index,
	}

	// extract the type (language) of the code block
	lines := strings.Split(codeStr, "\n")
	if len(lines) > 0 {
		line1 := lines[0]
		languageRegex := regexp.MustCompile(`^\w*`)
		languageMatch := languageRegex.FindString(line1)
		result.T = languageMatch
		// Remove the language specifier from the code block content
		result.Content = strings.TrimPrefix(result.Content, languageMatch)
		result.Content = strings.TrimSpace(result.Content)
	}
	if result.T == "" {
		result.T = "plain"
	}

	return result
}

// mapWithIndex applies a function to each element of a slice along with its index
func mapWithIndex[T, R any](f func(T, int, string) R, xs []T, body string) []R {
	result := make([]R, len(xs))
	for i, x := range xs {
		result[i] = f(x, i, body)
	}
	return result
}

func WriteCodeBlockToFile(c CodeBlock) error {
	if c.T == "go" {
		c.T = "gno"
	}

	fileName := fmt.Sprintf("%d.%s", c.Index, c.T)
	file, err := os.Create(fileName) // TODO: use temp file
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(c.Content)
	if err != nil {
		return err
	}

	return nil
}
