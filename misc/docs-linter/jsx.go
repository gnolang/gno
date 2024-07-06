package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	reCodeBlocks = regexp.MustCompile("(?s)```.*?```")
	reInlineCode = regexp.MustCompile("`[^`]*`")
)

// extractJSX extracts JSX tags from given file content
func extractJSX(fileContent []byte) []string {
	text := string(fileContent)

	// Remove code blocks
	contentNoCodeBlocks := reCodeBlocks.ReplaceAllString(text, "")

	// Remove inline code
	contentNoInlineCode := reInlineCode.ReplaceAllString(contentNoCodeBlocks, "")

	// Extract JSX/HTML elements
	reJSX := regexp.MustCompile("(?s)<[^>]+>")

	matches := reJSX.FindAllString(contentNoInlineCode, -1)

	filteredMatches := make([]string, 0)
	// Ignore HTML comments and escaped JSX
	for _, m := range matches {
		if !strings.Contains(m, "!--") && !strings.Contains(m, "\\>") {
			filteredMatches = append(filteredMatches, m)
		}
	}

	return filteredMatches
}

func lintJSX(filepathToJSX map[string][]string) (string, error) {
	var (
		found  bool
		output bytes.Buffer
	)
	for filePath, tags := range filepathToJSX {
		for _, tag := range tags {
			if !found {
				output.WriteString("Tags that need checking:\n")
				found = true
			}

			output.WriteString(fmt.Sprintf(">>> %s (found in file: %s)\n", tag, filePath))
		}
	}

	if found {
		return output.String(), errFoundUnescapedJSXTags
	}

	return "", nil
}
