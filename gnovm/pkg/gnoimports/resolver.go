package gnoimports

import (
	"fmt"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var debug bool

func init() {
	debug, _ = strconv.ParseBool(os.Getenv("GNOFMT_DEBUG"))
}

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

// LoadPackages lists all packages in the directory (excluding those which can't be processed).
func (r *FSResolver) LoadPackages(root string) error {
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

		// Skip already visited dir
		if r.visited[path] {
			return filepath.SkipDir
		}
		r.visited[path] = true

		pkg, err := ParsePackage(r.fset, root, path)
		if err != nil {
			return r.newStrictError("unable to inspect package %q: %w", path, err)
		}

		if pkg == nil || pkg.Path() == "" {
			// not a package
			return nil
		}

		// Check for conflict with previous import path
		if _, ok := r.pkgpath[pkg.Path()]; ok {
			// Stop on path conflict, has a package path should be uniq
			return r.newStrictError("%q has been declared twice\n", pkg.Path())
		}

		r.pkgpath[pkg.Path()] = pkg
		r.pkgs[pkg.Name()] = append(r.pkgs[pkg.Name()], pkg)

		if debug {
			fmt.Printf("added %s:%s\n", pkg.Path(), pkg.Name())
		}
		return nil
	})

	return err
}

// XXX: Use error handler
func (r *FSResolver) newStrictError(f string, args ...any) error {
	err := fmt.Errorf(f, args...)
	if debug {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	return nil
}
