package packages

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
)

type fsResolver struct {
	root string // Root folder
}

func NewFSResolver(rootpath string) Resolver {
	return &fsResolver{root: rootpath}
}

func (r *fsResolver) Name() string {
	return fmt.Sprintf("fs<%s>", filepath.Base(r.root))
}

func (r *fsResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	dir := filepath.Join(r.root, path)
	_, err := os.Stat(dir)
	if err != nil {
		return nil, ErrResolverPackageNotFound
	}

	return ReadPackageFromDir(fset, path, dir)
}
