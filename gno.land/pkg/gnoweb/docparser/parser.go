package docparser

import (
	"bufio"
	"errors"
	"html/template"
	"io"
	"strings"
)

// ParserConfig holds configuration for the documentation parser
type ParserConfig struct {
	TabWidth   int // Default: 4
	MaxDocSize int // Default: 100000
}

// DefaultConfig returns the default parser configuration
func DefaultConfig() ParserConfig {
	return ParserConfig{
		TabWidth:   4,
		MaxDocSize: 100000,
	}
}

// DocBlock represents a block in documentation
type DocBlock struct {
	Type    string // "text" or "code"
	Content string
}

var (
	ErrEmptyDocumentation = errors.New("empty documentation")
	ErrDocumentationTooLarge = errors.New("documentation too large")
)

// ParseDocumentation parses documentation into blocks
func ParseDocumentation(doc string) ([]DocBlock, error) {
	return ParseDocumentationWithConfig(doc, DefaultConfig())
}

// ParseDocumentationWithConfig parses documentation with custom configuration
func ParseDocumentationWithConfig(doc string, config ParserConfig) ([]DocBlock, error) {
	if doc == "" {
		return nil, ErrEmptyDocumentation
	}
	
	if len(doc) > config.MaxDocSize {
		return nil, ErrDocumentationTooLarge
	}
	
	return parseDocumentationReader(strings.NewReader(doc), config)
}

// ParseDocumentationReader parses documentation from an io.Reader
func ParseDocumentationReader(r io.Reader) ([]DocBlock, error) {
	return ParseDocumentationReaderWithConfig(r, DefaultConfig())
}

// ParseDocumentationReaderWithConfig parses documentation from an io.Reader with custom configuration
func ParseDocumentationReaderWithConfig(r io.Reader, config ParserConfig) ([]DocBlock, error) {
	return parseDocumentationReader(r, config)
}

// parseDocumentationReader is the core parsing function using streaming
func parseDocumentationReader(r io.Reader, config ParserConfig) ([]DocBlock, error) {
	scanner := bufio.NewScanner(r)
	var blocks []DocBlock
	var currentText strings.Builder
	var currentCode strings.Builder
	lineNumber := 0
	
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		
		// Skip empty lines at the beginning
		if lineNumber == 1 && strings.TrimSpace(line) == "" {
			continue
		}
		
		// Check if this line starts a code block
		if isIndented(line, config.TabWidth) {
			// Flush any pending text block
			if currentText.Len() > 0 {
				content := strings.TrimSpace(currentText.String())
				if content != "" {
					blocks = append(blocks, DocBlock{
						Type:    "text",
						Content: escapeHTML(content),
					})
				}
				currentText.Reset()
			}
			
			// Start code block
			currentCode.WriteString(line)
			currentCode.WriteString("\n")
			
			// Continue reading code block
			for scanner.Scan() {
				nextLine := scanner.Text()
				if isIndented(nextLine, config.TabWidth) || nextLine == "" {
					currentCode.WriteString(nextLine)
					currentCode.WriteString("\n")
				} else {
					// End of code block
					break
				}
			}
			
			// Process code block
			codeContent := normalizeIndentation(strings.TrimSuffix(currentCode.String(), "\n"), config.TabWidth)
			if codeContent != "" {
				blocks = append(blocks, DocBlock{
					Type:    "code",
					Content: codeContent,
				})
			}
			currentCode.Reset()
			
			// Add the non-indented line that ended the code block
			if scanner.Text() != "" {
				currentText.WriteString(scanner.Text())
				currentText.WriteString("\n")
			}
			continue
		}
		
		// Regular text
		if currentText.Len() > 0 {
			currentText.WriteString("\n")
		}
		currentText.WriteString(line)
	}
	
	// Flush remaining text
	if currentText.Len() > 0 {
		content := strings.TrimSpace(currentText.String())
		if content != "" {
			blocks = append(blocks, DocBlock{
				Type:    "text",
				Content: escapeHTML(content),
			})
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return blocks, nil
}

// isIndented checks if a line is indented (starts with tab or spaces)
func isIndented(line string, tabWidth int) bool {
	if line == "" {
		return false
	}
	
	// Check for tab
	if strings.HasPrefix(line, "\t") {
		return true
	}
	
	// Check for spaces (only if line is long enough)
	if len(line) >= tabWidth && strings.HasPrefix(line, strings.Repeat(" ", tabWidth)) {
		return true
	}
	
	return false
}

// normalizeIndentation removes common indentation from all lines in a code block
func normalizeIndentation(codeLines string, tabWidth int) string {
	if codeLines == "" {
		return ""
	}
	
	lines := strings.Split(codeLines, "\n")
	
	// Convert tabs to spaces first for consistent processing
	var normalizedLines []string
	for _, line := range lines {
		normalizedLines = append(normalizedLines, strings.ReplaceAll(line, "\t", strings.Repeat(" ", tabWidth)))
	}
	
	// Find minimum indentation
	minIndent := -1
	for _, line := range normalizedLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		indent := countIndentation(line, tabWidth)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}
	
	// If no indentation found, return as-is
	if minIndent <= 0 {
		return strings.Join(lines, "\n")
	}
	
	// Remove common indentation
	var resultLines []string
	for _, line := range normalizedLines {
		if strings.TrimSpace(line) == "" {
			resultLines = append(resultLines, "")
		} else if len(line) >= minIndent {
			resultLines = append(resultLines, line[minIndent:])
		} else {
			// Line is shorter than minIndent, keep as-is
			resultLines = append(resultLines, line)
		}
	}
	
	// Join and trim trailing newlines
	result := strings.Join(resultLines, "\n")
	return strings.TrimSuffix(result, "\n")
}

// countIndentation counts the number of indentation characters at the start of a line
func countIndentation(line string, tabWidth int) int {
	if line == "" {
		return 0
	}
	
	count := 0
	for _, r := range line {
		switch r {
		case '\t':
			count += tabWidth // Treat tab as configurable width
		case ' ':
			count++
		default:
			return count
		}
	}
	
	return count
}

// escapeHTML escapes HTML special characters in text content
func escapeHTML(text string) string {
	return template.HTMLEscapeString(text)
} 