package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/sync/errgroup"
	"io"
	"mvdan.cc/xurls/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	errEmptyPath     = errors.New("you need to pass in a path to scan")
	err404Link       = errors.New("link returned a 404")
	errFound404Links = errors.New("found links resulting in a 404 response status")
)

type cfg struct {
	docsPath string
}

func main() {
	cfg := &cfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "docs-linter",
			ShortUsage: "docs-linter [flags]",
			ShortHelp:  "Finds broken 404 links in the .md files in the given folder & subfolders",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execLint(cfg, ctx)
		})

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *cfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.docsPath,
		"path",
		"./",
		"path to dir to walk for .md files",
	)
}

func execLint(cfg *cfg, ctx context.Context) error {
	if cfg.docsPath == "" {
		return errEmptyPath
	}

	fmt.Println("Linting docs/")

	mdFiles, err := findFilePaths(cfg.docsPath)
	if err != nil {
		return fmt.Errorf("error finding .md files: %w", err)
	}

	urlFileMap := make(map[string]string)
	for _, filePath := range mdFiles {
		// Extract URLs from each file
		urls, err := extractUrls(filePath)
		if err != nil {
			fmt.Printf("Error extracting URLs from file: %s, %v", filePath, err)
			continue
		}
		// For each url, save what file it was found in
		for url, file := range urls {
			urlFileMap[url] = file
		}
	}

	// Filter links by prefix & ignore localhost
	var validUrls []string
	for url := range urlFileMap {
		// Look for http & https only
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			// Ignore localhost
			if !strings.Contains(url, "localhost") && !strings.Contains(url, "127.0.0.1") {
				validUrls = append(validUrls, url)
			}
		}
	}

	// Setup parallel checking for links
	g, _ := errgroup.WithContext(ctx)

	var (
		lock         sync.Mutex
		notFoundUrls []string
	)

	for _, url := range validUrls {
		url := url
		g.Go(func() error {
			if err := checkUrl(url); err != nil {
				lock.Lock()
				notFoundUrls = append(notFoundUrls, fmt.Sprintf(">>> %s (found in file: %s)", url, urlFileMap[url]))
				lock.Unlock()
			}

			return nil
		})
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

	return nil
}

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
