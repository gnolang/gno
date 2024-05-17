package main

import (
	"bufio"
	"fmt"
	"io"
	"mvdan.cc/xurls/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// findFilePaths gathers the file paths for specific file types
func findFilePaths(startPath string) ([]string, error) {
	filePaths := make([]string, 0)

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing file: %w", err)
		}

		// Check if the file is a dir
		if info.IsDir() {
			return nil
		}

		// Check if the file type matches
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// File is not a directory
		filePaths = append(filePaths, path)

		return nil
	}

	// Walk the directory root recursively
	if walkErr := filepath.Walk(startPath, walkFn); walkErr != nil {
		return nil, fmt.Errorf("unable to walk directory, %w", walkErr)
	}

	return filePaths, nil
}

// extractUrls extracts URLs from a file and maps them to the file
func extractUrls(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	cleanup := func() error {
		if closeErr := file.Close(); closeErr != nil {
			return fmt.Errorf("unable to gracefully close file, %w", closeErr)
		}
		return nil
	}

	scanner := bufio.NewScanner(file)
	urls := make(map[string]string)

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

		urls[url] = filePath
	}

	return urls, cleanup()
}

// checkUrl checks if a URL is a 404
func checkUrl(lock *sync.Mutex, url string, filePath string, results *[]string) {
	// Attempt to retrieve the HTTP header
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		// Lock the mutex before appending to results
		lock.Lock()
		*results = append(*results, fmt.Sprintf("%s (found in file: %s)", url, filePath))
		lock.Unlock()
		return
	}

	// Ensure the response body is closed properly
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("could not close response properly: %v", err)
		}
	}(resp.Body)
}
