package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/doctest"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type doctestCfg struct {
	path string
	index int
}

func newDoctestCmd() *commands.Command {
	cfg := &doctestCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "doctest",
			ShortUsage: "doctest [flags]",
			ShortHelp:  "executes a code block from a markdown file",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDoctest(cfg, args)
		},
	)
}

func (c *doctestCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.path,
		"markdown-path",
		"",
		"path to the markdown file",
	)

	fs.IntVar(
		&c.index,
		"index",
		0,
		"index of the code block to execute",
	)
}

func execDoctest(cfg *doctestCfg, args []string) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	if cfg.path == "" {
		return fmt.Errorf("missing markdown-path flag. Please provide a path to the markdown file")
	}

	content, err := doctest.ReadMarkdownFile(cfg.path)
	if err != nil {
		return err
	}

	codeBlocks := doctest.GetCodeBlocks(content)
	if cfg.index >= len(codeBlocks) {
		return fmt.Errorf("code block index out of range. max index: %d", len(codeBlocks)-1)
	}

	codeblock := codeBlocks[cfg.index]
	result, err := doctest.ExecuteCodeBlock(codeblock)
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}
