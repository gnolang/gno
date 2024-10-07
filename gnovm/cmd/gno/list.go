package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/importer"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type listCfg struct {
	json bool
}

func newListCmd(io commands.IO) *commands.Command {
	cfg := &listCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "list [flags] <package> [<package>...]",
			ShortHelp:  "runs the lister for the specified packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execList(cfg, args, io)
		},
	)
}

func (c *listCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.json, "json", false, "verbose output when listning")
}

func execList(cfg *listCfg, args []string, _ commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	if !cfg.json {
		pkgPaths, err := importer.ListPackages(args...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(strings.Join(pkgPaths, "\n"))
		return nil
	}

	pkgs, err := importer.ResolvePackages(args...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	pkgsBytes, err := json.MarshalIndent(pkgs, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(pkgsBytes))

	return nil
}
