package main

import (
	"context"
	"flag"
	"fmt"
	precompile "github.com/gnolang/gno/gnovm/pkg/precompile"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newPrecompileCmd(io commands.IO) *commands.Command {
	cfg := &precompile.PrecompileCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "precompile",
			ShortUsage: "precompile [flags] <package> [<package>...]",
			ShortHelp:  "Precompiles .gno files to .go",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execPrecompile(cfg, args, io)
		},
	)
}

// precompile an existed path, no stdin.
func execPrecompile(cfg *precompile.PrecompileCfg, paths []string, io commands.IO) error {
	fmt.Println("---execPrecompile")
	if len(paths) < 1 {
		return flag.ErrHelp
	}

	err, _ := precompile.PrecompileAndCheckPkg(false, nil, paths, cfg)
	// TODO: io?
	return err
}
