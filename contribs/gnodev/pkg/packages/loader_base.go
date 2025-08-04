package packages

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"strings"
)

type BaseLoader struct {
	Resolver
}

func NewLoader(res ...Resolver) *BaseLoader {
	return &BaseLoader{ChainResolvers(res...)}
}

func (l BaseLoader) Name() string {
	return l.Resolver.Name()
}

func (l BaseLoader) Load(paths ...string) ([]Package, error) {
	fset := token.NewFileSet()
	visited, stack := map[string]bool{}, map[string]bool{}
	pkgs := make([]Package, 0)
	for _, root := range paths {
		deps, err := load(root, fset, l.Resolver, visited, stack)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, deps...)
	}

	return pkgs, nil
}

func (l BaseLoader) Resolve(path string) (*Package, error) {
	fset := token.NewFileSet()
	return l.Resolver.Resolve(fset, path)
}

func load(path string, fset *token.FileSet, resolver Resolver, visited, stack map[string]bool) ([]Package, error) {
	if stack[path] {
		return nil, fmt.Errorf("cycle detected: %s", path)
	}
	if visited[path] {
		return nil, nil
	}

	visited[path] = true

	mempkg, err := resolver.Resolve(fset, path)
	if err != nil {
		if errors.Is(err, ErrResolverPackageSkip) {
			return nil, nil
		}

		return nil, fmt.Errorf("unable to resolve package %q: %w", path, err)
	}

	var name string
	imports := map[string]struct{}{}
	for _, file := range mempkg.Files {
		fname := file.Name
		if !isGnoFile(fname) {
			continue
		}

		f, err := parser.ParseFile(fset, fname, file.Body, parser.ImportsOnly)
		if err != nil {
			return nil, fmt.Errorf("unable to parse file %q: %w", file.Name, err)
		}

		for _, imp := range f.Imports {
			if len(imp.Path.Value) <= 2 {
				continue
			}

			val := imp.Path.Value[1 : len(imp.Path.Value)-1]
			imports[val] = struct{}{}
		}

		if strings.HasSuffix(fname, "_filetest.gno") {
			continue
		}

		pname := strings.TrimSuffix(f.Name.Name, "_test")
		if name != "" && name != pname {
			return nil, fmt.Errorf("conflict package name between %q and %q", name, f.Name.Name)
		}

		name = pname
	}

	pkgs := []Package{}
	for imp := range imports {
		subDeps, err := load(imp, fset, resolver, visited, stack)
		if err != nil {
			return nil, fmt.Errorf("importing %q: %w", imp, err)
		}

		pkgs = append(pkgs, subDeps...)
	}
	pkgs = append(pkgs, *mempkg)

	stack[path] = false

	return pkgs, nil
}
