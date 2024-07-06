package packages

import (
	"go/token"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

type LocalResolver struct {
	Path string
	Dir  string

	fset *token.FileSet
}

func NewLocalResolver(path, dir string) *LocalResolver {
	return &LocalResolver{
		fset: token.NewFileSet(),
		Path: path,
		Dir:  dir,
	}
}

func GuessLocalResolverFromRoots(dir string, roots []string) *LocalResolver {
	for _, root := range roots {
		if !strings.HasPrefix(dir, root) {
			continue
		}

		path := strings.TrimPrefix(dir, root)
		return NewLocalResolver(path, dir)
	}

	return nil
}

func GuessLocalResolverGnoMod(dir string) *LocalResolver {
	modfile, err := gnomod.ParseAt(dir)
	if err != nil {
		return nil
	}

	path := modfile.Module.Mod.Path
	return NewLocalResolver(path, dir)
}

func (lr LocalResolver) Resolve(path string) (*Package, error) {
	if path != lr.Path && path != lr.Dir {
		return nil, ErrResolverPackageNotFound
	}

	return ReadPackageFromDir(lr.fset, lr.Path, lr.Dir)
}
