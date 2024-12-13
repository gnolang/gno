package packages

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
)

var ErrNoResolvers = errors.New("no resolvers setup")

type Loader interface {
	Load(paths ...string) ([]Package, error)
}

type ResolverLoader struct {
	Resolver
}

func NewResolverLoader(res ...Resolver) *ResolverLoader {
	var loader ResolverLoader
	switch len(res) {
	case 0: // Skip
	case 1:
		loader.Resolver = res[0]
	default:
		loader.Resolver = ChainResolvers(res...)
	}

	return &loader
}

func (l ResolverLoader) Load(paths ...string) ([]Package, error) {
	fset := token.NewFileSet()
	visited, stack := map[string]bool{}, map[string]bool{}
	pkgs := make([]Package, 0)
	for _, root := range paths {
		deps, err := loadPackage(root, fset, l.Resolver, visited, stack)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, deps...)
	}

	return pkgs, nil
}

func loadPackage(path string, fset *token.FileSet, resolver Resolver, visited, stack map[string]bool) ([]Package, error) {
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

		return nil, fmt.Errorf("unable to resolve package: %w", err)
	}

	var name string
	imports := map[string]struct{}{}
	for _, file := range mempkg.Files {
		fname := file.Name
		if !isGnoFile(fname) || isTestFile(fname) {
			continue
		}

		f, err := parser.ParseFile(fset, fname, file.Body, parser.ImportsOnly)
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
		subDeps, err := loadPackage(imp, fset, resolver, visited, stack)
		if err != nil {
			return nil, fmt.Errorf("importing %q: %w", imp, err)
		}

		pkgs = append(pkgs, subDeps...)
	}
	pkgs = append(pkgs, *mempkg)

	stack[path] = false

	return pkgs, nil
}
