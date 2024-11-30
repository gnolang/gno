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

func (l *rootResolver) Name() string {
	return fmt.Sprintf("root<%s>", filepath.Base(l.root))
}

func NewRootResolver(rootpath string) Resolver {
	return &rootResolver{root: rootpath}
}

func (r *rootResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	dir := filepath.Join(r.root, path)
	_, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to determine dir for path %q: %w", path, ErrResolverPackageNotFound)
	}

	return ReadPackageFromDir(fset, path, dir)
}
