package packages

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type PackageKind int

const (
	PackageKindOther  = iota
	PackageKindRemote = iota
	PackageKindFS
)

type Package struct {
	std.MemPackage
	Kind     PackageKind
	Location string
}

func ReadPackageFromDir(fset *token.FileSet, path, dir string) (*Package, error) {
	if !gnolang.IsUserlib(path) && !gnolang.IsStdlib(path) {
		return nil, ErrResolverPackageSkip
	}

	mod, err := gnomod.ParseDir(dir)
	switch {
	case err == nil:
		if mod.Ignore {
			// Skip ignored package
			// XXX: We could potentially do that in a middleware, but doing this
			// here avoid to potentially parse broken files
			return nil, ErrResolverPackageSkip
		}
	case errors.Is(err, os.ErrNotExist) || errors.Is(err, gnomod.ErrNoModFile):
		// gnomod.toml is not present, continue anyway
	default:
		return nil, err
	}

	mempkg, err := gnolang.ReadMemPackage(dir, path, gnolang.MPAnyAll)
	switch {
	case err == nil: // ok
	case os.IsNotExist(err):
		return nil, ErrResolverPackageNotFound
	case mempkg == nil || mempkg.IsEmpty(): // XXX: should check an internal error instead
		return nil, ErrResolverPackageSkip
	default:
		return nil, fmt.Errorf("unable to read package %q: %w", dir, err)
	}

	if err := validateMemPackage(fset, mempkg); err != nil {
		return nil, err
	}

	return &Package{
		MemPackage: *mempkg,
		Location:   dir,
		Kind:       PackageKindFS,
	}, nil
}

func validateMemPackage(fset *token.FileSet, mempkg *std.MemPackage) error {
	if isMemPackageEmpty(mempkg) {
		return fmt.Errorf("empty package: %w", ErrResolverPackageSkip)
	}

	// Validate package name
	for _, file := range mempkg.Files {
		if !isGnoFile(file.Name) {
			continue
		}

		f, err := parser.ParseFile(fset, file.Name, file.Body, parser.PackageClauseOnly)
		if err != nil {
			return fmt.Errorf("unable to parse file %q: %w", file.Name, err)
		}

		if strings.HasSuffix(file.Name, "_filetest.gno") {
			continue
		}

		pname := strings.TrimSuffix(f.Name.Name, "_test")
		if pname != mempkg.Name {
			return fmt.Errorf("%q package name conflict, expected %q found %q",
				mempkg.Path, mempkg.Name, f.Name.Name)
		}
	}

	return nil
}

func isMemPackageEmpty(mempkg *std.MemPackage) bool {
	if mempkg.IsEmpty() {
		return true
	}

	for _, file := range mempkg.Files {
		if isGnoFile(file.Name) || file.Name == "gnomod.toml" {
			return false
		}
	}

	return true
}
