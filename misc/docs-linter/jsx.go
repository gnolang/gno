package main

import (
	"context"
	"regexp"
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
	reJSXHTML := regexp.MustCompile("(?s)<[^>]+>")

	return reJSXHTML.FindAllString(contentNoInlineCode, -1)
}

func lintJSX(fileUrlMap map[string][]string, ctx context.Context) error {

	return nil
}
