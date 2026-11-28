package gnomod

import (
	"fmt"
)

type Pkg struct {
	Dir     string   // absolute path to package dir
	Name    string   // package name
	Imports []string // direct imports of this pkg
	Ignore  bool     // whether the package is a ignored package
}

type (
	PkgList       []Pkg
	SortedPkgList []Pkg
)

// sortPkgs sorts the given packages by their dependencies.
func (pl PkgList) Sort() (SortedPkgList, error) {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	sortedPkgs := make([]Pkg, 0, len(pl))

	// Visit all packages
	for _, p := range pl {
		if err := visitPackage(p, pl, visited, onStack, &sortedPkgs); err != nil {
			return nil, err
		}
	}

	return sortedPkgs, nil
}

// visitNode visits a package's and its dependencies dependencies and adds them to the sorted list.
func visitPackage(pkg Pkg, pkgs []Pkg, visited, onStack map[string]bool, sortedPkgs *[]Pkg) error {
	if onStack[pkg.Name] {
		return fmt.Errorf("cycle detected: %s", pkg.Name)
	}
	if visited[pkg.Name] {
		return nil
	}

	visited[pkg.Name] = true
	onStack[pkg.Name] = true

	// Visit package's dependencies
	for _, imp := range pkg.Imports {
		found := false
		for _, p := range pkgs {
			if p.Name != imp {
				continue
			}
			if err := visitPackage(p, pkgs, visited, onStack, sortedPkgs); err != nil {
				return err
			}
			found = true
			break
		}
		if !found {
			return fmt.Errorf("missing dependency '%s' for package '%s'", imp, pkg.Name)
		}
	}

	onStack[pkg.Name] = false
	*sortedPkgs = append(*sortedPkgs, pkg)
	return nil
}

// GetNonIgnorePkgs returns packages that are not ignored
// and have no direct or indirect ignore dependencies.
func (sp SortedPkgList) GetNonIgnoredPkgs() SortedPkgList {
	res := make([]Pkg, 0, len(sp))
	ignored := make(map[string]bool)

	for _, pkg := range sp {
		if pkg.Ignore {
			ignored[pkg.Name] = true
			continue
		}
		dependsOnIgnored := false
		for _, req := range pkg.Imports {
			if ignored[req] {
				dependsOnIgnored = true
				ignored[pkg.Name] = true
				break
			}
		}
		if !dependsOnIgnored {
			res = append(res, pkg)
		}
	}
	return res
}
