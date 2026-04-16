package packages

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"
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
	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, ErrResolverPackageNotFound
	}

	dir := filepath.Join(r.root, clean)
	rel, err := filepath.Rel(r.root, dir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, ErrResolverPackageNotFound
	}

	_, err = os.Stat(dir)
	if err != nil {
		return nil, ErrResolverPackageNotFound
	}

	return ReadPackageFromDir(fset, path, dir)
}
