package examples_test

import (
	"fmt"
	"io/fs"
	"os"
	pathlib "path"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/stretchr/testify/require"
)

var injectedTestingLibs = []string{"encoding/json", "fmt", "os", "internal/os_test"}

// TestNoCycles checks that there is no import cycles in stdlibs and examples
func TestNoCycles(t *testing.T) {
	// find stdlibs
	gnoRoot := gnoenv.RootDir()
	pkgs, err := listPkgs(gnomod.Pkg{
		Dir:  filepath.Join(gnoRoot, "gnovm", "stdlibs"),
		Name: "",
	})
	require.NoError(t, err)

	// find examples
	examples, err := gnomod.ListPkgs(filepath.Join(gnoRoot, "examples"))
	require.NoError(t, err)
	for _, example := range examples {
		if example.Draft {
			continue
		}
		examplePkgs, err := listPkgs(example)
		require.NoError(t, err)
		pkgs = append(pkgs, examplePkgs...)
	}

	// detect cycles
	for _, p := range pkgs {
		require.NoError(t, detectCycles(p, pkgs))
	}
}

type testPkg struct {
	Dir     string
	PkgPath string
	Imports packages.ImportsMap
}

// listPkgs lists all packages in rootMod
func listPkgs(rootMod gnomod.Pkg) ([]testPkg, error) {
	res := []testPkg{}
	rootDir := rootMod.Dir
	visited := map[string]struct{}{}
	if err := fs.WalkDir(os.DirFS(rootDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".gno") {
			return nil
		}
		subPath := filepath.Dir(path)
		dir := filepath.Join(rootDir, subPath)
		if _, ok := visited[dir]; ok {
			return nil
		}
		visited[dir] = struct{}{}

		subPkgPath := pathlib.Join(rootMod.Name, subPath)

		pkg := testPkg{
			Dir:     dir,
			PkgPath: subPkgPath,
		}

		memPkg, err := readPkg(pkg.Dir, pkg.PkgPath)
		if err != nil {
			return fmt.Errorf("read pkg %q: %w", pkg.Dir, err)
		}
		pkg.Imports, err = packages.Imports(memPkg, nil)
		if err != nil {
			return fmt.Errorf("list imports of %q: %w", memPkg.Path, err)
		}

		res = append(res, pkg)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk dirs at %q: %w", rootDir, err)
	}
	return res, nil
}

func fileImportsToStrings(fis []packages.FileImport) []string {
	res := make([]string, len(fis))
	for i, fi := range fis {
		res[i] = fi.PkgPath
	}
	return res
}

// detectCycles detects import cycles
//
// We need to check
// 3 kinds of nodes
//
// - normal pkg: compiled source
//
// - xtest pkg: external test source (include xtests and filetests), can be treated as their own package
//
// - test pkg: embedded test sources,
// these should not have their corresponding normal package in their dependencies tree
//
// The tricky thing is that we need to split test sources and normal source
// while not considering them as distincitive packages.
// Because if we don't we will have false positive for example if we have these edges:
//
// - foo_pkg/foo_test.go imports bar_pkg
//
// - bar_pkg/bar_test.go import foo_pkg
//
// In go, the above example is allowed
// but the following is not
//
// - foo_pkg/foo.go imports bar_pkg
//
// - bar_pkg/bar_test.go imports foo_pkg
//
// This implementation can be optimized with better graph theory
func detectCycles(root testPkg, pkgs []testPkg) error {
	// check normal cycles
	{
		visited := make(map[string]bool)
		stack := []string{}
		if err := visitPackage(root, pkgs, visited, stack); err != nil {
			return fmt.Errorf("compiled import error: %w", err)
		}
	}

	// check xtest cycles
	{
		visited := make(map[string]bool)
		stack := []string{}
		for _, imp := range root.Imports.Merge(packages.FileKindXTest, packages.FileKindFiletest) {
			if slices.Contains(injectedTestingLibs, imp.PkgPath) {
				continue
			}
			pkg := (*testPkg)(nil)
			for _, p := range pkgs {
				if p.PkgPath != imp.PkgPath {
					continue
				}
				pkg = &p
				break
			}
			if pkg == nil {
				return fmt.Errorf("import %q not found for %q xtests", imp.PkgPath, root.PkgPath)
			}
			if err := visitPackage(*pkg, pkgs, visited, stack); err != nil {
				return fmt.Errorf("xtest import error: %w", err)
			}
		}
	}

	// check test cycles
	{
		visited := map[string]bool{root.PkgPath: true}
		stack := []string{root.PkgPath}
		for _, imp := range root.Imports.Merge(packages.FileKindPackageSource, packages.FileKindTest) {
			if slices.Contains(injectedTestingLibs, imp.PkgPath) {
				continue
			}
			pkg := (*testPkg)(nil)
			for _, p := range pkgs {
				if p.PkgPath != imp.PkgPath {
					continue
				}
				pkg = &p
				break
			}
			if pkg == nil {
				return fmt.Errorf("import %q not found for %q tests", imp.PkgPath, root.PkgPath)
			}
			if err := visitPackage(*pkg, pkgs, visited, stack); err != nil {
				return fmt.Errorf("test import error: %w", err)
			}
		}
	}
	return nil
}

// visitNode visits a package and its imports recursively. It only considers imports in PackageSource
func visitPackage(pkg testPkg, pkgs []testPkg, visited map[string]bool, stack []string) error {
	if slices.Contains(stack, pkg.PkgPath) {
		return fmt.Errorf("cycle detected: %s -> %s", strings.Join(stack, " -> "), pkg.PkgPath)
	}
	if visited[pkg.PkgPath] {
		return nil
	}

	visited[pkg.PkgPath] = true
	stack = append(stack, pkg.PkgPath)

	// Visit package's imports
	for _, imp := range pkg.Imports.Merge(packages.FileKindPackageSource) {
		if slices.Contains(injectedTestingLibs, imp.PkgPath) {
			continue
		}

		found := false
		for _, p := range pkgs {
			if p.PkgPath != imp.PkgPath {
				continue
			}
			if err := visitPackage(p, pkgs, visited, stack); err != nil {
				return err
			}
			found = true
			break
		}
		if !found {
			return fmt.Errorf("missing dependency '%s' for package '%s'", imp.PkgPath, pkg.PkgPath)
		}
	}

	return nil
}

// readPkg reads the sources of a package. It includes all .gno files but ignores the package name
func readPkg(dir string, pkgPath string) (*gnovm.MemPackage, error) {
	list, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	memPkg := &gnovm.MemPackage{Path: pkgPath}
	for _, entry := range list {
		fpath := filepath.Join(dir, entry.Name())
		if !strings.HasSuffix(fpath, ".gno") {
			continue
		}
		fname := filepath.Base(fpath)
		bz, err := os.ReadFile(fpath)
		if err != nil {
			return nil, err
		}
		memPkg.Files = append(memPkg.Files,
			&gnovm.MemFile{
				Name: fname,
				Body: string(bz),
			})
	}
	return memPkg, nil
}
