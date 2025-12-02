package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type docCfg struct {
	all           bool
	src           bool
	unexported    bool
	short         bool
	rootDir       string
	remote        string
	remoteTimeout time.Duration
}

func newDocCmd(io commands.IO) *commands.Command {
	c := &docCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "doc",
			ShortUsage: "doc [flags] <pkgsym>",
			ShortHelp:  "show documentation for package or symbol",
			LongHelp:   "get documentation for the specified package or symbol (type, function, method, or variable/constant)",
		},
		c,
		func(_ context.Context, args []string) error {
			return execDoc(c, args, io)
		},
	)
}

func (c *docCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.all,
		"all",
		false,
		"show documentation for all symbols in package",
	)

	fs.BoolVar(
		&c.src,
		"src",
		false,
		"show source code for symbols",
	)

	fs.BoolVar(
		&c.unexported,
		"u",
		false,
		"show unexported symbols as well as exported",
	)

	fs.BoolVar(
		&c.short,
		"short",
		false,
		"show a one line representation for each symbol",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gno binary tries to guess it)",
	)

	fs.StringVar(
		&c.remote,
		"remote",
		"https://rpc.gno.land:443",
		"remote gno.land node address (if needed)",
	)

	fs.DurationVar(
		&c.remoteTimeout,
		"remote-timeout",
		time.Minute,
		"defined how much time a request to the node should live before timeout",
	)
}

func execDoc(cfg *docCfg, args []string, io commands.IO) error {
	// guess opts.RootDir
	if cfg.rootDir == "" {
		cfg.rootDir = gnoenv.RootDir()
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	var modDirs []string

	if gnomod.IsGnomodRoot(wd) {
		modDirs = append(modDirs, wd)
	}

	examplesModules, err := findGnomodExamples(filepath.Join(cfg.rootDir, "examples"))
	if err != nil {
		io.Printfln("warning: error scanning examples directory for modules: %v", err)
	} else {
		modDirs = append(modDirs, examplesModules...)
	}

	// select dirs from which to gather directories
	dirs := []string{filepath.Join(cfg.rootDir, "gnovm/stdlibs")}
	res, err := doc.ResolveDocumentable(dirs, modDirs, args, cfg.unexported, cfg.remote, cfg.remoteTimeout)
	if res == nil {
		return err
	}
	if err != nil {
		io.Printfln("warning: error parsing some candidate packages:\n%v", err)
	}
	return res.WriteDocumentation(
		io.Out(),
		&doc.WriteDocumentationOptions{
			ShowAll:    cfg.all,
			Source:     cfg.src,
			Unexported: cfg.unexported,
			Short:      false,
		},
	)
}

func findGnomodExamples(dir string) ([]string, error) {
	dirs := make([]string, 0, 64) // "hint" about the size
	err := filepath.WalkDir(dir, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !e.IsDir() && e.Name() == "gnomod.toml" {
			dirs = append(dirs, filepath.Dir(path))
		}
		return nil
	})
	return dirs, err
}
