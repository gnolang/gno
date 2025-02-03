package packages

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/davecgh/go-spew/spew"
	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type (
	PkgList       []*Package
	SortedPkgList []*Package
)

var _ gnolang.MemPackageGetter = (*PkgList)(nil)

func (pl PkgList) Get(pkgPath string) *Package {
	for _, p := range pl {
		if p.ImportPath == pkgPath {
			return p
		}
	}
	return nil
}

func (pl PkgList) PkgPaths() []string {
	res := make([]string, 0, len(pl))
	for _, pkg := range pl {
		res = append(res, pkg.ImportPath)
	}
	return res
}

func (pl PkgList) Traverse(roots []string, filekinds []FileKind, cb func(p *Package) error) error {
	visited := []string{}

	for _, root := range roots {
		if err := pl.traverseVisit(&visited, root, filekinds, cb); err != nil {
			return err
		}
	}

	return nil
}

func (pl PkgList) traverseVisit(visited *[]string, pkgPath string, filekinds []FileKind, cb func(p *Package) error) error {
	if slices.Contains(*visited, pkgPath) {
		return nil
	}
	*visited = append(*visited, pkgPath)

	pkg := pl.Get(pkgPath)
	if pkg == nil {
		return fmt.Errorf("%s: %w", pkgPath, ErrPackageNotFound)
	}

	if err := cb(pkg); err != nil {
		return err
	}

	for _, imp := range pkg.ImportsSpecs.Merge(filekinds...) {
		if err := pl.traverseVisit(visited, imp.PkgPath, filekinds, cb); err != nil {
			return err
		}
	}

	return nil
}

func (pl PkgList) Explore(roots []string, cb func(p *Package) ([]string, error)) error {
	visited := []string{}

	for _, root := range roots {
		if err := pl.exploreVisit(&visited, root, cb); err != nil {
			return err
		}
	}

	return nil
}

func (pl PkgList) exploreVisit(visited *[]string, pkgPath string, cb func(p *Package) ([]string, error)) error {
	if slices.Contains(*visited, pkgPath) {
		return nil
	}
	*visited = append(*visited, pkgPath)

	pkg := pl.Get(pkgPath)
	if pkg == nil {
		return fmt.Errorf("%s: %w", pkgPath, ErrPackageNotFound)
	}

	next, err := cb(pkg)
	if err != nil {
		return err
	}

	for _, imp := range next {
		if err := pl.exploreVisit(visited, imp, cb); err != nil {
			return err
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

func (pl PkgList) Matches() PkgList {
	res := PkgList{}
	for _, p := range pl {
		if len(p.Match) == 0 {
			continue
		}
		res = append(res, p)
	}
	return res
}

func (pl PkgList) GetMemPackage(pkgPath string) *gnovm.MemPackage {
	pkg := pl.Get(pkgPath)
	if pkg == nil {
		return nil
	}
	memPkg, err := pkg.MemPkg()
	if err != nil {
		spew.Fdump(os.Stderr, "get err", err)
		panic(err)
	}
	return memPkg
}

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
	for _, imp := range pkg.Imports[FileKindPackageSource] {
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
		for _, req := range pkg.ImportsSpecs.Merge(FileKindPackageSource) {
			if draft[req.PkgPath] {
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
