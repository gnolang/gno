package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	dt "github.com/gnolang/gno/gnovm/pkg/doctest"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type doctestCfg struct {
	markdownPath string
	codeIndex    int
}

func newDoctestCmd(io commands.IO) *commands.Command {
	cfg := &doctestCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "doctest",
			ShortUsage: "doctest -path <markdown_file_path> -index <code_block_index>",
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

	fs.IntVar(
		&c.codeIndex,
		"index",
		-1,
		"index of the code block to execute",
	)
}

func execDoctest(cfg *doctestCfg, _ []string, io commands.IO) error {
	if cfg.markdownPath == "" {
		return fmt.Errorf("markdown file path is required")
	}

	if cfg.codeIndex < 0 {
		return fmt.Errorf("code block index must be non-negative")
	}

	content, err := fetchMarkdown(cfg.markdownPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	codeBlocks := dt.GetCodeBlocks(content)
	if cfg.codeIndex >= len(codeBlocks) {
		return fmt.Errorf("invalid code block index: %d", cfg.codeIndex)
	}

	selectedCodeBlock := codeBlocks[cfg.codeIndex]
	result, err := dt.ExecuteCodeBlock(selectedCodeBlock, dt.STDLIBS_DIR)
	if err != nil {
		return fmt.Errorf("failed to execute code block: %w", err)
	}

	io.Println("Execution Result:")
	io.Println(result)

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
