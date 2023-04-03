package pkg

import "fmt"

// e.g. "github.com/tendermint/tendermint/abci/types" ->
//
//	&Package{...}
type PackageSet map[string]*Package

func NewPackageSet() PackageSet {
	return make(map[string]*Package)
}

func (ps PackageSet) Add(pkg *Package) bool {
	if found, ok := ps[pkg.GoPkgPath]; ok {
		if found != pkg {
			panic(fmt.Errorf("found conflicting package mappkgng, %v -> %v but trying to overwrite with -> %v", pkg.GoPkgPath, found, pkg))
		}
		return false
	} else {
		ps[pkg.GoPkgPath] = pkg
		return true
	}
}

func (ps PackageSet) Get(gopkg string) *Package {
	pkg, ok := ps[gopkg]
	if !ok {
		panic(fmt.Sprintf("package info unrecognized for %v (not registered directly nor indirectly as dependency", gopkg))
	}
	return pkg
}

func (ps PackageSet) Has(gopkg string) bool {
	_, ok := ps[gopkg]
	return ok
}
