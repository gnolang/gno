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
	"golang.org/x/tools/txtar"
)

type fixCmd struct {
	verbose   bool
	diff      bool
	fix       string
	fixFilter func(s fix.Fix) bool
}

func newFixCmd(cio commands.IO) *commands.Command {
	cmd := &fixCmd{}

	bld := strings.Builder{}
	bld.WriteString(`The gno fix tool allows you to find Gno programs that use old APIs,
and rewrite them to use new APIs.
gno fix rewrites the files in-place. Use -diff to only show a diff of the
changes that should be applied. Both .gno files
and .txtar archives can be passed in the same invocation.
The available fixes are the following:
`)
	for _, fx := range fix.Fixes {
		desc := strings.ReplaceAll(fx.Desc, "\n", "\n\t")
		disabled := ""
		if fx.DisabledByDefault {
			disabled = " - disabled by default"
		}
		fmt.Fprintf(&bld, "- %s (%s)%s\n\t%s\n", fx.Name, fx.Date, disabled, desc)
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
	fs.StringVar(&c.fix, "fix", "", "comma-separated of fixes to run. refer to the list for the enabled fixes by default.")
}

func execFix(cmd *fixCmd, args []string, cio commands.IO) error {
	// Show a help message by default.
	if len(args) == 0 {
		return flag.ErrHelp
	}

	if cmd.fix != "" {
		fixes := strings.Split(cmd.fix, ",")
		cmd.fixFilter = func(fx fix.Fix) bool { return slices.Contains(fixes, fx.Name) }
	} else {
		cmd.fixFilter = func(fx fix.Fix) bool { return !fx.DisabledByDefault }
	}

	targets, err := targetsFromPatterns(args)
	if err != nil {
		return fmt.Errorf("unable to get targets paths from patterns: %w", err)
	}

	for _, targ := range targets {
		// txtar file must be an explicit target and not a directory.
		txtarFile, err := txtarFileFromArg(targ)
		if err != nil {
			return fmt.Errorf("unable to gather txtar file: %w", err)
		}
		if txtarFile != "" {
			if err := cmd.processFixTxtar(txtarFile); err != nil {
				return fmt.Errorf("unable to process txtar %q: %w", txtarFile, err)
			}
			continue
		}
		files, err := gnoFilesFromArgs([]string{targ})
		if err != nil {
			return fmt.Errorf("unable to gather gno files: %w", err)
		}
		if len(files) == 0 {
			continue
		}
		// individual file: process fix without version
		if len(files) == 1 && files[0] == cleanPath(targ) {
			if err := cmd.processFix(cio, files, nil); err != nil {
				return err
			}
			continue
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

func txtarFileFromArg(target string) (string, error) {
	info, err := os.Stat(target)
	if err != nil {
		return "", fmt.Errorf("invalid file %q: %w", target, err)
	}

	if !info.IsDir() && isTxtarFile(info) {
		return cleanPath(target), nil
	}

	return "", nil
}

func isTxtarFile(f fs.FileInfo) bool {
	return strings.HasSuffix(f.Name(), ".txtar") && !f.IsDir()
}

func (c *fixCmd) processFixTxtar(file string) error {
	archive, err := txtar.ParseFile(file)
	if err != nil {
		return err
	}
	// group files by folder to handle error gnomod versions.
	filesByDir := map[string][]*txtar.File{}
	for i, f := range archive.Files {
		dir := filepath.Dir(f.Name)
		filesByDir[dir] = append(filesByDir[dir], &archive.Files[i])
	}
	for _, files := range filesByDir {
		err = c.processFixTxtarDir(files)
		if err != nil {
			return err
		}
	}
	if !c.diff {
		archive := txtar.Format(archive)
		info, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("stat txtar file: %w", err)
		}
		err = os.WriteFile(file, archive, info.Mode())
		if err != nil {
			return fmt.Errorf("error writing txtar file: %w", err)
		}
	}
	return nil
}

func (c *fixCmd) processFixTxtarDir(files []*txtar.File) error {
	var gm *gnomod.File
	for _, f := range files {
		if f.Name == "gnomod.toml" {
			gm, _ = gnomod.ParseBytes("gnomod.toml", f.Data)
			break
		}
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name, ".gno") {
			continue
		}
		buf, fixed, err := c.applyFixesToFile(f.Name, f.Data, gm)
		if err != nil || !fixed {
			// NOTE: some txtar voluntarily expose invalid gno files.
			continue
		}
		if fixed {
			if c.diff {
				err := difflib.WriteUnifiedDiff(os.Stdout, difflib.UnifiedDiff{
					FromFile: f.Name,
					ToFile:   f.Name,
					A:        difflib.SplitLines(string(f.Data)),
					B:        difflib.SplitLines(string(buf)),
					Context:  3,
				})
				if err != nil {
					return err
				}
			} else {
				f.Data = buf
			}
		}
	}
	return nil
}

func (c *fixCmd) processFix(cio commands.IO, files []string, gm *gnomod.File) error {
	var newVersion string
	for _, file := range files {
		if c.verbose {
			cio.ErrPrintln(file)
		}
		src, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}
		buf, fixed, err := c.applyFixesToFile(file, src, gm)
		if err != nil {
			return err
		}
		if !fixed {
			// onto the next file.
			continue
		}
		if c.diff {
			err := difflib.WriteUnifiedDiff(cio.Out(), difflib.UnifiedDiff{
				FromFile: file,
				ToFile:   file,
				A:        difflib.SplitLines(string(src)),
				B:        difflib.SplitLines(string(buf)),
				Context:  3,
			})
			if err != nil {
				return err
			}
		} else {
			err = os.WriteFile(file, buf, 0o644)
			if err != nil {
				return fmt.Errorf("cannot write to dst file: %w", err)
			}
		}
	}
	if gm != nil && newVersion != "" {
		gm.Gno = newVersion
	}
	return nil
}

func (c *fixCmd) applyFixesToFile(
	filename string,
	src []byte,
	gm *gnomod.File,
) (rewritten []byte, fixed bool, err error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(
		fset, filename, src,
		parser.SkipObjectResolution|parser.ParseComments,
	)
	if err != nil {
		return nil, false, err
	}

	fixed = false
	for _, fx := range fix.Fixes {
		if fx.F == nil || !c.fixFilter(fx) {
			continue
		}
		if fx.Version != "" && gm != nil {
			cmpv, ok := gno.CompareVersions(fx.Version, gm.Gno)
			if ok && cmpv <= 0 {
				if c.verbose {
					fmt.Printf("%s: %s: skipping fix (fix version %q <= gnomod version %q)\n",
						filename, fx.Name, fx.Version, gm.Gno)
				}
				continue
			}
		}
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					switch rec := rec.(type) {
					case error:
						panic(fmt.Errorf("%s: %s: %w", filename, fx.Name, rec))
					default:
						panic(fmt.Errorf("%s: %s: %v", filename, fx.Name, rec))
					}
				}
			}()
			fixed = fx.F(parsed) || fixed
		}()
	}

	if !fixed {
		return nil, false, nil
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, parsed); err != nil {
		return nil, false, fmt.Errorf("error formatting: %w", err)
	}

	return buf.Bytes(), true, nil
}
