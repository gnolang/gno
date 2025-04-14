package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Valid start to an embedmd link
const embedmd = `[embedmd]:# `

// Regular expression to match markdown links
var regex = regexp.MustCompile(`]\(([^)]+)\)`)

// extractLocalLinks extracts links to local files from the given file content
func extractLocalLinks(fileContent []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	links := make([]string, 0)

	// Scan file line by line
	for scanner.Scan() {
		line := scanner.Text()

		// Check for embedmd links
		if embedmdPos := strings.Index(line, embedmd); embedmdPos != -1 {
			link := line[embedmdPos+len(embedmd)+1:]

			// Find closing parentheses
			if closePar := strings.LastIndex(link, ")"); closePar != -1 {
				link = link[:closePar]
			}

			// Remove space
			if pos := strings.Index(link, " "); pos != -1 {
				link = link[:pos]
			}

			// Add link to be checked
			links = append(links, link)
			continue
		}

		// Find all matches
		matches := regex.FindAllString(line, -1)

		// Extract and print the local file links
		for _, match := range matches {
			// Remove ]( from the beginning and ) from end of link
			match = match[2 : len(match)-1]

			// Ignore http, https, tcp, ws links
			if shouldIgnoreLink(match) {
				continue
			}

			// Remove markdown headers in links
			if pos := strings.Index(match, "#"); pos != -1 {
				match = match[:pos]
			}

			links = append(links, match)
		}
	}

	return links
}

func lintLocalLinks(filepathToLinks map[string][]string) (string, error) {
	var (
		found  bool
		output bytes.Buffer
	)

	for filePath, links := range filepathToLinks {
		for _, link := range links {
			path := filepath.Join(filepath.Dir(filePath), link)
			if _, err := os.Stat(path); err != nil {
				if !found {
					output.WriteString("Could not find files with the following paths:\n")
					found = true
				}
				// Make the source file clickable (file:// URI)
				absSourcePath, _ := filepath.Abs(filePath)
				output.WriteString(
					fmt.Sprintf(">>> %s (found in file: file://%s)\n", link, absSourcePath),
				)
			}
		}
	}

	if found {
		return output.String(), errFoundUnreachableLocalLinks
	}

	return "", nil
}

func shouldIgnoreLink(m string) bool {
	return strings.HasPrefix(m, "http") || strings.HasPrefix(m, "https") || strings.HasPrefix(m, "ws") || strings.HasPrefix(m, "tcp")
}
