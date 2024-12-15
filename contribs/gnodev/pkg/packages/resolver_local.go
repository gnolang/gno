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

func (l *LocalResolver) Name() string {
	return fmt.Sprintf("local<%s>", filepath.Base(l.Dir))
}

func NewLocalResolver(path, dir string) *LocalResolver {
	return &LocalResolver{
		Path: path,
		Dir:  dir,
	}
}

func (lr LocalResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	after, found := strings.CutPrefix(path, lr.Path)
	if !found {
		return nil, ErrResolverPackageNotFound
	}

	dir := filepath.Join(lr.Dir, after)
	pkg, err := ReadPackageFromDir(fset, path, dir)
	if err != nil && after == "" && errors.Is(err, ErrResolverPackageSkip) {
		return nil, fmt.Errorf("empty local package %q", err)
	}

	return pkg, nil
}
