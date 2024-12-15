package packages

import (
	"errors"
	"fmt"
	"go/token"
	"path/filepath"
	"strings"
)

type LocalResolver struct {
	Path string
	Dir  string
}

func NewLocalResolver(path, dir string) *LocalResolver {
	return &LocalResolver{
		Path: path,
		Dir:  dir,
	}
}

func (r *LocalResolver) Name() string {
	return fmt.Sprintf("local<%s>", filepath.Base(r.Dir))
}

func (r LocalResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	after, found := strings.CutPrefix(path, r.Path)
	if !found {
		return nil, ErrResolverPackageNotFound
	}

	dir := filepath.Join(r.Dir, after)
	pkg, err := ReadPackageFromDir(fset, path, dir)

	if err != nil && after == "" && errors.Is(err, ErrResolverPackageSkip) {
		return nil, fmt.Errorf("empty local package %w", err) // local package cannot be empty
	}

	return pkg, nil
}
