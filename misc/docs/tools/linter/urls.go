package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"mvdan.cc/xurls/v2"
)

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

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
				!strings.Contains(url, "example.land") &&
				// deployment-specific hosts whose uptime is not a CI concern
				!strings.Contains(url, "staging.gno.land") {
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

// checkUrl checks if a URL is a 404, retrying on transient errors.
func checkUrl(url string) error {
	const maxRetries = 3

	for attempt := range maxRetries {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}

		resp, err := httpClient.Get(url)
		if err != nil {
			// Network error: retry unless this is the last attempt.
			if attempt < maxRetries-1 {
				continue
			}

			return err404Link
		}
		resp.Body.Close()

		switch {
		case resp.StatusCode == http.StatusNotFound:
			// 404 is a definitive failure; no point retrying.
			return err404Link
		case resp.StatusCode == http.StatusTooManyRequests:
			// Rate-limited; retry.
			if attempt < maxRetries-1 {
				continue
			}

			return err404Link
		}

		// Treat everything else (including 5xx) as reachable.
		return nil
	}

	return nil
}
