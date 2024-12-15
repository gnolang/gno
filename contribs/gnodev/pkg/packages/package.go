package packages

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

type PackageKind int

const (
	PackageKindOther  = iota
	PackageKindRemote = iota
	PackageKindFS
)

type Package struct {
	gnovm.MemPackage
	Kind     PackageKind
	Location string
}

func ReadPackageFromDir(fset *token.FileSet, path, dir string) (*Package, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir %q: %w", dir, err)
	}

	draft, err := isDraftPackages(dir, files)
	if err != nil {
		return nil, err
	}

	// Skip draft package
	// XXX: We could potentially do that in a middleware, but doing this
	// here avoid to potentially parse broken files
	if draft {
		return nil, ErrResolverPackageSkip
	}

	var name string
	memFiles := []*gnovm.MemFile{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fname := file.Name()
		filepath := filepath.Join(dir, fname)
		body, err := os.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("unable to read file %q: %w", filepath, err)
		}

		if isModFile(fname) {
			file, err := gnomod.Parse(fname, body)
			if err != nil {
				return nil, fmt.Errorf("unable to read `gno.mod`: %w", err)
			}

			// Skip draft package
			if file.Draft {
				return nil, ErrResolverPackageSkip
			}
		}

		if isGnoFile(fname) {
			memfile, pkgname, err := parseFile(fset, fname, body)
			if err != nil {
				return nil, fmt.Errorf("unable to parse file %q: %w", fname, err)
			}

			if !isTestFile(fname) {
				if name != "" && name != pkgname {
					return nil, fmt.Errorf("conflict package name between %q and %q", name, memfile.Name)
				}

				name = pkgname
			}

			memFiles = append(memFiles, memfile)
			continue // continue
		}

		if isValidPackageFile(fname) {
			memFiles = append(memFiles, &gnovm.MemFile{
				Name: fname, Body: string(body),
			})
		}

		// ignore the file
	}

	// Empty package, skipping
	if name == "" {
		return nil, fmt.Errorf("empty package: %w", ErrResolverPackageSkip)
	}

	return &Package{
		MemPackage: gnovm.MemPackage{
			Name:  name,
			Path:  path,
			Files: memFiles,
		},
		Location: dir,
		Kind:     PackageKindFS,
	}, nil
}

func isDraftPackages(dir string, files []fs.DirEntry) (bool, error) {
	for _, file := range files {
		fname := file.Name()
		if !isModFile(fname) {
			continue
		}

		filepath := filepath.Join(dir, fname)
		body, err := os.ReadFile(filepath)
		if err != nil {
			return false, fmt.Errorf("unable to read file %q: %w", filepath, err)
		}

		mod, err := gnomod.Parse(fname, body)
		if err != nil {
			return false, fmt.Errorf("unable to parse `gno.mod`: %w", err)
		}

		return mod.Draft, nil
	}

	return false, nil
}

func parseFile(fset *token.FileSet, fname string, body []byte) (*gnovm.MemFile, string, error) {
	f, err := parser.ParseFile(fset, fname, body, parser.PackageClauseOnly)
	if err != nil {
		return nil, "", fmt.Errorf("unable to parse file %q: %w", fname, err)
	}

	return &gnovm.MemFile{
		Name: fname,
		Body: string(body),
	}, f.Name.Name, nil
}
