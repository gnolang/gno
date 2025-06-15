package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"mvdan.cc/xurls/v2"
)

// extractUrls extracts urls from given file content
func extractUrls(fileContent []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	urls := make([]string, 0)

	// Scan file line by line
	for scanner.Scan() {
		line := scanner.Text()

		// Extract links
		rxStrict := xurls.Strict()
		url := rxStrict.FindString(line)

		// Check for empty links and skip them
		if url == " " || len(url) == 0 {
			continue
		}

		// Look for http & https only
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			// Ignore localhost
			if !strings.Contains(url, "localhost") &&
				!strings.Contains(url, "127.0.0.1") &&
				// placeholder for examples
				!strings.Contains(url, "example.land") {
				urls = append(urls, url)
			}
		}
	}

	return urls
}

func lintURLs(ctx context.Context, filepathToURLs map[string][]string, treatAsError bool) (string, error) {
	// Setup parallel checking for links
	g, _ := errgroup.WithContext(ctx)

	var (
		lock            sync.Mutex
		output          bytes.Buffer
		hasInvalidLinks bool
	)

	for filePath, urls := range filepathToURLs {
		filePath := filePath
		for _, url := range urls {
			url := url
			g.Go(func() error {
				if err := checkUrl(url); err != nil {
					lock.Lock()
					if !hasInvalidLinks {
						output.WriteString("Remote links that need checking:\n")
						hasInvalidLinks = true
					}

					output.WriteString(fmt.Sprintf(">>> %s (found in file: %s)\n", url, filePath))
					lock.Unlock()
				}

				return nil
			})
		}
	}

	// Check for possible thread errors
	if err := g.Wait(); err != nil {
		return "", err
	}

	if !treatAsError {
		errFound404Links = nil
	}
	if hasInvalidLinks {
		return output.String(), errFound404Links
	}

	return "", nil
}

// checkUrl checks if a URL is a 404
func checkUrl(url string) error {
	// Attempt to retrieve the HTTP header
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		return err404Link
	}

	// Ensure the response body is closed properly
	cleanup := func(Body io.ReadCloser) error {
		if err := Body.Close(); err != nil {
			return fmt.Errorf("could not close response properly: %w", err)
		}

		return nil
	}

	return cleanup(resp.Body)
}
