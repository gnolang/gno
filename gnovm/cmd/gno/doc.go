package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type docCfg struct {
	all        bool
	src        bool
	unexported bool
	short      bool
	rootDir    string
}

func newDocCmd(io *commands.IO) *commands.Command {
	c := &docCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "doc",
			ShortUsage: "doc [flags] <pkgsym>",
			ShortHelp:  "get documentation for the specified package or symbol (type, function, method, or variable/constant).",
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
		"clone location of github.com/gnolang/gno (gnodev tries to guess it)",
	)
}

func execDoc(cfg *docCfg, args []string, io *commands.IO) error {
	// guess opts.RootDir
	if cfg.rootDir == "" {
		cfg.rootDir = guessRootDir()
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	rd, err := gnomod.FindRootDir(wd)
	if err != nil && !errors.Is(err, gnomod.ErrGnoModNotFound) {
		return fmt.Errorf("error determining root gno.mod file: %w", err)
	}
	modDirs := []string{rd}

	// select dirs from which to gather directories
	dirs := []string{filepath.Join(cfg.rootDir, "gnovm/stdlibs"), filepath.Join(cfg.rootDir, "examples")}
	res, err := doc.ResolveDocumentable(dirs, modDirs, args, cfg.unexported)
	if res == nil {
		return err
	}
	if err != nil {
		io.Printfln("warning: error parsing some candidate packages:\n%v", err)
	}
	return res.WriteDocumentation(
		io.Out,
		&doc.WriteDocumentationOptions{
			ShowAll:    cfg.all,
			Source:     cfg.src,
			Unexported: cfg.unexported,
			Short:      false,
		},
	)
}
