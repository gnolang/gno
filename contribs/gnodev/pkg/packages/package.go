package packages

import (
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Kind classifies a package by where it lives.
// KindUnknown is the zero value: a package was constructed without an
// explicit Kind. FS packages are in a user workspace or extra root;
// Remote packages are fetched from a chain (RPC or modcache) and aren't
// user-editable.
type Kind int

const (
	KindUnknown Kind = iota
	KindFS
	KindRemote
)

// Package is the simplified package type used by the native Loader.
type Package struct {
	ImportPath string
	Dir        string
	Kind       Kind
	Name       string

	memPkg *std.MemPackage // set only for in-memory-backed test packages
}

// ToMemPackage reads the package content. In-memory-backed packages return
// the embedded MemPackage directly. Filesystem-backed packages are read
// lazily on first call.
func (p *Package) ToMemPackage() (*std.MemPackage, error) {
	if p.memPkg != nil {
		return p.memPkg, nil
	}
	if p.Dir == "" {
		return nil, fmt.Errorf("package %s has no directory", p.ImportPath)
	}

	// Use MPUserProd / MPStdlibProd — the deployed package doesn't ship test
	// files. gnodev is a dev-time tool so skipping tests is fine; including
	// them (MPUserAll) triggers chain-side type-checks that fail when a
	// test file imports a package whose own tests haven't been deployed yet.
	mptype := gnolang.MPUserProd
	if gnolang.IsStdlib(p.ImportPath) {
		mptype = gnolang.MPStdlibProd
	}
	mp, err := gnolang.ReadMemPackage(p.Dir, p.ImportPath, mptype)
	if err != nil {
		return nil, fmt.Errorf("read package %s at %s: %w", p.ImportPath, p.Dir, err)
	}
	return mp, nil
}

func packageFromMemPackage(mp *std.MemPackage) *Package {
	return &Package{
		ImportPath: mp.Path,
		Name:       mp.Name,
		Kind:       KindFS, // irrelevant for in-memory; classification happens at resolve time
		memPkg:     mp,
	}
}
