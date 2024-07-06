package packages

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
)

type fsResolver struct {
	rootsPath []string // Root folder
	fset      *token.FileSet
}

func NewFSResolver(rootpath ...string) Resolver {
	return &fsResolver{
		rootsPath: rootpath,
		fset:      token.NewFileSet(),
	}
}

func (res *fsResolver) Resolve(path string) (*Package, error) {
	dir, ok := res.findDirForPath(path)
	if !ok {
		return nil, fmt.Errorf("unable to determine dir for path %q: %w", path, ErrResolverPackageNotFound)
	}

	return ReadPackageFromDir(res.fset, path, dir)
}

func (res *fsResolver) findDirForPath(path string) (dir string, ok bool) {
	for _, root := range res.rootsPath {
		dir = filepath.Join(root, path)
		_, err := os.Stat(dir)
		if err != nil {
			continue
		}

		return dir, true
	}

	return "", false
}
