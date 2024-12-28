package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newPkgAddrCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "pkgaddr",
			ShortUsage: "pkgaddr <pkgpath>",
			ShortHelp:  "`pkgaddr` converts a package path to a package address",
		},
		nil,
		func(_ context.Context, args []string) error {
			return execPkgAddr(args, io)
		},
	)
}

func execPkgAddr(args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	io.Println(gnolang.DerivePkgAddr(args[0]))

	return nil
}
