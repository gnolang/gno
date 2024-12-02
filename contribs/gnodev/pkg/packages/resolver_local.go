package packages

import (
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
	return ReadPackageFromDir(fset, path, dir)
}
