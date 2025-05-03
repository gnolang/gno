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
	test bool
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
	fs.BoolVar(&c.json, "json", false, `The -json flag causes the package data to be printed in JSON format.`)
	// XXX: support template format
	fs.BoolVar(&c.deps, "deps", false, `The -deps flag causes list to iterate over not just the named packages
but also all their dependencies.`)
	// XXX: add DepOnly field and respect golang traversal order
	fs.BoolVar(&c.test, "test", false, `The -test flag causes test dependencies to be loaded as well`)
	// XXX: match golang behavior for test flag, constructing test packages
}

func execList(cfg *listCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	conf := &packages.LoadConfig{
		Out:     io.Err(),
		Fetcher: testPackageFetcher,
		Test:    cfg.test,
		Deps:    cfg.deps,
	}

	pkgs, err := packages.Load(conf, args...)
	if err != nil {
		io.ErrPrintln(err)
		os.Exit(1)
	}

	// try sort
	sorted, err := pkgs.Sort(false)
	if err == nil {
		pkgs = packages.PkgList(sorted)
	}

	if !cfg.json {
		// XXX: support template format
		pkgPaths := make([]string, len(pkgs))
		for i, pkg := range pkgs {
			pkgPaths[i] = pkg.ImportPath
		}
		fmt.Println(strings.Join(pkgPaths, "\n"))
		return nil
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
