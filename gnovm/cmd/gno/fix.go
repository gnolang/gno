package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/fix"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/pmezard/go-difflib/difflib"
)

type fixCfg struct {
	diff bool
	// TODO: -force and -r to select specific fixes.
}

func newFixCmd(io commands.IO) *commands.Command {
	bld := strings.Builder{}
	bld.WriteString(`The gno fix tool allows you to find Gno programs that use old APIs,
and rewrite them to use new APIs.

gno fix rewrites the files in-place. Use -diff to only show a diff of the
changes that should be applied.

The available fixes are the following:
`)
	for _, fx := range fix.Fixes {
		fmt.Fprintf(&bld, "- %s (%s)\n\t%s\n", fx.Name, fx.Date, fx.Desc)
	}
	cfg := &fixCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "fix",
			ShortUsage: "gno fix [flags] [path ...]",
			ShortHelp:  "update and fix old gno source files",
			LongHelp:   bld.String(),
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execFix(cfg, args, io)
		})
}

func (c *fixCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.diff, "diff", false, "only show a diff of the changes that would be applied")
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
		src, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}
		parsed, err := parser.ParseFile(
			fset, file, src,
			parser.SkipObjectResolution|parser.ParseComments,
		)
		if err != nil {
			return err
		}
		// set if any of the fixes changed the AST.
		fixed := false
		for _, fx := range fix.Fixes {
			// wrap in anonymous func so we can recover and wrap errors with file name.
			func() {
				defer func() {
					rec := recover()
					switch rec := rec.(type) {
					case nil:
					case error:
						panic(fmt.Errorf("%s: %w", file, rec))
					default:
						panic(fmt.Errorf("%s: %v", file, rec))
					}
				}()
				fixed = fx.F(parsed) || fixed
			}()
		}
		if !fixed {
			// onto the next file.
			continue
		}
		if cfg.diff {
			var buf bytes.Buffer
			if err := format.Node(&buf, fset, parsed); err != nil {
				return fmt.Errorf("error formatting: %w", err)
			}
			difflib.WriteUnifiedDiff(io.Out(), difflib.UnifiedDiff{
				FromFile: file,
				ToFile:   file,
				A:        difflib.SplitLines(string(src)),
				B:        difflib.SplitLines(buf.String()),
				Context:  3,
			})
		} else {
			f, err := os.Create(file)
			if err != nil {
				return fmt.Errorf("cannot write to dst file: %w", err)
			}
			err = format.Node(f, fset, parsed)
			f.Close()
			if err != nil {
				return fmt.Errorf("error formatting: %w", err)
			}
		}
	}

	return nil
}
