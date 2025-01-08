package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type listCfg struct {
	json bool
	deps bool
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
	fs.BoolVar(&c.json, "json", false, `The -json flag causes the package data to be printed in JSON format
instead of using the template format. The JSON flag can optionally be
provided with a set of comma-separated required field names to be output.
If so, those required fields will always appear in JSON output, but
others may be omitted to save work in computing the JSON struct.`)
	fs.BoolVar(&c.deps, "deps", false, `The -deps flag causes list to iterate over not just the named packages
but also all their dependencies. It visits them in a depth-first post-order
traversal, so that a package is listed only after all its dependencies.
Packages not explicitly listed on the command line will have the DepOnly
field set to true`)
}

func execList(cfg *listCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	conf := &packages.LoadConfig{IO: io, Fetcher: testPackageFetcher}

	if cfg.deps {
		conf.Deps = true
	}

	if !cfg.json {
		pkgs, err := packages.Load(conf, args...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pkgPaths := make([]string, len(pkgs))
		for i, pkg := range pkgs {
			pkgPaths[i] = pkg.ImportPath
		}
		fmt.Println(strings.Join(pkgPaths, "\n"))
		return nil
	}

	pkgs, err := packages.Load(conf, args...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, pkg := range pkgs {
		pkgBytes, err := json.MarshalIndent(pkg, "", "\t")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(pkgBytes))
	}

	return nil
}
