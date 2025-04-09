package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/fix"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type fixCfg struct{}

func newFixCmd(io commands.IO) *commands.Command {
	cfg := &fixCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "fix",
			ShortUsage: "gno fix [flags] [path ...]",
			ShortHelp:  "update and fix old gno source files",
			LongHelp:   "The `gno fix` tool processes, fixes, and cleans up `gno` source files.",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execFix(cfg, args, io)
		})
}

func (c *fixCfg) RegisterFlags(fs *flag.FlagSet) {
}

func execFix(cfg *fixCfg, args []string, io commands.IO) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	paths, err := targetsFromPatterns(args)
	if err != nil {
		return fmt.Errorf("unable to get targets paths from patterns: %w", err)
	}

	files, err := gnoFilesFromArgs(paths)
	if err != nil {
		return fmt.Errorf("unable to gather gno files: %w", err)
	}

	for _, file := range files {
		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(
			fset, file, nil,
			parser.SkipObjectResolution|parser.ParseComments,
		)
		if err != nil {
			return err
		}
		for _, fx := range fix.Fixes {
			res, err := func() (res bool, err error) {
				defer func() {
					rec := recover()
					switch rec := rec.(type) {
					case nil:
						return
					case error:
						err = rec
					default:
						err = fmt.Errorf("panic: %v", rec)
					}
				}()
				res = fx.F(parsed)
				return
			}()
			if err != nil {
				io.ErrPrintfln("error fixing %q: %v", file, err)
			}
			if res {
				var buf bytes.Buffer
				if err := format.Node(&buf, fset, parsed); err != nil {
					io.ErrPrintfln("format error: %s", err.Error())
					return nil
				}
				fmt.Println("converted", file)
			}
		}
	}

	return nil
}
