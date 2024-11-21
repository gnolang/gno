package stdlibs

import (
	"embed"

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
	pkgPaths := initOrder
	pkgs := make([]*gnovm.MemPackage, len(pkgPaths))
	for i, pkgPath := range pkgPaths {
		pkgs[i] = gnolang.ReadMemPackageFromFS(embeddedSources, pkgPath, pkgPath)
	}
	return pkgs
}

// EmbeddedMemPackage returns a slice of [gnovm.MemPackage] generated from embedded stdlibs sources
func EmbeddedMemPackage(pkgPath string) *gnovm.MemPackage {
	return gnolang.ReadMemPackageFromFS(embeddedSources, pkgPath, pkgPath)
}
