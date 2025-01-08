package packages

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/mod/module"
)

type LoadConfig struct {
	IO      commands.IO
	Fetcher pkgdownload.PackageFetcher
	Deps    bool
}

func (conf *LoadConfig) applyDefaults() {
	if conf.IO == nil {
		conf.IO = commands.NewDefaultIO()
	}
	if conf.Fetcher == nil {
		conf.Fetcher = rpcpkgfetcher.New(nil)
	}
}

func Load(conf *LoadConfig, patterns ...string) ([]*Package, error) {
	conf.applyDefaults()

	expanded, err := expandPatterns(conf, patterns...)
	if err != nil {
		return nil, err
	}

	pkgs := readPackages(expanded)

	if !conf.Deps {
		return pkgs, nil
	}

	byPkgPath := make(map[string]*Package)
	for _, pkg := range pkgs {
		if pkg.ImportPath == "" {
			continue
		}
		if _, ok := byPkgPath[pkg.ImportPath]; ok {
			continue
		}
		byPkgPath[pkg.ImportPath] = pkg
	}

	visited := map[string]struct{}{}
	list := []*Package{}
	for pile := pkgs; len(pile) > 0; pile = pile[1:] {
		pkg := pile[0]
		if _, ok := visited[pkg.ImportPath]; ok {
			continue
		}
		visited[pkg.ImportPath] = struct{}{}

		for _, imp := range pkg.Imports.Merge(FileKindPackageSource, FileKindTest, FileKindXTest, FileKindFiletest) {
			if gnolang.IsStdlib(imp) {
				continue
			}

			if _, ok := byPkgPath[imp]; ok {
				continue
			}

			// must download

			dir := gnomod.PackageDir("", module.Version{Path: imp})
			if err := downloadPackage(conf, imp, dir); err != nil {
				pkg.Error = errors.Join(pkg.Error, err)
				byPkgPath[imp] = nil // stop trying to download pkg
				continue
			}

			byPkgPath[imp] = readPkg(dir)
		}

		list = append(list, pkg)
	}

	for _, pkg := range list {
		// TODO: this could be optimized
		pkg.Deps, err = listDeps(pkg.ImportPath, byPkgPath)
		if err != nil {
			pkg.Error = errors.Join(pkg.Error, err)
		}
	}

	return list, nil
}

func listDeps(target string, pkgs map[string]*Package) ([]string, error) {
	deps := []string{}
	err := listDepsRecursive(target, target, pkgs, &deps, make(map[string]struct{}))
	return deps, err
}

func listDepsRecursive(rootTarget string, target string, pkgs map[string]*Package, deps *[]string, visited map[string]struct{}) error {
	if _, ok := visited[target]; ok {
		return nil
	}
	pkg := pkgs[target]
	if pkg == nil {
		return fmt.Errorf("package %s not found", target)
	}
	visited[target] = struct{}{}
	var outErr error
	for _, imp := range pkg.Imports.Merge(FileKindPackageSource, FileKindTest, FileKindXTest, FileKindFiletest) {
		err := listDepsRecursive(rootTarget, imp, pkgs, deps, visited)
		if err != nil {
			outErr = errors.Join(outErr, err)
		}
	}
	if target != rootTarget {
		(*deps) = append(*deps, target)
	}
	return outErr
}

func (p *Package) MemPkg() (*gnovm.MemPackage, error) {
	files := []*gnovm.MemFile{}
	for _, cat := range p.Files {
		for _, f := range cat {
			body, err := os.ReadFile(filepath.Join(p.Dir, f))
			if err != nil {
				return nil, err
			}
			files = append(files, &gnovm.MemFile{
				Name: f,
				Body: string(body),
			})
		}
	}
	return &gnovm.MemPackage{
		Name:  p.Name,
		Path:  p.ImportPath,
		Files: files,
	}, nil
}
