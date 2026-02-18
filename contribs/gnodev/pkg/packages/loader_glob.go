package packages

import (
	"fmt"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type GlobLoader struct {
	Root     string
	Resolver Resolver
}

func NewGlobLoader(rootpath string, res ...Resolver) *GlobLoader {
	return &GlobLoader{rootpath, ChainResolvers(res...)}
}

func (l GlobLoader) Name() string {
	return l.Resolver.Name()
}

func (l GlobLoader) MatchPaths(globs ...string) ([]string, error) {
	if l.Root == "" {
		return globs, nil
	}

	if _, err := os.Stat(l.Root); err != nil {
		return nil, fmt.Errorf("unable to stat root: %w", err)
	}

	mpaths := []string{}
	for _, input := range globs {
		cleanInput := path.Clean(input)
		gpath, err := Parse(cleanInput)
		if err != nil {
			return nil, fmt.Errorf("invalid glob path %q: %w", input, err)
		}

		base := gpath.StarFreeBase()
		if base == cleanInput {
			mpaths = append(mpaths, base)
			continue
		}

		root := l.Root
		err = filepath.WalkDir(root, func(dirpath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			path, err := filepath.Rel(root, dirpath)
			if err != nil {
				return err
			}

			// normalize filepath to path
			path = NormalizeFilepathToPath(path)

			if !d.IsDir() {
				return nil
			}

			if strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}

			if d.Name() == "filetests" {
				// We don't need to load file tests
				return filepath.SkipDir
			}

			if gpath.Match(path) {
				mpaths = append(mpaths, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walking directory %q: %w", root, err)
		}
	}

	return mpaths, nil
}

func (l GlobLoader) Load(gpaths ...string) ([]Package, error) {
	paths, err := l.MatchPaths(gpaths...)
	if err != nil {
		return nil, fmt.Errorf("match glob pattern error: %w", err)
	}

	loader := &BaseLoader{Resolver: l.Resolver}
	return loader.Load(paths...)
}

func (l GlobLoader) Resolve(path string) (*Package, error) {
	return l.Resolver.Resolve(token.NewFileSet(), path)
}
