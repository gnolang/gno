package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	dt "github.com/gnolang/gno/gnovm/pkg/doctest"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type doctestCfg struct {
	markdownPath string
	// codeIndex    int
	runPattern string
}

func newDoctestCmd(io commands.IO) *commands.Command {
	cfg := &doctestCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "doctest",
			ShortUsage: "doctest -path <markdown_file_path> [-run <pattern>]",
			ShortHelp:  "executes a specific code block from a markdown file",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDoctest(cfg, args, io)
		},
	)
}

func (c *doctestCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.markdownPath,
		"path",
		"",
		"path to the markdown file",
	)
	fs.StringVar(
		&c.runPattern,
		"run",
		"",
		"pattern to match code block names",
	)
}

func execDoctest(cfg *doctestCfg, _ []string, io commands.IO) error {
	if cfg.markdownPath == "" {
		return fmt.Errorf("markdown file path is required")
	}

	content, err := fetchMarkdown(cfg.markdownPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	results, err := dt.ExecuteMatchingCodeBlock(content, cfg.runPattern)
	if err != nil {
		return fmt.Errorf("failed to execute code block: %w", err)
	}

	if len(results) == 0 {
		io.Println("No code blocks matched the pattern")
		return nil
	}

	io.Println("Execution Result:")
	io.Println(strings.Join(results, "\n\n"))

	return nil
}

// fetchMarkdown reads a markdown file and returns its content
func fetchMarkdown(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}
