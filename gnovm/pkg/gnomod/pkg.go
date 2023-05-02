package gnomod

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type pkg struct {
	name     string
	path     string
	requires []string
}

// Name returns the name of the package.
func (p pkg) Name() string {
	return p.name
}

// Path returns the path of the package.
func (p pkg) Path() string {
	return p.path
}

// Requires returns the required packages of the package.
func (p pkg) Requires() []string {
	return p.requires
}

// sortPkgs sorts the given packages by their dependencies.
func SortPkgs(pkgs []pkg) error {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	sortedPkgs := make([]pkg, 0, len(pkgs))

	var visit func(pkg pkg) error
	visit = func(pkg pkg) error {
		if onStack[pkg.name] {
			return fmt.Errorf("cycle detected: %s", pkg.name)
		}
		if visited[pkg.name] {
			return nil
		}

		visited[pkg.name] = true
		onStack[pkg.name] = true

		// Visit package's dependencies
		for _, req := range pkg.requires {
			found := false
			for _, p := range pkgs {
				if p.name != req {
					continue
				}
				if err := visit(p); err != nil {
					return err
				}
				found = true
				break
			}
			if !found {
				return fmt.Errorf("missing dependency '%s' for package '%s'", req, pkg.name)
			}
		}

		onStack[pkg.name] = false
		sortedPkgs = append(sortedPkgs, pkg)
		return nil
	}

	// Visit all packages
	for _, p := range pkgs {
		if err := visit(p); err != nil {
			return err
		}
	}

	copy(pkgs, sortedPkgs)
	return nil
}

// ListPkgs lists all gno packages in the given root directory.
func ListPkgs(root string) ([]pkg, error) {
	var pkgs []pkg

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		goModPath := filepath.Join(path, "gno.mod")
		data, err := os.ReadFile(goModPath)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}

		gnoMod, err := Parse(goModPath, data)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		gnoMod.Sanitize()
		if err := gnoMod.Validate(); err != nil {
			return fmt.Errorf("validate: %w", err)
		}
		pkgs = append(pkgs, pkg{
			name: gnoMod.Module.Mod.Path,
			path: path,
			requires: func() []string {
				var reqs []string
				for _, req := range gnoMod.Require {
					reqs = append(reqs, req.Mod.Path)
				}
				return reqs
			}(),
		})
		return fs.SkipDir
	})
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}
