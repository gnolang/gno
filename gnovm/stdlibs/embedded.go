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
	return embeddedMemPackagesMap()[pkgPath]
}

var embeddedMemPackages = sync.OnceValue(func() []*gnovm.MemPackage {
	pkgsMap := embeddedMemPackagesMap()
	pkgs := make([]*gnovm.MemPackage, len(pkgsMap))
	for i, pkgPath := range initOrder {
		pkgs[i] = pkgsMap[pkgPath]
	}
	return pkgs
})

var embeddedMemPackagesMap = sync.OnceValue(func() map[string]*gnovm.MemPackage {
	pkgPaths := initOrder
	pkgs := make(map[string]*gnovm.MemPackage, len(pkgPaths))
	for _, pkgPath := range pkgPaths {
		pkgs[pkgPath] = gnolang.ReadMemPackageFromFS(embeddedSources, pkgPath, pkgPath)
	}
	return pkgs
})
