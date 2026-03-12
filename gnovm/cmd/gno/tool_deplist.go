package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type deplistCfg struct {
	json    bool
	testDep bool
}

func newDeplistCmd(io commands.IO) *commands.Command {
	cfg := &deplistCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "deplist",
			ShortUsage: "gno tool deplist [flags] <package> [<package>...]",
			ShortHelp:  "list dependencies in topological order",
			LongHelp:   "Deplist resolves transitive dependencies for the given packages and prints them in topological order (dependencies first).",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDeplist(cfg, args, io)
		},
	)
}

func (c *deplistCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.json, "json", false, "output in JSON format")
	fs.BoolVar(&c.testDep, "test-dep", false, "include test dependencies")
}

func execDeplist(cfg *deplistCfg, args []string, io commands.IO) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	loadCfg := packages.LoadConfig{
		Fetcher: testPackageFetcher,
		Deps:    true,
		Test:    cfg.testDep,
		Out:     io.Err(),
	}
	pkgs, err := packages.Load(loadCfg, args...)
	if err != nil {
		return err
	}

	// Filter out stdlibs — they're built into the VM and not deployed.
	var userPkgs packages.PkgList
	for _, pkg := range pkgs {
		if gnolang.IsStdlib(pkg.ImportPath) {
			continue
		}
		userPkgs = append(userPkgs, pkg)
	}

	// Topological sort by source imports only. Test deps may form cycles
	// (e.g. avl_test → uassert → avl) which is fine — they expand the
	// package set but don't affect deployment order.
	sorted, err := sortSkipMissing(userPkgs)
	if err != nil {
		return err
	}

	if cfg.json {
		lw := newJsonListWriter(io.Out())
		for _, pkg := range sorted {
			if pkg.Ignore {
				continue
			}
			if err := lw.write(pkg); err != nil {
				return err
			}
		}
		return nil
	}

	for _, pkg := range sorted {
		if pkg.Ignore {
			continue
		}
		fmt.Fprintln(io.Out(), pkg.Dir)
	}
	return nil
}

// sortSkipMissing is a topological sort that silently skips imports
// not present in the package list (e.g. stdlibs filtered out earlier).
func sortSkipMissing(pkgs packages.PkgList) ([]*packages.Package, error) {
	byPath := make(map[string]*packages.Package, len(pkgs))
	for _, p := range pkgs {
		byPath[p.ImportPath] = p
	}

	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	var sorted []*packages.Package

	var visit func(pkg *packages.Package) error
	visit = func(pkg *packages.Package) error {
		if onStack[pkg.ImportPath] {
			return fmt.Errorf("cycle detected: %s", pkg.ImportPath)
		}
		if visited[pkg.ImportPath] {
			return nil
		}
		visited[pkg.ImportPath] = true
		onStack[pkg.ImportPath] = true

		for _, imp := range pkg.Imports[packages.FileKindPackageSource] {
			dep, ok := byPath[imp]
			if !ok {
				continue // stdlib or other non-user package
			}
			if err := visit(dep); err != nil {
				return err
			}
		}

		onStack[pkg.ImportPath] = false
		sorted = append(sorted, pkg)
		return nil
	}

	for _, p := range pkgs {
		if err := visit(p); err != nil {
			return nil, err
		}
	}
	return sorted, nil
}
