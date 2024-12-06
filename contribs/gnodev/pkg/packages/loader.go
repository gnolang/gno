package packages

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"strings"
)

var ErrNoResolvers = errors.New("no resolvers setup")

type Loader struct {
	Host     string
	Paths    []string
	Resolver Resolver
}

func (l Loader) LoadPackages() ([]Package, error) {
	if l.Resolver == nil {
		return nil, ErrNoResolvers
	}

	fset := token.NewFileSet()
	visited, stack := map[string]bool{}, map[string]bool{}
	pkgs := make([]Package, 0)
	for _, root := range l.Paths {
		deps, err := l.loadPackage(root, fset, l.Resolver, visited, stack)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, deps...)
	}

	return pkgs, nil
}

func (l Loader) loadPackage(path string, fset *token.FileSet, resolver Resolver, visited, stack map[string]bool) ([]Package, error) {
	if stack[path] {
		return nil, fmt.Errorf("cycle detected: %s", path)
	}
	if visited[path] {
		return nil, nil
	}

	visited[path] = true

	// do not try to load package that hasn't been prefixed
	if l.Host != "" && !strings.HasPrefix(path, l.Host) {
		return nil, nil
	}

	mempkg, err := resolver.Resolve(fset, path)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve package: %w", err)
	}

	var name string
	imports := map[string]struct{}{}
	for _, file := range mempkg.Files {
		f, err := parser.ParseFile(fset, file.Name, file.Body, parser.ImportsOnly)
		if err != nil {
			return nil, fmt.Errorf("unable to parse file %q: %w", file.Name, err)
		}

		if name != "" && name != f.Name.Name {
			return nil, fmt.Errorf("conflict package name between %q and %q", name, f.Name.Name)
		}

		for _, imp := range f.Imports {
			if len(imp.Path.Value) <= 2 {
				continue
			}

			val := imp.Path.Value[1 : len(imp.Path.Value)-1]
			imports[val] = struct{}{}
		}

		name = f.Name.Name
	}

	pkgs := []Package{}
	for imp := range imports {
		subDeps, err := l.loadPackage(imp, fset, resolver, visited, stack)
		if err != nil {
			return nil, fmt.Errorf("importing %q: %w", imp, err)
		}

		pkgs = append(pkgs, subDeps...)
	}
	pkgs = append(pkgs, *mempkg)

	stack[path] = false

	return pkgs, nil
}
