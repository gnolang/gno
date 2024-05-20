package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrEmptyPath = errors.New("you need to pass in a path to scan")
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
		return ErrEmptyPath
	}

	abs, err := filepath.Abs(cfg.docsPath)
	if err != nil {
		return err
	}

	fmt.Printf("Linting %s...\n", abs)

	mdFiles, err := findFilePaths(cfg.docsPath)
	if err != nil {
		return fmt.Errorf("error finding .md files: %w", err)
	}

	urlFileMap := make(map[string]string)
	for _, filePath := range mdFiles {
		// Extract URLs from each file
		urls, err := extractUrls(filePath)
		if err != nil {
			fmt.Println("Error extracting URLs from file:", filePath, err)
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
	var lock sync.Mutex
	var notFoundUrls []string
	g, ctx := errgroup.WithContext(ctx)

	for _, url := range validUrls {
		url := url

		g.Go(func() error {
			err := checkUrl(&lock, url, urlFileMap[url], &notFoundUrls)
			if err != nil {
				return err
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Print out the URLs that returned a 404 along with the file names
	if len(notFoundUrls) > 0 {
		fmt.Println("The following URLs are broken or returned a 404 status:")
		for _, result := range notFoundUrls {
			fmt.Println(result)
		}
	} else {
		fmt.Println("No URLs returned a 404 status.")
	}

	return nil
}
