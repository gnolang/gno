package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"mvdan.cc/xurls/v2"
	"net/http"
	"strings"
	"sync"
)

func lintLinks(fileUrlMap map[string][]string, ctx context.Context) error {
	// Filter links by prefix & ignore localhost
	validUrls := make(map[string][]string)

	for file, urls := range fileUrlMap {
		file := file
		for _, url := range urls {
			// Look for http & https only
			if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
				// Ignore localhost
				if !strings.Contains(url, "localhost") && !strings.Contains(url, "127.0.0.1") {
					validUrls[file] = append(validUrls[file], url)
				}
			}
		}
	}

	// Setup parallel checking for links
	g, _ := errgroup.WithContext(ctx)

	var (
		lock         sync.Mutex
		notFoundUrls []string
	)

	for filePath, urls := range validUrls {
		filePath := filePath

		if len(urls) != 0 {
			println("extracted for: " + filePath)
			fmt.Println(urls)
		}

		for _, url := range urls {
			url := url
			g.Go(func() error {
				if err := checkUrl(url); err != nil {
					lock.Lock()
					notFoundUrls = append(notFoundUrls, fmt.Sprintf(">>> %s (found in file: %s)", url, filePath))
					lock.Unlock()
				}

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Print out the URLs that returned a 404 along with the file names
	if len(notFoundUrls) > 0 {
		for _, result := range notFoundUrls {
			fmt.Println(result)
		}

		return errFound404Links
	}

	fmt.Println("No broken links were found.")

	return nil
}

// extractUrls extracts URLs from a file and maps them to the file
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

		urls = append(urls, url)
	}

	return urls
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
