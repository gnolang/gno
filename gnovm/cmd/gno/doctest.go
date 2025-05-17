package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	dt "github.com/gnolang/gno/gnovm/pkg/doctest"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/sync/errgroup"
)

type doctestCfg struct {
	markdownPath string
	runPattern   string
	timeout      time.Duration
}

func newDoctestCmd(io commands.IO) *commands.Command {
	cfg := &doctestCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "doctest",
			ShortUsage: "doctest -path <markdown_file_path> [-run <pattern>] [-timeout <duration>]",
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
	fs.DurationVar(
		&c.timeout,
		"timeout",
		time.Second*30,
		"timeout for code execution (e.g., 30s, 1m)",
	)
}

func execDoctest(cfg *doctestCfg, _ []string, io commands.IO) error {
	if cfg.markdownPath == "" {
		return errors.New("markdown file path is required")
	}

	content, err := fetchMarkdown(cfg.markdownPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)

	var results []string
	g.Go(func() error {
		res, err := dt.ExecuteMatchingCodeBlock(gctx, content, cfg.runPattern)
		if err != nil {
			return err
		}
		results = res
		return nil
	})

	if err := g.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("execution timed out after %v", cfg.timeout)
		}
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
