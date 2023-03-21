package main

import (
	"context"
	"flag"
	"path/filepath"

	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/commands/doc"
)

type docCfg struct {
	all        bool
	src        bool
	unexported bool
	rootDirStruct
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

	c.rootDirStruct.RegisterFlags(fs)
}

func execDoc(cfg *docCfg, args []string, io *commands.IO) error {
	// guess opts.RootDir
	if cfg.rootDir == "" {
		cfg.rootDir = guessRootDir()
	}
	dirs := doc.NewDirs(filepath.Join(cfg.rootDir, "stdlibs"), filepath.Join(cfg.rootDir, "examples"))
	res, err := doc.ResolveDocumentable(dirs, args, cfg.unexported)
	switch {
	case res == nil:
		return err
	case err != nil:
		io.Printfln("warning: error parsing some candidate packages:\n%v", err)
	}
	output, err := res.Document(
		doc.WithShowAll(cfg.all),
		doc.WithSource(cfg.src),
		doc.WithUnexported(cfg.unexported),
	)
	if err != nil {
		return err
	}
	io.Out.Write([]byte(output))
	return nil
}
