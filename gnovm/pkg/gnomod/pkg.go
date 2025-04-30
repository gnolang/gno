package gnomod

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
)

type Pkg struct {
	Dir     string   // absolute path to package dir
	Name    string   // package name
	Imports []string // direct imports of this pkg
	Draft   bool     // whether the package is a draft
}

type SubPkg struct {
	Dir        string   // absolute path to package dir
	ImportPath string   // import path of package
	Root       string   // Root dir containing this package, i.e dir containing gno.mod file
	Imports    []string // imports used by this package

	GnoFiles         []string // .gno source files (excluding TestGnoFiles, FiletestGnoFiles)
	TestGnoFiles     []string // _test.gno source files
	FiletestGnoFiles []string // _filetest.gno source files
}

type (
	PkgList       []Pkg
	SortedPkgList []Pkg
)

// sortPkgs sorts the given packages by their dependencies.
func (pl PkgList) Sort() (SortedPkgList, error) {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	sortedPkgs := make([]Pkg, 0, len(pl))

	// Visit all packages
	for _, p := range pl {
		if err := visitPackage(p, pl, visited, onStack, &sortedPkgs); err != nil {
			return nil, err
		}
	}

	return sortedPkgs, nil
}

// visitNode visits a package's and its dependencies dependencies and adds them to the sorted list.
func visitPackage(pkg Pkg, pkgs []Pkg, visited, onStack map[string]bool, sortedPkgs *[]Pkg) error {
	if onStack[pkg.Name] {
		return fmt.Errorf("cycle detected: %s", pkg.Name)
	}
	if visited[pkg.Name] {
		return nil
	}

	visited[pkg.Name] = true
	onStack[pkg.Name] = true

	// Visit package's dependencies
	for _, imp := range pkg.Imports {
		found := false
		for _, p := range pkgs {
			if p.Name != imp {
				continue
			}
			if err := visitPackage(p, pkgs, visited, onStack, sortedPkgs); err != nil {
				return err
			}
			found = true
			break
		}
		if !found {
			return fmt.Errorf("missing dependency '%s' for package '%s'", imp, pkg.Name)
		}
	}

	onStack[pkg.Name] = false
	*sortedPkgs = append(*sortedPkgs, pkg)
	return nil
}

// ListPkgs lists all gno packages in the given root directory.
func ListPkgs(root string) (PkgList, error) {
	var pkgs []Pkg

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
			return fmt.Errorf("failed to validate gno.mod in %s: %w", gnoModPath, err)
		}

		pkg, err := gnolang.ReadMemPackage(path, gnoMod.Module.Mod.Path)
		if err != nil {
			// ignore package files on error
			pkg = &gnovm.MemPackage{}
		}

		importsMap, err := packages.Imports(pkg, nil)
		if err != nil {
			// ignore imports on error
			importsMap = nil
		}
		importsRaw := importsMap.Merge(packages.FileKindPackageSource, packages.FileKindTest, packages.FileKindXTest)

		imports := make([]string, 0, len(importsRaw))
		for _, imp := range importsRaw {
			// remove self and standard libraries from imports
			if imp.PkgPath != gnoMod.Module.Mod.Path &&
				!gnolang.IsStdlib(imp.PkgPath) {
				imports = append(imports, imp.PkgPath)
			}
		}

		pkgs = append(pkgs, Pkg{
			Dir:     path,
			Name:    gnoMod.Module.Mod.Path,
			Draft:   gnoMod.Draft,
			Imports: imports,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}

// GetNonDraftPkgs returns packages that are not draft
// and have no direct or indirect draft dependencies.
func (sp SortedPkgList) GetNonDraftPkgs() SortedPkgList {
	res := make([]Pkg, 0, len(sp))
	draft := make(map[string]bool)

	for _, pkg := range sp {
		if pkg.Draft {
			draft[pkg.Name] = true
			continue
		}
		dependsOnDraft := false
		for _, req := range pkg.Imports {
			if draft[req] {
				dependsOnDraft = true
				draft[pkg.Name] = true
				break
			}
		}
		if !dependsOnDraft {
			res = append(res, pkg)
		}
	}
	return res
}

// SubPkgsFromPaths returns a list of subpackages from the given paths.
func SubPkgsFromPaths(paths []string) ([]*SubPkg, error) {
	for _, path := range paths {
		fi, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			continue
		}
		if filepath.Ext(path) != ".gno" {
			return nil, fmt.Errorf("files must be .gno files: %s", path)
		}

		subPkg, err := GnoFileSubPkg(paths)
		if err != nil {
			return nil, err
		}
		return []*SubPkg{subPkg}, nil
	}

	subPkgs := make([]*SubPkg, 0, len(paths))
	for _, path := range paths {
		subPkg := SubPkg{}

		matches, err := filepath.Glob(filepath.Join(path, "*.gno"))
		if err != nil {
			return nil, fmt.Errorf("failed to match pattern: %w", err)
		}

		subPkg.Dir = path
		for _, match := range matches {
			if strings.HasSuffix(match, "_test.gno") {
				subPkg.TestGnoFiles = append(subPkg.TestGnoFiles, match)
				continue
			}

			if strings.HasSuffix(match, "_filetest.gno") {
				subPkg.FiletestGnoFiles = append(subPkg.FiletestGnoFiles, match)
				continue
			}
			subPkg.GnoFiles = append(subPkg.GnoFiles, match)
		}

		subPkgs = append(subPkgs, &subPkg)
	}

	return subPkgs, nil
}

// GnoFileSubPkg returns a subpackage from the given .gno files.
func GnoFileSubPkg(files []string) (*SubPkg, error) {
	subPkg := SubPkg{}
	firstDir := ""
	for _, file := range files {
		if filepath.Ext(file) != ".gno" {
			return nil, fmt.Errorf("files must be .gno files: %s", file)
		}

		fi, err := os.Stat(file)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			return nil, fmt.Errorf("%s is a directory, should be a Gno file", file)
		}

		dir := filepath.Dir(file)
		if firstDir == "" {
			firstDir = dir
		}
		if dir != firstDir {
			return nil, fmt.Errorf("all files must be in one directory; have %s and %s", firstDir, dir)
		}

		if strings.HasSuffix(file, "_test.gno") {
			subPkg.TestGnoFiles = append(subPkg.TestGnoFiles, file)
			continue
		}

		if strings.HasSuffix(file, "_filetest.gno") {
			subPkg.FiletestGnoFiles = append(subPkg.FiletestGnoFiles, file)
			continue
		}
		subPkg.GnoFiles = append(subPkg.GnoFiles, file)
	}
	subPkg.Dir = firstDir

	return &subPkg, nil
}
