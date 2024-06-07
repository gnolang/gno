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
	ResolveName(pkgname string) []*Package
	ResolvePath(pkgpath string) *Package
}

type FSResolver struct {
	root, prefix string
	pkgpath      map[string]*Package   // pkg path-> pkg
	stdlibs      map[string][]*Package // pkg name -> []pkg
	extlibs      map[string][]*Package // pkg name -> []pkg
}

func NewFSResolver(root, prefix string) *FSResolver {
	return &FSResolver{
		root: root, prefix: prefix,
		pkgpath: map[string]*Package{},
		stdlibs: map[string][]*Package{},
		extlibs: map[string][]*Package{},
	}
}

func (p *FSResolver) ResolveName(pkgname string) []*Package {
	// first stdlibs, then external packages
	return append(p.stdlibs[pkgname], p.extlibs[pkgname]...)
}

func (p *FSResolver) ResolvePath(pkgpath string) *Package {
	return p.pkgpath[pkgpath]
}

func (p *FSResolver) LoadStdPackages(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		files, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		var gnofiles []string
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".gno" {
				gnofiles = append(gnofiles, filepath.Join(path, file.Name()))
			}
		}
		if len(gnofiles) == 0 {
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

		if oldPkg, ok := p.extlibs[memPkg.Path]; ok {
			return fmt.Errorf("conflict between %q and %q", oldPkg[0].Dir, newPkg.Dir)
		}

		p.pkgpath[memPkg.Path] = newPkg
		p.stdlibs[memPkg.Name] = append(p.stdlibs[memPkg.Name], newPkg)
		return nil
	})
}

func (p *FSResolver) LoadPackages(root string) error {
	pkgs, err := ListPkgs(root)
	if err != nil {
		return fmt.Errorf("unable to resolve example folder: %w", err)
	}

	for _, pkg := range pkgs {
		if oldPkg, ok := p.extlibs[pkg.Path]; ok {
			return fmt.Errorf("conflict between %q and %q", oldPkg[0].Dir, pkg.Dir)
		}

		p.pkgpath[pkg.Path] = pkg
		p.extlibs[pkg.Name] = append(p.extlibs[pkg.Name], pkg)
	}

	return nil
}

func ListPkgs(root string) ([]*Package, error) {
	var pkgs []*Package
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() || root == path {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		pkg, err := inspectPackage(root, path, fset)
		if err != nil {
			if debug {
				fmt.Fprintf(os.Stderr, "unable to inspect package %q: %s\n", path, err)
			}

			return nil // Skip on error
		}

		if pkg != nil {
			pkgs = append(pkgs, pkg)
		}

		return nil
	})

	return pkgs, err
}

func isValidGnoFile(name string) bool {
	return filepath.Ext(name) == ".gno" &&
		// Ignore testfile
		!strings.HasSuffix(name, "_filetest.gno") &&
		!strings.HasSuffix(name, "_test.gno") &&
		// Ignore dotfile
		!strings.HasPrefix(name, ".")
}

func inspectPackage(root string, path string, fset *token.FileSet) (*Package, error) {
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
		f, err := parser.ParseFile(fset, filename, nil, parser.PackageClauseOnly)
		if err != nil {
			return nil, fmt.Errorf("unable to parse file %q: %w", filename, err)
		}

		if pkgname != "" && pkgname != f.Name.Name {
			return nil, fmt.Errorf("invalid package name in %q", filename)
		}

		pkgname = f.Name.Name

	}

	if pkgname == "" {
		return nil, nil // not a package
	}

	var pkgpath string

	gnoModPath := filepath.Join(path, "gno.mod")
	data, err := os.ReadFile(gnoModPath)
	switch {
	case os.IsNotExist(err):
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
