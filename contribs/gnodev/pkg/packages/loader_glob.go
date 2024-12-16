package packages

import (
	"fmt"
	"io/fs"
	"os"
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

func (l GlobLoader) MatchPaths(globs ...string) ([]string, error) {
	if l.Root == "" {
		return globs, nil
	}

	if _, err := os.Stat(l.Root); err != nil {
		return nil, fmt.Errorf("unable to stats: %w", err)
	}

	mpaths := []string{}
	for _, input := range globs {
		cleanInputs := filepath.Clean(input)
		gpath, err := Parse(cleanInputs)
		if err != nil {
			return nil, fmt.Errorf("invalid glob path %q: %w", input, err)
		}

		base := gpath.StarFreeBase()
		if base == cleanInputs {
			mpaths = append(mpaths, base)
			continue
		}

		root := filepath.Join(l.Root, base)
		err = filepath.WalkDir(root, func(dirpath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				return nil
			}

			if strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}

			path := strings.TrimPrefix(dirpath, l.Root+"/")
			if gpath.Match(path) {
				mpaths = append(mpaths, path)
				return nil
			}

			return nil
		})
	}

	return mpaths, nil
}

func (l GlobLoader) Load(gpaths ...string) ([]Package, error) {
	paths, err := l.MatchPaths(gpaths...)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve dir paths: %w", err)
	}

	loader := &BaseLoader{Resolver: l.Resolver}
	return loader.Load(paths...)
}
