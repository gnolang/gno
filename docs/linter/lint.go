package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// Function to read .md files from a directory
func readMdFiles(dir string) ([]string, error) {
	var mdFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	return mdFiles, err
}

// Function to extract URLs from a file and map them to the file
func extractUrls(filePath string) (map[string]string, error) {
	urls := make(map[string]string)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	urlRegex := regexp.MustCompile(`https?://[^\s\)\],\\]+`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := urlRegex.FindAllString(line, -1)
		for _, url := range matches {
			// Remove any trailing backslashes
			url = strings.TrimSuffix(url, `\`)
			urls[url] = filePath
		}
	}
	return urls, scanner.Err()
}

// Function to check if a URL is a 404
func checkUrl(url string, filePath string, wg *sync.WaitGroup, mu *sync.Mutex, results *[]string) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		mu.Lock()
		*results = append(*results, fmt.Sprintf("%s (found in file: %s)", url, filePath))
		mu.Unlock()
	}
}

func main() {
	// Parse command-line flag for the directory
	dir := flag.String(
		"dir",
		".",
		"Directory containing .md files")
	flag.Parse()

	// Step 1: Read .md files from directory
	mdFiles, err := readMdFiles(*dir)
	if err != nil {
		fmt.Println("Error reading .md files:", err)
		return
	}

	urlFileMap := make(map[string]string)
	for _, filePath := range mdFiles {
		// Step 2: Extract URLs from each file
		urls, err := extractUrls(filePath)
		if err != nil {
			fmt.Println("Error extracting URLs from file:", filePath, err)
			continue
		}
		for url, file := range urls {
			urlFileMap[url] = file
		}
	}

	// Step 3: Filter out localhost and non-http/https URLs
	var validUrls []string
	for url := range urlFileMap {
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			if !strings.Contains(url, "localhost") {
				validUrls = append(validUrls, url)
			}
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var notFoundUrls []string

	// Step 4: Check the status of each URL
	for _, url := range validUrls {
		wg.Add(1)
		go checkUrl(url, urlFileMap[url], &wg, &mu, &notFoundUrls)
	}

	wg.Wait()

	// Print out the URLs that returned a 404 along with the file names
	if len(notFoundUrls) > 0 {
		fmt.Println("The following URLs returned a 404 status:")
		for _, result := range notFoundUrls {
			fmt.Println(result)
		}
	} else {
		fmt.Println("No URLs returned a 404 status.")
	}
}
