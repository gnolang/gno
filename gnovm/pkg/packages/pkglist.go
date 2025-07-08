package packages

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type (
	PkgList       []*Package
	SortedPkgList []*Package
)

func (pl PkgList) Get(pkgPath string) *Package {
	for _, p := range pl {
		if p.ImportPath == pkgPath {
			return p
		}
	}
	return nil
}

var ErrPackageNotFound = errors.New("package not found")

func (pl PkgList) GetByDir(dir string) *Package {
	for _, p := range pl {
		if p.Dir == dir {
			return p
		}
	}
	return nil
}

// sortPkgs sorts the given packages by their dependencies.
func (pl PkgList) Sort(ignoreStdlibs bool) (SortedPkgList, error) {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	sortedPkgs := make([]*Package, 0, len(pl))

	// Visit all packages
	for _, p := range pl {
		if err := visitPackage(ignoreStdlibs, p, pl, visited, onStack, &sortedPkgs); err != nil {
			return nil, err
		}
	}

	return sortedPkgs, nil
}

// visitNode visits a package's and its dependencies dependencies and adds them to the sorted list.
func visitPackage(ignoreStdlibs bool, pkg *Package, pkgs []*Package, visited, onStack map[string]bool, sortedPkgs *[]*Package) error {
	if onStack[pkg.ImportPath] {
		return fmt.Errorf("cycle detected: %s", pkg.ImportPath)
	}
	if visited[pkg.ImportPath] {
		return nil
	}

	visited[pkg.ImportPath] = true
	onStack[pkg.ImportPath] = true

	// Visit package's dependencies
	for _, imp := range pkg.Imports[FileKindPackageSource] {
		if gnolang.IsStdlib(imp) && ignoreStdlibs {
			continue
		}

		found := false
		for _, p := range pkgs {
			if p.ImportPath != imp {
				continue
			}
			if err := visitPackage(ignoreStdlibs, p, pkgs, visited, onStack, sortedPkgs); err != nil {
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

// XXX: consider remove this

// GetNonIgnoredPkgs returns packages that are not draft
// and have no direct or indirect draft dependencies.
func (sp SortedPkgList) GetNonIgnoredPkgs() SortedPkgList {
	res := make([]*Package, 0, len(sp))
	ingore := make(map[string]bool)

	for _, pkg := range sp {
		if pkg.Ignore {
			ingore[pkg.ImportPath] = true
			continue
		}
		dependsOnIgnored := false
		for _, req := range pkg.ImportsSpecs.Merge(FileKindPackageSource) {
			if ingore[req.PkgPath] {
				dependsOnIgnored = true
				ingore[pkg.ImportPath] = true
				break
			}
		}
		if !dependsOnIgnored {
			res = append(res, pkg)
		}
	}
	return res
}
