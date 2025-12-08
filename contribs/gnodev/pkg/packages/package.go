package packages

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Common errors
var (
	ErrResolverPackageNotFound = errors.New("package not found")
	ErrResolverPackageSkip     = errors.New("package should be skipped")
)

type PackageKind int

const (
	PackageKindOther PackageKind = iota
	PackageKindRemote
	PackageKindFS
)

// Package represents a Gno package with its location information.
// Unlike the old design, MemPackage content is loaded on-demand via ToMemPackage().
type Package struct {
	ImportPath string      // Import path (e.g., gno.land/r/demo/boards)
	Dir        string      // Filesystem directory
	Kind       PackageKind // FS or Remote
	Name       string      // Package name

	// memPkg is used for mock/test packages that don't have a filesystem directory.
	// When set, ToMemPackage returns this directly instead of reading from disk.
	memPkg *std.MemPackage
}

// ToMemPackage reads the full package content from disk,
// or returns the mock package if set.
func (p *Package) ToMemPackage() (*std.MemPackage, error) {
	// If we have a mock package, return it directly
	if p.memPkg != nil {
		return p.memPkg, nil
	}

	if p.Dir == "" {
		return nil, fmt.Errorf("package %s has no directory", p.ImportPath)
	}

	mptype := gnolang.MPUserAll
	if gnolang.IsStdlib(p.ImportPath) {
		mptype = gnolang.MPStdlibAll
	}

	mempkg, err := gnolang.ReadMemPackage(p.Dir, p.ImportPath, mptype)
	if err != nil {
		return nil, fmt.Errorf("failed to read package %s from %s: %w", p.ImportPath, p.Dir, err)
	}

	return mempkg, nil
}

// NewPackageFromMemPackage creates a Package from an in-memory MemPackage.
// This is primarily used for testing.
func NewPackageFromMemPackage(mempkg *std.MemPackage) *Package {
	return &Package{
		ImportPath: mempkg.Path,
		Name:       mempkg.Name,
		Dir:        "",
		Kind:       PackageKindOther,
		memPkg:     mempkg,
	}
}
