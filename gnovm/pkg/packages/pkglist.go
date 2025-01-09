package packages

import (
	"fmt"
	"slices"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type (
	PkgList       []*Package
	SortedPkgList []*Package
)

// sortPkgs sorts the given packages by their dependencies.
func (pl PkgList) Sort() (SortedPkgList, error) {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	sortedPkgs := make([]*Package, 0, len(pl))

	// Visit all packages
	for _, p := range pl {
		if err := visitPackage(p, pl, visited, onStack, &sortedPkgs); err != nil {
			return nil, err
		}
	}

	return sortedPkgs, nil
}

var injectedTestingLibs = []string{"encoding/json", "fmt", "os", "crypto/sha256"}

// visitNode visits a package's and its dependencies dependencies and adds them to the sorted list.
func visitPackage(pkg *Package, pkgs []*Package, visited, onStack map[string]bool, sortedPkgs *[]*Package) error {
	if onStack[pkg.ImportPath] {
		return fmt.Errorf("cycle detected: %s", pkg.ImportPath)
	}
	if visited[pkg.ImportPath] {
		return nil
	}

	visited[pkg.ImportPath] = true
	onStack[pkg.ImportPath] = true

	// Visit package's dependencies
	for _, imp := range pkg.Imports.Merge(FileKindPackageSource) {
		if gnolang.IsStdlib(imp) {
			continue
		}
		if slices.Contains(injectedTestingLibs, imp) {
			continue
		}
		found := false
		for _, p := range pkgs {
			if p.ImportPath != imp {
				continue
			}
			if err := visitPackage(p, pkgs, visited, onStack, sortedPkgs); err != nil {
				return err
			}
			found = true
			break
		}
		if !found {
			return fmt.Errorf("missing dependency '%s' for package '%s'", imp, pkg.ImportPath)
		}
	}

	onStack[pkg.ImportPath] = false
	*sortedPkgs = append(*sortedPkgs, pkg)
	return nil
}

// GetNonDraftPkgs returns packages that are not draft
// and have no direct or indirect draft dependencies.
func (sp SortedPkgList) GetNonDraftPkgs() SortedPkgList {
	res := make([]*Package, 0, len(sp))
	draft := make(map[string]bool)

	for _, pkg := range sp {
		if pkg.Draft {
			draft[pkg.ImportPath] = true
			continue
		}
		dependsOnDraft := false
		for _, req := range pkg.Imports.Merge(FileKindPackageSource) {
			if draft[req] {
				dependsOnDraft = true
				draft[pkg.ImportPath] = true
				break
			}
		}
		if !dependsOnDraft {
			res = append(res, pkg)
		}
	}
	return res
}

type PackagesMap map[string]*Package

func NewPackagesMap(pkgs ...*Package) PackagesMap {
	pm := make(PackagesMap, len(pkgs))
	pm.AddBulk(pkgs...)
	return pm
}

func (pm PackagesMap) Add(pkg *Package) bool {
	if pkg.ImportPath == "" {
		return false
	}

	if _, ok := pm[pkg.ImportPath]; ok {
		return false
	}

	pm[pkg.ImportPath] = pkg
	return true
}

func (pm PackagesMap) AddBulk(pkgs ...*Package) {
	for _, pkg := range pkgs {
		_ = pm.Add(pkg)
	}
}
