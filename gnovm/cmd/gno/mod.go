package main

import (
	"context"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"go.uber.org/multierr"
)

func newModCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "mod",
			ShortUsage: "mod <command>",
			ShortHelp:  "manage gno.mod",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newModInitCmd(),
		newModTidy(io),
		newModWhy(io),
	)

	return cmd
}

func newModInitCmd() *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [module-path]",
			ShortHelp:  "initialize gno.mod file in current directory",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execModInit(args)
		},
	)
}

func newModTidy(io commands.IO) *commands.Command {
	cfg := &modTidyCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "tidy",
			ShortUsage: "tidy [flags]",
			ShortHelp:  "add missing and remove unused modules",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execModTidy(cfg, args, io)
		},
	)
}

func newModWhy(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "why",
			ShortUsage: "why <package> [<package>...]",
			ShortHelp:  "Explains why modules are needed",
			LongHelp: `Explains why modules are needed.

gno mod why shows a list of files where specified packages or modules are
being used, explaining why those specified packages or modules are being
kept by gno mod tidy.

The output is a sequence of stanzas, one for each module/package name
specified, separated by blank lines. Each stanza begins with a
comment line "# module" giving the target module/package. Subsequent lines
show files that import the specified module/package, one filename per line.
If the package or module is not being used/needed/imported, the stanza
will display a single parenthesized note indicating that fact.

For example:

	$ gno mod why gno.land/p/demo/avl gno.land/p/demo/users
	# gno.land/p/demo/avl
	[FILENAME_1.gno]
	[FILENAME_2.gno]

	# gno.land/p/demo/users
	(module [MODULE_NAME] does not need package gno.land/p/demo/users)
	$
`,
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execModWhy(args, io)
		},
	)
}

func execModInit(args []string) error {
	if len(args) > 1 {
		return flag.ErrHelp
	}
	var modPath string
	if len(args) == 1 {
		modPath = args[0]
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := gnomod.CreateGnoModFile(dir, modPath); err != nil {
		return fmt.Errorf("create gno.mod file: %w", err)
	}

	return nil
}

type modTidyCfg struct {
	verbose   bool
	recursive bool
}

func (c *modTidyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"v",
		false,
		"verbose output when running",
	)
	fs.BoolVar(
		&c.recursive,
		"recursive",
		false,
		"walk subdirs for gno.mod files",
	)
}

func execModTidy(cfg *modTidyCfg, args []string, io commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if cfg.recursive {
		pkgs, err := gnomod.ListPkgs(wd)
		if err != nil {
			return err
		}
		var errs error
		for _, pkg := range pkgs {
			err := modTidyOnce(cfg, wd, pkg.Dir, io)
			errs = multierr.Append(errs, err)
		}
		return errs
	}

	// XXX: recursively check parents if no $PWD/gno.mod
	return modTidyOnce(cfg, wd, wd, io)
}

func modTidyOnce(cfg *modTidyCfg, wd, pkgdir string, io commands.IO) error {
	fname := filepath.Join(pkgdir, "gno.mod")
	relpath, err := filepath.Rel(wd, fname)
	if err != nil {
		return err
	}
	if cfg.verbose {
		io.ErrPrintfln("%s", relpath)
	}

	gm, err := gnomod.ParseGnoMod(fname)
	if err != nil {
		return err
	}

	// Drop all existing requires
	for _, r := range gm.Require {
		gm.DropRequire(r.Mod.Path)
	}

	imports, err := getGnoPackageImports(pkgdir)
	if err != nil {
		return err
	}
	for _, im := range imports {
		// skip if importpath is modulepath
		if im == gm.Module.Mod.Path {
			continue
		}
		gm.AddRequire(im, "v0.0.0-latest")
		if cfg.verbose {
			io.ErrPrintfln("   %s", im)
		}
	}

	gm.Write(fname)
	return nil
}

func execModWhy(args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fname := filepath.Join(wd, "gno.mod")
	gm, err := gnomod.ParseGnoMod(fname)
	if err != nil {
		return err
	}

	importToFilesMap, err := getImportToFilesMap(wd)
	if err != nil {
		return err
	}

	// Format and print `gno mod why` output stanzas
	out := formatModWhyStanzas(gm.Module.Mod.Path, args, importToFilesMap)
	io.Printf(out)

	return nil
}

// formatModWhyStanzas returns a formatted output for the go mod why command.
// It takes three parameters:
//   - modulePath (the path of the module)
//   - args (input arguments)
//   - importToFilesMap (a map of import to files).
func formatModWhyStanzas(modulePath string, args []string, importToFilesMap map[string][]string) (out string) {
	for i, arg := range args {
		out += fmt.Sprintf("# %s\n", arg)
		files, ok := importToFilesMap[arg]
		if !ok {
			out += fmt.Sprintf("(module %s does not need package %s)\n", modulePath, arg)
		} else {
			for _, file := range files {
				out += file + "\n"
			}
		}
		if i < len(args)-1 { // Add a newline if it's not the last stanza
			out += "\n"
		}
	}
	return
}

// getImportToFilesMap returns a map where each key is an import path and its
// value is a list of files importing that package with the specified import path.
func getImportToFilesMap(pkgPath string) (map[string][]string, error) {
	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]string) // import -> []file
	for _, e := range entries {
		filename := e.Name()
		if ext := filepath.Ext(filename); ext != ".gno" {
			continue
		}
		if strings.HasSuffix(filename, "_filetest.gno") {
			continue
		}
		imports, err := getGnoFileImports(filepath.Join(pkgPath, filename))
		if err != nil {
			return nil, err
		}

		for _, imp := range imports {
			m[imp] = append(m[imp], filename)
		}
	}
	return m, nil
}

// getGnoPackageImports returns the list of gno imports from a given path.
// Note: It ignores subdirs. Since right now we are still deciding on
// how to handle subdirs.
// See:
// - https://github.com/gnolang/gno/issues/1024
// - https://github.com/gnolang/gno/issues/852
//
// TODO: move this to better location.
func getGnoPackageImports(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	allImports := make([]string, 0)
	seen := make(map[string]struct{})
	for _, e := range entries {
		filename := e.Name()
		if ext := filepath.Ext(filename); ext != ".gno" {
			continue
		}
		if strings.HasSuffix(filename, "_filetest.gno") {
			continue
		}
		imports, err := getGnoFileImports(filepath.Join(path, filename))
		if err != nil {
			return nil, err
		}
		for _, im := range imports {
			if !strings.HasPrefix(im, "gno.land/") {
				continue
			}
			if _, ok := seen[im]; ok {
				continue
			}
			allImports = append(allImports, im)
			seen[im] = struct{}{}
		}
	}
	sort.Strings(allImports)

	return allImports, nil
}

func getGnoFileImports(fname string) ([]string, error) {
	if !strings.HasSuffix(fname, ".gno") {
		return nil, fmt.Errorf("not a gno file: %q", fname)
	}
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, fname, data, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for _, im := range f.Imports {
		importPath := strings.TrimPrefix(strings.TrimSuffix(im.Path.Value, `"`), `"`)
		res = append(res, importPath)
	}
	return res, nil
}
