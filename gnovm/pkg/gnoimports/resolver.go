package gnoimports

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

var debug bool

func init() {
	debug, _ = strconv.ParseBool(os.Getenv("GNOFMT_DEBUG"))
}

type Resolver interface {
	// ResolveName should resolve the given package name by returning a list
	// of packages matching the given name
	ResolveName(pkgname string) []*Package
	// ResolvePath should resolve the given package path by returning a
	// single package
	ResolvePath(pkgpath string) *Package
}

type FSResolver struct {
	// When strict is enable resolve will fail on any error
	strict bool

	fset    *token.FileSet
	visited map[string]bool
	pkgpath map[string]*Package   // pkg path -> pkg
	stdlibs map[string][]*Package // pkg name -> []pkg
	extlibs map[string][]*Package // pkg name -> []pkg
}

func NewFSResolver(strict bool) *FSResolver {
	return &FSResolver{
		strict:  strict,
		fset:    token.NewFileSet(),
		visited: map[string]bool{},
		pkgpath: map[string]*Package{},
		stdlibs: map[string][]*Package{},
		extlibs: map[string][]*Package{},
	}
}

func (p *FSResolver) ResolveName(pkgname string) []*Package {
	// First stdlibs, then external packages
	return append(p.stdlibs[pkgname], p.extlibs[pkgname]...)
}

func (p *FSResolver) ResolvePath(pkgpath string) *Package {
	return p.pkgpath[pkgpath]
}

// LoadStdPackages loads all standard packages from the root directory.
// Std packages are not prefixed by the root directory.
func (r *FSResolver) LoadStdPackages(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if debug {
				fmt.Fprintf(os.Stderr, err.Error())
			}

			return err
		}

		if !d.IsDir() {
			return nil
		}

		// Skip already visited dir
		if r.visited[path] {
			return filepath.SkipDir
		}
		r.visited[path] = true

		files, err := os.ReadDir(path)
		if err != nil {
			return r.newStrictError("unable to read directory %q: %w", path, err)
		}

		var gnofiles []string
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".gno" {
				gnofiles = append(gnofiles, filepath.Join(path, file.Name()))
			}
		}
		if len(gnofiles) == 0 {
			// Skip as directory does not contain any gno files
			return nil
		}

		pkgname, ok := strings.CutPrefix(path, root)
		if !ok {
			return nil
		}

		memPkg := gnolang.ReadMemPackageFromList(gnofiles, strings.TrimPrefix(pkgname, "/"))
		newPkg := &Package{
			Name: memPkg.Name,
			Path: memPkg.Path,
			Dir:  path,
		}

		// Check for conflict with previous import path
		if oldPkg, ok := r.pkgpath[memPkg.Path]; ok {
			// Stop on path conflict, has a package path should be uniq
			return r.newStrictError("conflict between %q and %q", oldPkg.Dir, newPkg.Dir)
		}

		r.pkgpath[memPkg.Path] = newPkg
		r.stdlibs[memPkg.Name] = append(r.stdlibs[memPkg.Name], newPkg)
		return nil
	})
}

// listAllPkgsFromRoot lists all packages in the directory (excluding those which can't be processed).
func (r *FSResolver) LoadPackages(root string) error {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // skip error
		}

		if !d.IsDir() {
			return nil
		}

		// Skip already visited dir
		if r.visited[path] {
			return filepath.SkipDir
		}
		r.visited[path] = true

		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		pkg, err := r.parsePackage(root, path)
		if err != nil {
			return r.newStrictError("unable to inspect package %q: %w", path, err)
		}

		if pkg == nil {
			// not a package
			return nil
		}

		// Check for conflict with previous import path
		if oldPkg, ok := r.pkgpath[pkg.Path]; ok {
			// Stop on path conflict, has a package path should be uniq
			return r.newStrictError("conflict between %q and %q", oldPkg.Dir, pkg.Dir)
		}

		r.pkgpath[pkg.Path] = pkg
		r.extlibs[pkg.Name] = append(r.extlibs[pkg.Name], pkg)
		return nil
	})

	return err
}

func (r *FSResolver) parsePackage(root string, path string) (*Package, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir %q: %w", path, err)
	}

	var pkgname string
	for _, file := range files {
		name := file.Name()
		if !isValidGnoFile(name) {
			continue
		}

		filename := filepath.Join(path, name)
		f, err := parser.ParseFile(r.fset, filename, nil, parser.PackageClauseOnly)
		if err != nil {
			return nil, fmt.Errorf("unable to parse file %q: %w", filename, err)
		}

		if pkgname != "" && pkgname != f.Name.Name {
			return nil, fmt.Errorf("invalid package name in %q", filename)
		}

		pkgname = f.Name.Name

	}

	if pkgname == "" {
		return nil, nil // Not a package
	}

	var pkgpath string

	// Check for a gno.mod, in which case it will define the module path
	gnoModPath := filepath.Join(path, "gno.mod")
	data, err := os.ReadFile(gnoModPath)
	switch {
	case os.IsNotExist(err):
		// Fallback on dir path
		pkgpath = strings.TrimPrefix(path, root+"/")
	case err == nil:
		gnoMod, err := gnomod.Parse(gnoModPath, data)
		if err != nil {
			return nil, fmt.Errorf("unable to parse gnomod %q: %w", gnoModPath, err)
		}

		gnoMod.Sanitize()
		if err := gnoMod.Validate(); err != nil {
			return nil, fmt.Errorf("unable to validate gnomod %q: %w", gnoModPath, err)
		}

		pkgpath = gnoMod.Module.Mod.Path
	default:
		return nil, fmt.Errorf("unable to read %q: %w", gnoModPath, err)
	}

	return &Package{
		Path: pkgpath,
		Name: pkgname,
		Dir:  path,
	}, nil
}

func (r *FSResolver) newStrictError(f string, args ...any) error {
	err := fmt.Errorf(f, args...)
	if r.strict {
		return err
	}

	if debug {
		fmt.Fprintf(os.Stderr, err.Error())
	}

	return nil
}

func isValidGnoFile(name string) bool {
	return filepath.Ext(name) == ".gno" &&
		// Ignore testfile
		!strings.HasSuffix(name, "_filetest.gno") &&
		!strings.HasSuffix(name, "_test.gno") &&
		// Ignore dotfile
		!strings.HasPrefix(name, ".")
}
