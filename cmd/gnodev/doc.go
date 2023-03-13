package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/pkgs/commands"
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
			LongHelp:   "",
		},
		c,
		func(_ context.Context, args []string) error {
			return execDoc(c, args, io)
		},
	)
}

func (c *docCfg) RegisterFlags(fs *flag.FlagSet) {
	c.rootDirStruct.RegisterFlags(fs)
}

func execDoc(cfg *docCfg, args []string, io *commands.IO) error {
	panic("not implemented")
}
