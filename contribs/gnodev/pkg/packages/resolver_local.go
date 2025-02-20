package packages

import (
	"fmt"
	"go/token"
	"path"
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
	return fmt.Sprintf("local<%s>", path.Base(r.Dir))
}

func (r LocalResolver) IsValid() bool {
	pkg, err := r.Resolve(token.NewFileSet(), r.Path)
	return err == nil && pkg != nil
}

func (r LocalResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	after, found := strings.CutPrefix(path, r.Path)
	if !found {
		return nil, ErrResolverPackageNotFound
	}

	dir := filepath.Join(r.Dir, after)
	return ReadPackageFromDir(fset, path, dir)
}
