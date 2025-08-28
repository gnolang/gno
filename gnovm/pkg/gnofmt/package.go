package gnofmt

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

type Package interface {
	// Should return the package path
	Path() string
	// Should return the name of the package as defined at the top level of each file
	Name() string
	// Should return all gno filenames inside the package
	Files() []string
	// Should return a content reader for the given filename within the package
	Read(filename string) (io.ReadCloser, error)
}

type PackageReadWalkFunc func(filename string, r io.Reader, err error) error

func ReadWalkPackage(pkg Package, fn PackageReadWalkFunc) error {
	for _, filename := range pkg.Files() {
		if !isGnoFile(filename) {
			return nil
		}

		r, err := pkg.Read(filename)
		fnErr := fn(filename, r, err)
		r.Close()
		if fnErr != nil {
			return fnErr
		}
	}

	return nil
}

type fsPackage struct {
	path  string
	name  string
	dir   string
	files []string // filenames
}

// ParsePackage parses package from the given directory.
// It will return a nil package if no gno files are found.
// If a gnomod.toml is found, it will be used to determine the pkg path.
// If root is specified, it will be trimmed from the actual given dir to create the pkgpath if no gnomod.toml is found.
func ParsePackage(fset *token.FileSet, root string, dir string) (Package, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir %q: %w", dir, err)
	}

	var pkgname string

	gnofiles := []string{}
	for _, file := range files {
		name := file.Name()
		if !isGnoFile(name) {
			continue
		}

		// Ignore package name from test files
		if isTestFile(name) {
			gnofiles = append(gnofiles, name)
			continue
		}

		filename := filepath.Join(dir, name)
		f, err := parser.ParseFile(fset, filename, nil, parser.PackageClauseOnly)
		if err != nil {
			return nil, fmt.Errorf("unable to parse file %q: %w", filename, err)
		}

		if pkgname != "" && pkgname != f.Name.Name {
			return nil, fmt.Errorf("conflict package name between %q and %q", pkgname, f.Name.Name)
		}

		pkgname = f.Name.Name
		gnofiles = append(gnofiles, name)
	}

	if len(gnofiles) == 0 {
		return nil, nil // Not a package
	}

	var pkgpath string

	// Check for a gnomod.toml, in which case it will define the module path
	modpath := filepath.Join(dir, "gnomod.toml")
	data, err := os.ReadFile(modpath)
	switch {
	case os.IsNotExist(err):
		if len(root) > 0 {
			// Fallback on dir path trimmed from the root
			pkgpath = strings.TrimPrefix(dir, filepath.Clean(root))
			pkgpath = strings.TrimPrefix(pkgpath, "/")
		}

	case err == nil:
		mod, err := gnomod.ParseBytes(modpath, data)
		if err != nil {
			return nil, fmt.Errorf("unable to parse gnomod.toml %q: %w", modpath, err)
		}

		mod.Sanitize()
		if err := mod.Validate(); err != nil {
			return nil, fmt.Errorf("unable to validate gnomod.toml %q: %w", modpath, err)
		}

		pkgpath = mod.Module
	default:
		return nil, fmt.Errorf("unable to read %q: %w", modpath, err)
	}

	return &fsPackage{
		path:  pkgpath,
		files: gnofiles,
		dir:   dir,
		name:  pkgname,
	}, nil
}

func (p *fsPackage) Path() string {
	return p.path
}

func (p *fsPackage) Name() string {
	return p.name
}

func (p *fsPackage) Files() []string {
	return p.files
}

func (p *fsPackage) Read(filename string) (io.ReadCloser, error) {
	if !isGnoFile(filename) {
		return nil, fmt.Errorf("invalid gno file %q", filename)
	}

	path := filepath.Join(p.dir, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open %q: %w", path, err)
	}

	return file, nil
}
