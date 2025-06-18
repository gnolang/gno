package main

import (
	"context"
	"flag"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/fix"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type fixCmd struct {
	verbose        bool
	rootDir        string
	filetestsOnly  bool
	filetestsMatch string
	diff           bool
}

func newFixCmd(cio commands.IO) *commands.Command {
	cmd := &fixCmd{}

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

	return commands.NewCommand(
		commands.Metadata{
			Name:       "fix",
			ShortUsage: "fix [flags] [<package>...]",
			ShortHelp:  "update and fix old gno source files",
			LongHelp:   bld.String(),
		},
		cmd,
		func(_ context.Context, args []string) error {
			return execFix(cmd, args, cio)
		},
	)
}

func (c *fixCmd) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.verbose, "v", false, "verbose output when fixing")
	fs.StringVar(&c.rootDir, "root-dir", "", "clone location of github.com/gnolang/gno (gno tries to guess it)")
	fs.BoolVar(&c.filetestsOnly, "filetests-only", false, "dir only contains filetests. not recursive.")
	fs.StringVar(&c.filetestsMatch, "filetests-match", "", "if --filetests-only=true, filters by substring match.")
}

func execFix(cmd *fixCmd, args []string, cio commands.IO) error {
	// Show a help message by default.
	if len(args) == 0 {
		return flag.ErrHelp
	}

	// Guess cmd.RootDir.
	if cmd.rootDir == "" {
		cmd.rootDir = gnoenv.RootDir()
	}

	opts := fix.Options{
		RootDir: cmd.rootDir,
	}
	if cmd.verbose {
		opts.Error = cio.Err()
	}

	paths, err := targetsFromPatterns(args)
	if err != nil {
		return fmt.Errorf("unable to get targets paths from patterns: %w", err)
	}

	for _, fx := range fix.Fixes {
		if fx.DirsF == nil {
			continue
		}
		err := fx.DirsF(opts, paths)
		if err != nil {
			return fmt.Errorf("%s: %w", fx.Name, err)
		}
	}

	files, err := gnoFilesFromArgs(paths)
	if err != nil {
		return fmt.Errorf("unable to gather gno files: %w", err)
	}

	for _, file := range files {
		if cmd.verbose {
			cio.ErrPrintln(file)
		}
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
			if fx.F == nil {
				continue
			}
			// wrap in anonymous func so we can recover and wrap errors with file name.
			func() {
				defer func() {
					rec := recover()
					switch rec := rec.(type) {
					case nil:
					case error:
						panic(fmt.Errorf("%s: %s: %w", fx.Name, file, rec))
					default:
						panic(fmt.Errorf("%s: %s: %v", fx.Name, file, rec))
					}
				}()
				fixed = fx.F(opts, parsed) || fixed
			}()
		}
		if !fixed {
			// onto the next file.
			continue
		}
		// XXX: diff option - https://github.com/thehowl/gno/pull/1/files
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
	return err
}
