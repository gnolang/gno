package main

import (
	"fmt"
	"regexp"
	"strings"
)

func extractJSX(fileContent []byte) []string {
	text := string(fileContent)

	// Remove code blocks
	reCodeBlocks := regexp.MustCompile("(?s)```.*?```")
	contentNoCodeBlocks := reCodeBlocks.ReplaceAllString(text, "")

	// Remove inline code
	reInlineCode := regexp.MustCompile("`[^`]*`")
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

func lintJSX(fileJSXMap map[string][]string) error {
	found := false
	for filePath, tags := range fileJSXMap {
		filePath := filePath
		for _, tag := range tags {
			if !found {
				fmt.Println("Tags that need checking:")
				found = true
			}

			fmt.Printf(">>> %s (found in file: %s)\n", tag, filePath)
		}
	}

	if found {
		return errFoundUnescapedJSXTags
	}

	return nil
}
