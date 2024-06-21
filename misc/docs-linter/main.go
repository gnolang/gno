package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"strings"
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
			ShortHelp: `Lints the .md files in the given folder & subfolders.
Checks for 404 links, as well as improperly escaped JSX tags.`,
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

	absPath, err := filepath.Abs(cfg.docsPath)
	if err != nil {
		return fmt.Errorf("error getting absolute path for docs folder: %w", err)
	}

	fmt.Printf("Linting %s...\n", absPath)

	// Find docs files to lint
	mdFiles, err := findFilePaths(cfg.docsPath)
	if err != nil {
		return fmt.Errorf("error finding .md files: %w", err)
	}

	// Make storage maps for tokens to analyze
	fileUrlMap := make(map[string][]string)       // file path > [urls]
	fileJSXMap := make(map[string][]string)       // file path > [JSX items]
	fileLocalLinkMap := make(map[string][]string) // file path > [local links]

	// Extract tokens from files
	for _, filePath := range mdFiles {
		// Read file content once and pass it to linters
		fileContents, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Execute JSX extractor
		fileJSXMap[filePath] = extractJSX(fileContents)

		// Execute URL extractor
		fileUrlMap[filePath] = extractUrls(fileContents)

		// Execute local link extractor
		fileLocalLinkMap[filePath] = extractLocalLinks(fileContents)
	}

	// Run linters in parallel
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		return lintJSX(fileJSXMap)
	})

	g.Go(func() error {
		return lintURLs(fileUrlMap, ctx)
	})

	g.Go(func() error {
		return lintLocalLinks(fileLocalLinkMap, cfg.docsPath)
	})

	if err := g.Wait(); err != nil {
		return errFoundLintItems
	}

	fmt.Println("Lint complete, no issues found.")
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
