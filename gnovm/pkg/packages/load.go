package packages

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/mod/module"
)

type LoadConfig struct {
	IO              commands.IO
	Fetcher         pkgdownload.PackageFetcher
	Deps            bool
	Cache           map[string]*Package
	GnorootExamples bool // allow loading packages from gnoroot examples if not found in the set
	SelfContained   bool
}

func (conf *LoadConfig) applyDefaults() {
	if conf.IO == nil {
		conf.IO = commands.NewDefaultIO()
	}
	if conf.Fetcher == nil {
		conf.Fetcher = rpcpkgfetcher.New(nil)
	}
	if conf.Cache == nil {
		conf.Cache = map[string]*Package{}
	}
}

func Load(conf *LoadConfig, patterns ...string) (PkgList, error) {
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
	index := func(pkg *Package) {
		if pkg.ImportPath == "" {
			return
		}
		if _, ok := byPkgPath[pkg.ImportPath]; ok {
			return
		}
		byPkgPath[pkg.ImportPath] = pkg
	}
	for _, pkg := range pkgs {
		index(pkg)
	}

	gnoroot := gnoenv.RootDir()

	visited := map[string]struct{}{}
	list := []*Package{}
	pile := pkgs
	pileDown := func(pkg *Package) {
		index(pkg)
		pile = append(pile, pkg)
	}
	for ; len(pile) > 0; pile = pile[1:] {
		pkg := pile[0]
		if _, ok := visited[pkg.ImportPath]; ok {
			continue
		}
		visited[pkg.ImportPath] = struct{}{}

		for _, imp := range pkg.Imports.Merge(FileKindPackageSource, FileKindTest, FileKindXTest, FileKindFiletest) {
			if _, ok := byPkgPath[imp]; ok {
				continue
			}

			if cached, ok := conf.Cache[imp]; ok {
				pileDown(cached)
				continue
			}

			if gnolang.IsStdlib(imp) {
				dir := filepath.Join(gnoroot, "gnovm", "stdlibs", filepath.FromSlash(imp))
				finfo, err := os.Stat(dir)
				if err == nil && !finfo.IsDir() {
					err = fmt.Errorf("stdlib %q not found", imp)
				}
				if err != nil {
					pkg.Error = errors.Join(pkg.Error, err)
					byPkgPath[imp] = nil // stop trying to get this lib, we can't
					continue
				}

				pileDown(readPkg(dir, imp))
				continue
			}

			if conf.GnorootExamples {
				examplePkgDir := filepath.Join(gnoroot, "examples", filepath.FromSlash(imp))
				finfo, err := os.Stat(examplePkgDir)
				if err == nil && finfo.IsDir() {
					pileDown(readPkg(examplePkgDir, imp))
					continue
				}
			}

			if conf.SelfContained {
				pkg.Error = errors.Join(pkg.Error, fmt.Errorf("self-contained: package %q not found", imp))
				byPkgPath[imp] = nil // stop trying to get this lib, we can't
				continue
			}

			dir := gnomod.PackageDir("", module.Version{Path: imp})
			if err := downloadPackage(conf, imp, dir); err != nil {
				pkg.Error = errors.Join(pkg.Error, err)
				byPkgPath[imp] = nil // stop trying to download pkg, we can't
				continue
			}
			pileDown(readPkg(dir, imp))
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
