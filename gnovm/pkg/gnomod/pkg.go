package gnomod

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Pkg struct {
	name     string
	path     string
	draft    bool
	requires []string
}

// Name returns the name of the package.
func (p Pkg) Name() string {
	return p.name
}

// Path returns the path of the package.
func (p Pkg) Path() string {
	return p.path
}

// Draft returns whether the package is a draft.
func (p Pkg) Draft() bool {
	return p.draft
}

// Requires returns the required packages of the package.
func (p Pkg) Requires() []string {
	return p.requires
}

// sortPkgs sorts the given packages by their dependencies.
func SortPkgs(pkgs []Pkg) error {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	sortedPkgs := make([]Pkg, 0, len(pkgs))

	// Visit all packages
	for _, p := range pkgs {
		if err := visitPackage(p, pkgs, visited, onStack, &sortedPkgs); err != nil {
			return err
		}
	}

	copy(pkgs, sortedPkgs)
	return nil
}

// visitNode visits a package's and its dependencies dependencies and adds them to the sorted list.
func visitPackage(pkg Pkg, pkgs []Pkg, visited, onStack map[string]bool, sortedPkgs *[]Pkg) error {
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
			if err := visitPackage(p, pkgs, visited, onStack, sortedPkgs); err != nil {
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
	*sortedPkgs = append(*sortedPkgs, pkg)
	return nil
}

// ListPkgs lists all gno packages in the given root directory.
func ListPkgs(root string) ([]Pkg, error) {
	var pkgs []Pkg
	draft := make(map[string]bool)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		gnoModPath := filepath.Join(path, "gno.mod")
		data, err := os.ReadFile(gnoModPath)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}

		gnoMod, err := Parse(gnoModPath, data)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		gnoMod.Sanitize()
		if err := gnoMod.Validate(); err != nil {
			return fmt.Errorf("validate: %w", err)
		}

		// Ignore if draft
		if gnoMod.Draft {
			draft[gnoMod.Module.Mod.Path] = true
			return fs.SkipDir
		}

		pkgs = append(pkgs, Pkg{
			name:  gnoMod.Module.Mod.Path,
			path:  path,
			draft: gnoMod.Draft,
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

	return pkgsNotDependsOnDraft(pkgs, draft), nil
}

func pkgsNotDependsOnDraft(pkgs []Pkg, draft map[string]bool) []Pkg {
	res := make([]Pkg, 0, len(pkgs))
	for _, pkg := range pkgs {
		dependsOnDraft := false
		for _, req := range pkg.requires {
			if draft[req] {
				dependsOnDraft = true
				draft[pkg.name] = true
				break
			}
		}
		if !dependsOnDraft {
			res = append(res, pkg)
		}
	}
	return res
}
