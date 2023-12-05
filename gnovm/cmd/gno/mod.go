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
	"github.com/gnolang/gno/tm2/pkg/errors"
)

type modDownloadCfg struct {
	remote  string
	verbose bool
}

func newModCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "mod",
			ShortUsage: "mod <command>",
			ShortHelp:  "Manage gno.mod",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newModDownloadCmd(io),
		newModInitCmd(),
		newModTidy(io),
		newModWhy(),
	)

	return cmd
}

func newModDownloadCmd(io commands.IO) *commands.Command {
	cfg := &modDownloadCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "download",
			ShortUsage: "download [flags]",
			ShortHelp:  "Download modules to local cache",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execModDownload(cfg, args, io)
		},
	)
}

func newModInitCmd() *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [module-path]",
			ShortHelp:  "Initialize gno.mod file in current directory",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execModInit(args)
		},
	)
}

func newModTidy(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "tidy",
			ShortUsage: "tidy",
			ShortHelp:  "Add missing and remove unused modules",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execModTidy(args, io)
		},
	)
}

func newModWhy() *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "why",
			ShortUsage: "why <package> [<package>...]",
			ShortHelp:  "", // TODO
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execModWhy(args)
		},
	)
}

func (c *modDownloadCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"test3.gno.land:36657",
		"remote for fetching gno modules",
	)

	fs.BoolVar(
		&c.verbose,
		"verbose",
		false,
		"verbose output when running",
	)
}

func execModDownload(cfg *modDownloadCfg, args []string, io commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	modPath := filepath.Join(path, "gno.mod")
	if !isFileExist(modPath) {
		return errors.New("gno.mod not found")
	}

	// read gno.mod
	data, err := os.ReadFile(modPath)
	if err != nil {
		return fmt.Errorf("readfile %q: %w", modPath, err)
	}

	// parse gno.mod
	gnoMod, err := gnomod.Parse(modPath, data)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	// sanitize gno.mod
	gnoMod.Sanitize()

	// validate gno.mod
	if err := gnoMod.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	// fetch dependencies
	if err := gnoMod.FetchDeps(gnomod.GetGnoModPath(), cfg.remote, cfg.verbose); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	gomod, err := gnomod.GnoToGoMod(*gnoMod)
	if err != nil {
		return fmt.Errorf("sanitize: %w", err)
	}

	// write go.mod file
	err = gomod.Write(filepath.Join(path, "go.mod"))
	if err != nil {
		return fmt.Errorf("write go.mod file: %w", err)
	}

	return nil
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

func execModTidy(args []string, io commands.IO) error {
	if len(args) > 0 {
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

	// Drop all existing requires
	for _, r := range gm.Require {
		gm.DropRequire(r.Mod.Path)
	}

	imports, err := getGnoModuleImports(wd)
	if err != nil {
		return err
	}
	for _, im := range imports {
		// skip if importpath is modulepath
		if im == gm.Module.Mod.Path {
			continue
		}
		gm.AddRequire(im, "v0.0.0-latest")
	}

	gm.Write(fname)
	return nil
}

func execModWhy(args []string) error {
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
	entries, err := os.ReadDir(wd)
	if err != nil {
		return err
	}

	importsMap := make(map[string][]string)
	for _, e := range entries {
		filename := e.Name()
		if ext := filepath.Ext(filename); ext != ".gno" {
			continue
		}
		if strings.HasSuffix(filename, "_filetest.gno") {
			continue
		}
		imports, err := getGnoFileImports(filepath.Join(wd, filename))
		if err != nil {
			return err
		}

		for _, imp := range imports {
			importsMap[imp] = append(importsMap[imp], filename)
		}
	}

	for i, arg := range args {
		fmt.Println("#", arg)

		files, ok := importsMap[arg]
		if !ok {
			fmt.Println(fmt.Sprintf("(module %s does not need package %s)", gm.Module.Mod.Path, arg))
		} else {
			for _, file := range files {
				fmt.Println(file)
			}
		}

		if i != len(args)-1 {
			fmt.Println()
		}
	}

	return nil
}

// getGnoModuleImports returns the list of gno imports from a given path.
// Note: It ignores subdirs. Since right now we are still deciding on
// how to handle subdirs.
// See:
// - https://github.com/gnolang/gno/issues/1024
// - https://github.com/gnolang/gno/issues/852
//
// TODO: move this to better location.
func getGnoModuleImports(path string) ([]string, error) {
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

	res := make([]string, len(f.Imports))
	for _, im := range f.Imports {
		importPath := strings.TrimPrefix(strings.TrimSuffix(im.Path.Value, `"`), `"`)
		res = append(res, importPath)
	}
	return res, nil
}
