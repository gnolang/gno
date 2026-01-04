package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// sortPackages sorts pkgs into their initialization order.
func sortPackages(pkgs []*pkgData) []string {
	res := make([]string, 0, len(pkgs))

	var process func(pkg *pkgData, chain []string)
	process = func(pkg *pkgData, chain []string) {
		if idx := slices.Index(chain, pkg.importPath); idx != -1 {
			panic(
				fmt.Errorf("cyclical package initialization on %q (%s -> %s)",
					pkg.importPath,
					strings.Join(chain[idx:], " -> "),
					pkg.importPath,
				),
			)
		}
		// for a deterministic result, sort the imports.
		imports := slices.Sorted(maps.Keys(pkg.imports))
		for _, imp := range imports {
			if slices.Contains(res, imp) {
				continue
			}

			// import does not exist; find it in pkg and process it.
			idx := slices.IndexFunc(pkgs, func(p *pkgData) bool { return p.importPath == imp })
			if idx == -1 {
				panic(fmt.Errorf("package does not exist: %q (while processing imports from %q)", imp, pkg.importPath))
			}
			process(pkgs[idx], append(chain, pkg.importPath))
		}
		res = append(res, pkg.importPath)
	}

	// 16 is a guess of maximum depth of dependency initialization
	ch := make([]string, 0, 16)
	for _, pkg := range pkgs {
		if !slices.Contains(res, pkg.importPath) {
			process(pkg, ch)
		}
	}

	return res
}
