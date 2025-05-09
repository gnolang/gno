package packages

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

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
	modpath := filepath.Join(dir, "gno.mod")
	if _, err := os.Stat(modpath); err == nil {
		draft, err := isDraftFile(modpath)
		if err != nil {
			return nil, err
		}

		// Skip draft package
		// XXX: We could potentially do that in a middleware, but doing this
		// here avoid to potentially parse broken files
		if draft {
			return nil, ErrResolverPackageSkip
		}
	}

	mempkg, err := gnolang.ReadMemPackage(dir, path)
	switch {
	case err == nil: // ok
	case os.IsNotExist(err):
		return nil, ErrResolverPackageNotFound
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
	if mempkg.IsEmpty() {
		return fmt.Errorf("empty package: %w", ErrResolverPackageSkip)
	}

	// Validate package name
	for _, file := range mempkg.Files {
		if !isGnoFile(file.Name) || isTestFile(file.Name) {
			continue
		}

		f, err := parser.ParseFile(fset, file.Name, file.Body, parser.PackageClauseOnly)
		if err != nil {
			return fmt.Errorf("unable to parse file %q: %w", file.Name, err)
		}

		if f.Name.Name != mempkg.Name {
			return fmt.Errorf("%q package name conflict, expected %q found %q",
				mempkg.Path, mempkg.Name, f.Name.Name)
		}
	}

	return nil
}

func isDraftFile(modpath string) (bool, error) {
	modfile, err := os.ReadFile(modpath)
	if err != nil {
		return false, fmt.Errorf("unable to read file %q: %w", modpath, err)
	}

	mod, err := gnomod.Parse(modpath, modfile)
	if err != nil {
		return false, fmt.Errorf("unable to parse `gno.mod`: %w", err)
	}

	return mod.Draft, nil
}
