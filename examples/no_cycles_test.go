package examples_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/stretchr/testify/require"
)

var injectedTestingLibs = []string{"encoding/json", "fmt", "os", "internal/os_test"}

// TestNoCycles checks that there is no import cycles in stdlibs and non-draft examples
func TestNoCycles(t *testing.T) {
	// find examples and stdlibs
	cfg := &packages.LoadConfig{SelfContained: true, Deps: true}
	pkgs, err := packages.Load(cfg, filepath.Join(gnoenv.RootDir(), "examples", "..."))
	require.NoError(t, err)

	// detect cycles
	visited := make(map[string]bool)
	for _, p := range pkgs {
		if p.Draft {
			continue
		}
		require.NoError(t, detectCycles(p, pkgs, visited))
	}
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
// Otherwise we will have false positive for example if we have these edges:
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
func detectCycles(root *packages.Package, pkgs []*packages.Package, visited map[string]bool) error {
	// check cycles in package's sources
	stack := []string{}
	if err := visitPackage(root, pkgs, visited, stack); err != nil {
		return fmt.Errorf("pkgsrc import: %w", err)
	}
	// check cycles in external tests' dependencies we might have missed
	if err := visitImports([]packages.FileKind{packages.FileKindXTest, packages.FileKindFiletest}, root, pkgs, visited, stack); err != nil {
		return fmt.Errorf("xtest import: %w", err)
	}

	// check cycles in tests' imports by marking the current package as visited while visiting the tests' imports
	// we also consider PackageSource imports here because tests can call package code
	visited = map[string]bool{root.ImportPath: true}
	stack = []string{root.ImportPath}
	if err := visitImports([]packages.FileKind{packages.FileKindPackageSource, packages.FileKindTest}, root, pkgs, visited, stack); err != nil {
		return fmt.Errorf("test import: %w", err)
	}

	return nil
}

// visitImports resolves and visits imports by kinds
func visitImports(kinds []packages.FileKind, root *packages.Package, pkgs []*packages.Package, visited map[string]bool, stack []string) error {
	for _, imp := range root.Imports.Merge(kinds...) {
		if slices.Contains(injectedTestingLibs, imp) {
			continue
		}
		idx := slices.IndexFunc(pkgs, func(p *packages.Package) bool { return p.ImportPath == imp })
		if idx == -1 {
			return fmt.Errorf("import %q not found for %q tests", imp, root.ImportPath)
		}
		if err := visitPackage(pkgs[idx], pkgs, visited, stack); err != nil {
			return fmt.Errorf("test import error: %w", err)
		}
	}

	return nil
}

// visitNode visits a package and its imports recursively. It only considers imports in PackageSource
func visitPackage(pkg *packages.Package, pkgs []*packages.Package, visited map[string]bool, stack []string) error {
	if slices.Contains(stack, pkg.ImportPath) {
		return fmt.Errorf("cycle detected: %s -> %s", strings.Join(stack, " -> "), pkg.ImportPath)
	}
	if visited[pkg.ImportPath] {
		return nil
	}

	visited[pkg.ImportPath] = true
	stack = append(stack, pkg.ImportPath)

	if err := visitImports([]packages.FileKind{packages.FileKindPackageSource}, pkg, pkgs, visited, stack); err != nil {
		return err
	}

	return nil
}
