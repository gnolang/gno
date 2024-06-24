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

// extractLocalLinks extracts links to local files from the given file content
func extractLocalLinks(fileContent []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	links := make([]string, 0)
	// Regular expression to match markdown links
	re := regexp.MustCompile(`]\((\.\.?/.+?)\)`)

	// Scan file line by line
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "[embedmd]") {
			openPar := strings.Index(line, "(")
			closePar := strings.LastIndex(line, ")")

			link := line[openPar+1 : closePar]

			if pos := strings.Index(link, " "); pos != -1 {
				link = link[:pos]
			}

			links = append(links, link)
			continue
		}

		// Find all matches
		matches := re.FindAllString(line, -1)

		// Extract and print the local file links
		for _, match := range matches {
			// Remove ]( from the beginning and ) from end of link
			match = match[2 : len(match)-1]

			// Remove markdown headers in links
			if pos := strings.Index(match, "#"); pos != -1 {
				match = match[:pos]
			}

			links = append(links, match)
		}
	}

	return links
}

func lintLocalLinks(filepathToLinks map[string][]string, docsPath string) error {
	var found bool
	for filePath, links := range filepathToLinks {
		for _, link := range links {
			path := filepath.Join(docsPath, filepath.Dir(filePath), link)

			if _, err := os.Stat(path); err != nil {
				if !found {
					fmt.Println("Could not find files with the following paths:")
					found = true
				}

				fmt.Printf(">>> %s (found in file: %s)\n", link, filePath)
			}
		}
	}

	if found {
		return errFoundUnreachableLocalLinks
	}

	return nil
}
