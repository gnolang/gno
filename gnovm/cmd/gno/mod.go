package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/multierr"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var testPackageFetcher pkgdownload.PackageFetcher

func newModCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "mod",
			ShortUsage: "mod <command>",
			ShortHelp:  "module maintenance",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newModDownloadCmd(io),
		newModGraphCmd(io),
		newModInitLegacyCmd(io),
		newModTidy(io),
		newModWhy(io),
	)

	return cmd
}

func newModDownloadCmd(io commands.IO) *commands.Command {
	cfg := &modDownloadCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "download",
			ShortUsage: "download [flags]",
			ShortHelp:  "download modules to local cache",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execModDownload(cfg, args, io)
		},
	)
}

func newModGraphCmd(io commands.IO) *commands.Command {
	cfg := &modGraphCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "graph",
			ShortUsage: "graph [path]",
			ShortHelp:  "print module requirement graph",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execModGraph(cfg, args, io)
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

	$ gno mod why gno.land/p/nt/avl/v0 gno.land/p/demo/users
	# gno.land/p/nt/avl/v0
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

type modDownloadCfg struct {
	remoteOverrides string
}

const remoteOverridesArgName = "remote-overrides"

func (c *modDownloadCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remoteOverrides,
		remoteOverridesArgName,
		"",
		"chain-domain=rpc-url comma-separated list",
	)
}

type modGraphCfg struct {
	format string
}

func (c *modGraphCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.format, "format", "", "Output format, must be one of 'dot' or empty. Empty is a minimalist format.")
}

func execModGraph(cfg *modGraphCfg, args []string, io commands.IO) error {
	if len(args) == 0 {
		args = []string{"."}
	}
	if len(args) > 1 {
		return flag.ErrHelp
	}

	loadConf := packages.LoadConfig{
		Fetcher: testPackageFetcher,
		Deps:    true,
		Test:    true,
		Out:     io.Err(),
	}
	pkgs, err := packages.Load(loadConf, args...)
	if err != nil {
		return err
	}

	sb := &strings.Builder{}

	if cfg.format == "dot" {
		fmt.Fprint(sb, "digraph gno {\n")
	}

	errCount := uint(0)

	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			fmt.Fprintf(io.Err(), "%s: %v", pkg.ImportPath, err)
			errCount++
		}
		deps := pkg.ImportsSpecs.Merge()
		for _, dep := range deps {
			if cfg.format == "dot" {
				fmt.Fprintf(sb, "    %q -> %q;\n", pkg.ImportPath, dep.PkgPath)
			} else {
				fmt.Fprintf(sb, "%s %s\n", pkg.ImportPath, dep.PkgPath)
			}
		}
	}

	if cfg.format == "dot" {
		fmt.Fprint(sb, "}\n")
	}

	io.Out().Write([]byte(sb.String()))

	if errCount != 0 {
		return fmt.Errorf("%d build error(s)", errCount)
	}

	return nil
}

func execModDownload(cfg *modDownloadCfg, args []string, io commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	fetcher := testPackageFetcher
	if fetcher == nil {
		remoteOverrides, err := parseRemoteOverrides(cfg.remoteOverrides)
		if err != nil {
			return fmt.Errorf("invalid %s flag: %w", remoteOverridesArgName, err)
		}
		fetcher = rpcpkgfetcher.New(remoteOverrides)
	} else if len(cfg.remoteOverrides) != 0 {
		return fmt.Errorf("can't use %s flag with a custom package fetcher", remoteOverridesArgName)
	}

	loadCfg := packages.LoadConfig{
		Fetcher:    fetcher,
		Deps:       true,
		Test:       true,
		AllowEmpty: true,
		Out:        io.Err(),
	}
	pkgs, err := packages.Load(loadCfg, "./...")
	if err != nil {
		return err
	}

	errCount := uint(0)
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			fmt.Fprintf(io.Err(), "%s: %v", pkg.ImportPath, err)
			errCount++
		}
	}
	if errCount != 0 {
		return fmt.Errorf("%d build error(s)", errCount)
	}

	return nil
}

func parseRemoteOverrides(arg string) (map[string]string, error) {
	if arg == "" {
		return map[string]string{}, nil
	}
	pairs := strings.Split(arg, ",")
	res := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("expected 2 parts in chain-domain=rpc-url pair %q", arg)
		}
		domain := strings.TrimSpace(parts[0])
		rpcURL := strings.TrimSpace(parts[1])
		res[domain] = rpcURL
	}
	return res, nil
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
		pkgs, err := packages.ReadPkgListFromDir(wd, gno.MPAnyAll)
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

	return modTidyOnce(cfg, wd, wd, io)
}

func modTidyOnce(cfg *modTidyCfg, wd, pkgdir string, io commands.IO) error {
	modExists := false
	for _, fname := range []string{"gnomod.toml", "gno.mod"} {
		fpath := filepath.Join(pkgdir, fname)
		if !isFileExist(fpath) {
			continue
		}

		modExists = true

		relpath, err := filepath.Rel(wd, fpath)
		if err != nil {
			return err
		}
		if cfg.verbose {
			io.ErrPrintfln("%s", relpath)
		}

		gm, err := gnomod.ParseFilepath(fpath)
		if err != nil {
			return err
		}

		if fname == "gno.mod" {
			newPath := filepath.Join(pkgdir, "gnomod.toml")
			gm.WriteFile(newPath)
		} else {
			gm.WriteFile(fpath)
		}
	}

	if !modExists {
		return gnomod.ErrNoModFile
	}

	oldpath := filepath.Join(pkgdir, "gno.mod")
	os.Remove(oldpath)

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
	gm, err := gnomod.ParseDir(wd)
	if err != nil {
		return err
	}

	importToFilesMap, err := getImportToFilesMap(wd)
	if err != nil {
		return err
	}

	out := formatModWhyStanzas(gm.Module, args, importToFilesMap)
	io.Printf(out)

	return nil
}

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
		if i < len(args)-1 {
			out += "\n"
		}
	}
	return
}

func getImportToFilesMap(pkgPath string) (map[string][]string, error) {
	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]string)
	for _, e := range entries {
		filename := e.Name()
		if ext := filepath.Ext(filename); ext != ".gno" {
			continue
		}
		if strings.HasSuffix(filename, "_filetest.gno") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(pkgPath, filename))
		if err != nil {
			return nil, err
		}
		imports, err := packages.FileImports(filename, string(data), nil)
		if err != nil {
			return nil, err
		}

		for _, imp := range imports {
			m[imp.PkgPath] = append(m[imp.PkgPath], filename)
		}
	}
	return m, nil
}
