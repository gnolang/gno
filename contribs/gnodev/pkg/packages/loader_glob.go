package packages

import (
	"fmt"
	"go/token"
	"path/filepath"
	"strings"
)

type GlobLoader struct {
	Resolver Resolver
	Root     string
}

func NewGlobResolverLoader(rootpath string, res ...Resolver) Loader {
	loader := GlobLoader{Root: rootpath}
	switch len(res) {
	case 0: // Skip
	case 1:
		loader.Resolver = res[0]
	default:
		loader.Resolver = ChainResolvers(res...)
	}

	return &loader
}

func (l GlobLoader) Load(paths ...string) ([]Package, error) {
	fset := token.NewFileSet()
	visited, stack := map[string]bool{}, map[string]bool{}
	pkgs := make([]Package, 0)
	for _, path := range paths {
		// format path to match directory from loader `Root`
		path = filepath.Clean(filepath.Join(l.Root, path) + "/")

		matches, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("invalid glob path: %w", err)
		}

		for _, match := range matches {
			// extract path
			mpath, _ := strings.CutPrefix(match, l.Root)
			mpath = strings.Trim(mpath, "/")

			deps, err := loadPackage(mpath, fset, l.Resolver, visited, stack)
			if err != nil {
				return nil, err
			}
			pkgs = append(pkgs, deps...)
		}
	}

	return pkgs, nil
}
