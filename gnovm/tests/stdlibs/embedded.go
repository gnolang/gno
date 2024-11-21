package stdlibs

import (
	"embed"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// embeddedSources embeds the testing stdlibs.
// Be careful to remove transpile artifacts before building release binaries or they will be included
//
//go:embed */*
var embeddedSources embed.FS

// EmbeddedSources returns embedded testing stdlibs sources
func EmbeddedSources() embed.FS {
	return embeddedSources
}

// EmbeddedMemPackages returns a slice of [gnovm.MemPackage] generated from embedded testing stdlibs sources
func EmbeddedMemPackages() ([]*gnovm.MemPackage, error) {
	pkgPaths := initOrder
	pkgs := make([]*gnovm.MemPackage, len(pkgPaths))
	for i, pkgPath := range pkgPaths {
		pkgs[i] = gnolang.ReadMemPackageFromFS(embeddedSources, pkgPath, pkgPath)
	}
	return pkgs, nil
}
