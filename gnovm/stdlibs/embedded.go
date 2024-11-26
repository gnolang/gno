package stdlibs

import (
	"embed"
	"sync"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// embeddedSources embeds the stdlibs.
// Be careful to remove transpile artifacts before building release binaries or they will be included
//
//go:embed */*
var embeddedSources embed.FS

// EmbeddedSources returns embedded stdlibs sources
func EmbeddedSources() embed.FS {
	return embeddedSources
}

// EmbeddedMemPackages returns a slice of [gnovm.MemPackage] generated from embedded stdlibs sources
func EmbeddedMemPackages() []*gnovm.MemPackage {
	return embeddedMemPackages()
}

// EmbeddedMemPackage returns a slice of [gnovm.MemPackage] generated from embedded stdlibs sources
func EmbeddedMemPackage(pkgPath string) *gnovm.MemPackage {
	for _, pkg := range embeddedMemPackages() {
		if pkg.Path == pkgPath {
			return pkg
		}
	}
	return &gnovm.MemPackage{}
}

var embeddedMemPackages = sync.OnceValue(func() []*gnovm.MemPackage {
	pkgPaths := initOrder
	pkgs := make([]*gnovm.MemPackage, len(pkgPaths))
	for i, pkgPath := range pkgPaths {
		pkgs[i] = gnolang.ReadMemPackageFromFS(embeddedSources, pkgPath, pkgPath)
	}
	return pkgs
})
