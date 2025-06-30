package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/sync/errgroup"
)

type cfg struct {
	docsPath       string
	treatUrlsAsErr bool
}

func main() {
	cfg := &cfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "docs-linter",
			ShortUsage: "docs-linter [flags]",
			ShortHelp: `Lints the .md files in the given folder & subfolders.
Checks for 404 links (local and remote), as well as improperly escaped JSX tags.`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			res, err := execLint(cfg, ctx)
			if len(res) != 0 {
				fmt.Println(res)
			}

			return err
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

	fs.BoolVar(
		&c.treatUrlsAsErr,
		"treat-urls-as-err",
		true,
		"treat URL 404s as errors instead of warnings",
	)
}

func execLint(cfg *cfg, ctx context.Context) (string, error) {
	if cfg.docsPath == "" {
		return "", errEmptyPath
	}

	absPath, err := filepath.Abs(cfg.docsPath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path for docs folder: %w", err)
	}

	// Main buffer to write to the end user after linting
	var output bytes.Buffer
	output.WriteString(fmt.Sprintf("Linting %s...\n", absPath))

	// Find docs files to lint
	mdFiles, err := findFilePaths(cfg.docsPath)
	if err != nil {
		return "", fmt.Errorf("error finding .md files: %w", err)
	}

	// Make storage maps for tokens to analyze
	filepathToURLs := make(map[string][]string)      // file path > [urls]
	filepathToJSX := make(map[string][]string)       // file path > [JSX items]
	filepathToLocalLink := make(map[string][]string) // file path > [local links]

	// Extract tokens from files
	for _, filePath := range mdFiles {
		// Read file content once and pass it to linters
		fileContents, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}

		// Execute JSX extractor
		filepathToJSX[filePath] = extractJSX(fileContents)

		// Execute URL extractor
		filepathToURLs[filePath] = extractUrls(fileContents)

		// Execute local link extractor
		filepathToLocalLink[filePath] = extractLocalLinks(fileContents)
	}

	// Run linters in parallel
	g, _ := errgroup.WithContext(ctx)

	var writeLock sync.Mutex

	g.Go(func() error {
		res, err := lintJSX(filepathToJSX)
		if err != nil {
			writeLock.Lock()
			output.WriteString(res)
			writeLock.Unlock()
		}

		return err
	})

	g.Go(func() error {
		res, err := lintURLs(ctx, filepathToURLs, cfg.treatUrlsAsErr)
		writeLock.Lock()
		output.WriteString(res)
		writeLock.Unlock()

		return err
	})

	g.Go(func() error {
		res, err := lintLocalLinks(filepathToLocalLink)
		if err != nil {
			writeLock.Lock()
			output.WriteString(res)
			writeLock.Unlock()
		}

		return err
	})

	if err = g.Wait(); err != nil {
		return output.String(), errFoundLintItems
	}

	output.WriteString("Lint complete, no issues found.")
	return output.String(), nil
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
