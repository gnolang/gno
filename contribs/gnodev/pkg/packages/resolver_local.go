package packages

import (
	"fmt"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
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

func GuessLocalResolverFromRoots(dir string, roots []string) (res Resolver, path string) {
	for _, root := range roots {
		if !strings.HasPrefix(dir, root) {
			continue
		}

		path = strings.TrimPrefix(dir, root)
		return NewLocalResolver(path, dir), path
	}

	return nil, ""
}

func GuessLocalResolverGnoMod(dir string) (res Resolver, path string) {
	modfile, err := gnomod.ParseAt(dir)
	if err != nil {
		return nil, ""
	}

	path = modfile.Module.Mod.Path
	return NewLocalResolver(path, dir), path
}

func (lr LocalResolver) Resolve(fset *token.FileSet, path string) (*Package, error) {
	after, found := strings.CutPrefix(path, lr.Path)
	if !found {
		return nil, ErrResolverPackageNotFound
	}

	dir := filepath.Join(lr.Dir, after)
	return ReadPackageFromDir(fset, path, dir)
}
