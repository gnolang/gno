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

// lintLocalLinks reports local links that do not resolve to a file, and links
// that climb out of root: docs.gno.land is built from root alone, so such a
// link works on GitHub but is a 404 on the site.
func lintLocalLinks(root string, filepathToLinks map[string][]string) (string, error) {
	var (
		found, outside     bool
		output, outsideBuf bytes.Buffer
	)

	for filePath, links := range filepathToLinks {
		for _, link := range links {
			path := filepath.Join(filepath.Dir(filePath), link)
			absSourcePath, _ := filepath.Abs(filePath)

			abs, _ := filepath.Abs(path)
			if rel, err := filepath.Rel(root, abs); err == nil && !filepath.IsLocal(rel) {
				if !outside {
					outsideBuf.WriteString("Local links outside the docs folder (404 on the docs site, link to GitHub instead):\n")
					outside = true
				}
				outsideBuf.WriteString(fmt.Sprintf(">>> %s (found in file: file://%s)\n", link, absSourcePath))
				continue
			}

			if _, err := os.Stat(path); err != nil {
				if !found {
					output.WriteString("Could not find files with the following paths:\n")
					found = true
				}

				output.WriteString(
					fmt.Sprintf(">>> %s (found in file: file://%s)\n", link, absSourcePath),
				)
			}
		}
	}

	output.Write(outsideBuf.Bytes())

	if found {
		return output.String(), errFoundUnreachableLocalLinks
	}

	if outside {
		return output.String(), errFoundLinksOutsideDocs
	}

	return "", nil
}

func shouldIgnoreLink(m string) bool {
	return strings.HasPrefix(m, "http") || strings.HasPrefix(m, "https") || strings.HasPrefix(m, "ws") || strings.HasPrefix(m, "tcp")
}
