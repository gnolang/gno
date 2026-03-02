package gnofmt

import (
	"fmt"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

type Resolver interface {
	// ResolveName should resolve the given package name by returning a list
	// of packages matching the given name
	ResolveName(pkgname string) []Package
	// ResolvePath should resolve the given package path by returning a
	// single package
	ResolvePath(pkgpath string) Package
}

type FSResolver struct {
	fset    *token.FileSet
	visited map[string]bool
	pkgpath map[string]Package   // pkg path -> pkg
	pkgs    map[string][]Package // pkg name -> []pkg
}

func NewFSResolver() *FSResolver {
	return &FSResolver{
		fset:    token.NewFileSet(),
		visited: map[string]bool{},
		pkgpath: map[string]Package{},
		pkgs:    map[string][]Package{},
	}
}

func (r *FSResolver) ResolveName(pkgname string) []Package {
	// First stdlibs, then external packages
	return r.pkgs[pkgname]
}

func (r *FSResolver) ResolvePath(pkgpath string) Package {
	return r.pkgpath[pkgpath]
}

// PackageHandler is a callback passed to the resolver during package loading.
// PackageHandler will be called on each package. If no error is passed, that
// means that the package has been fully loaded.
// If any handled error is returned from the handler, the package process will
// immediately stop.
type PackageHandler func(path string, err error) error

func basicPkgHandler(path string, err error) error {
	return err
}

// LoadPackages lists all packages in the directory (excluding those which can't be processed).
func (r *FSResolver) LoadPackages(root string, pkgHandler PackageHandler) error {
	if pkgHandler == nil {
		pkgHandler = basicPkgHandler
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // skip error
		}

		if !d.IsDir() {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		if d.Name() == "filetests" {
			// This isn't a package dir and should be ignored by ParsePackage, but skip just to be sure
			return filepath.SkipDir
		}

		// Skip already visited dir
		if r.visited[path] {
			return filepath.SkipDir
		}
		r.visited[path] = true

		pkg, err := ParsePackage(r.fset, root, path)
		if err != nil {
			return pkgHandler(
				path,
				fmt.Errorf("unable to inspect package %q: %w", path, err),
			)
		}

		if pkg == nil || pkg.Path() == "" {
			// not a package
			return nil
		}

		// Check for conflict with previous import path
		if _, ok := r.pkgpath[pkg.Path()]; ok {
			// Stop on path conflict, has a package path should be uniq
			return pkgHandler(
				path,
				fmt.Errorf("%q has been declared twice", pkg.Path()),
			)
		}

		r.pkgpath[pkg.Path()] = pkg
		r.pkgs[pkg.Name()] = append(r.pkgs[pkg.Name()], pkg)

		return pkgHandler(path, nil)
	})

	return err
}
