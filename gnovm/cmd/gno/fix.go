package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/fix"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/pmezard/go-difflib/difflib"
)

type fixCmd struct {
	verbose   bool
	diff      bool
	fix       string
	fixFilter func(s string) bool
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
	fs.BoolVar(&c.diff, "diff", false, "show diffs of files which are meant to be changed (without writing to them)")
	fs.StringVar(&c.fix, "fix", "", "comma-separated of fixes to run. by default all known fixes.")
}

func execFix(cmd *fixCmd, args []string, cio commands.IO) error {
	// Show a help message by default.
	if len(args) == 0 {
		return flag.ErrHelp
	}

	if cmd.fix != "" {
		fixes := strings.Split(cmd.fix, ",")
		cmd.fixFilter = func(s string) bool { return slices.Contains(fixes, s) }
	} else {
		cmd.fixFilter = func(s string) bool { return true }
	}

	targets, err := targetsFromPatterns(args)
	if err != nil {
		return fmt.Errorf("unable to get targets paths from patterns: %w", err)
	}

	for _, targ := range targets {
		files, err := gnoFilesFromArgs([]string{targ})
		if err != nil {
			return fmt.Errorf("unable to gather gno files: %w", err)
		}
		if len(files) == 0 {
			continue
		}
		// individual file: process fix without version
		if len(files) == 1 && files[0] == targ {
			if err := cmd.processFix(cio, files, nil); err != nil {
				return err
			}
		}
		gm, isDotMod := gnoFixParseGnomod(targ)
		if err := cmd.processFix(cio, files, gm); err != nil {
			return err
		}
		if !cmd.diff {
			if err = gm.WriteFile(filepath.Join(targ, "gnomod.toml")); err != nil {
				return fmt.Errorf("writing gnomod.toml: %w", err)
			}
			if isDotMod {
				err := os.Remove(filepath.Join(targ, "gno.mod"))
				if err != nil {
					return fmt.Errorf("removing gno.mod: %w", err)
				}
			}
		}
	}

	return err
}

func gnoFixParseGnomod(dir string) (mod *gnomod.File, isDotMod bool) {
	fpath := filepath.Join(dir, "gnomod.toml")
	mod, err := gnomod.ParseFilepath(fpath)
	if errors.Is(err, fs.ErrNotExist) {
		// We try a lazy migration from gno.mod if it exists and is valid.
		deprecatedDotmod := filepath.Join(dir, "gno.mod")
		mod, err = gnomod.ParseFilepath(deprecatedDotmod)
		if err != nil {
			// It doesn't exist or we can't parse it.
			// Make a temporary gnomod.toml (but don't write it yet)
			modstr := gno.GenGnoModMissing("gno.land/r/xxx_myrealm_xxx/xxx_fixme_xxx")
			mod, err = gnomod.ParseBytes("gnomod.toml", []byte(modstr))
			if err != nil {
				panic(fmt.Errorf("unexpected panic parsing default gnomod.toml bytes: %w", err))
			}
		} else {
			isDotMod = true
		}
	}
	return
}

func (cmd *fixCmd) processFix(cio commands.IO, files []string, gm *gnomod.File) error {
	var newVersion string
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
			if fx.F == nil || !cmd.fixFilter(fx.Name) {
				continue
			}
			if fx.Version != "" && gm != nil {
				cmpv, ok := gno.CompareVersions(fx.Version, gm.Gno)
				if ok && cmpv <= 0 {
					if cmd.verbose {
						cio.ErrPrintfln(
							"%s: %s: skipping fix (fix version %q <= gnomod version %q)",
							file, fx.Name, fx.Version, gm.Gno,
						)
					}
					continue
				}
				newVersion = fx.Version
			}
			// wrap in anonymous func so we can recover and wrap errors with file name.
			func() {
				defer func() {
					rec := recover()
					switch rec := rec.(type) {
					case nil:
					case error:
						panic(fmt.Errorf("%s: %s: %w", file, fx.Name, rec))
					default:
						panic(fmt.Errorf("%s: %s: %v", file, fx.Name, rec))
					}
				}()
				fixed = fx.F(parsed) || fixed
			}()
		}
		if !fixed {
			// onto the next file.
			continue
		}
		if cmd.diff {
			var buf bytes.Buffer
			if err := format.Node(&buf, fset, parsed); err != nil {
				return fmt.Errorf("error formatting: %w", err)
			}
			difflib.WriteUnifiedDiff(cio.Out(), difflib.UnifiedDiff{
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
	if gm != nil && newVersion != "" {
		gm.Gno = newVersion
	}
	return nil
}
