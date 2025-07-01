package packages

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
)

type rootResolver struct {
	root string // Root folder
}

func NewRootResolver(path string) Resolver {
	if abs, err := filepath.Abs(path); err == nil {
		return &rootResolver{root: abs}
	}

	// fallback on path
	return &rootResolver{root: path}
}

func (r *rootResolver) Name() string {
	return fmt.Sprintf("root<%s>", filepath.Base(r.root))
}

func (r *rootResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	dir := filepath.Join(r.root, path)
	_, err := os.Stat(dir)
	if err != nil {
		return nil, ErrResolverPackageNotFound
	}

	return ReadPackageFromDir(fset, path, dir)
}
