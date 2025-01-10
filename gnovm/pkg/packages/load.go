package packages

import (
	"errors"
	"fmt"
	"go/token"
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
	IO            commands.IO
	Fetcher       pkgdownload.PackageFetcher
	Deps          bool
	Cache         PackagesMap
	SelfContained bool
	AllowEmpty    bool
	DepsPatterns  []string
}

func (conf *LoadConfig) applyDefaults() {
	if conf.IO == nil {
		conf.IO = commands.NewTestIO()
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

	fset := token.NewFileSet()

	expanded, err := expandPatterns(conf, patterns...)
	if err != nil {
		return nil, err
	}

	pkgs, err := readPackages(expanded, fset)
	if err != nil {
		return nil, err
	}

	if !conf.AllowEmpty {
		if len(pkgs) == 0 {
			return nil, errors.New("no packages")
		}
	}

	if !conf.Deps {
		return pkgs, nil
	}

	extra, err := expandPatterns(conf, conf.DepsPatterns...)
	if err != nil {
		return nil, err
	}
	for _, m := range extra {
		m.Match = []string{}
	}

	extraPkgs, err := readPackages(extra, fset)
	if err != nil {
		return nil, err
	}
	extraMap := NewPackagesMap(extraPkgs...)

	gnoroot := gnoenv.RootDir()

	toVisit := pkgs
	queuedByPkgPath := NewPackagesMap(pkgs...)
	markForVisit := func(pkg *Package) {
		queuedByPkgPath.Add(pkg)
		toVisit = append(toVisit, pkg)
	}

	visited := map[string]struct{}{}
	loaded := []*Package{}

	for {
		pkg, ok := fifoNext(&toVisit)
		if !ok {
			break
		}

		if added := setAdd(visited, pkg.Dir); !added {
			continue
		}

		for _, imp := range pkg.Imports.Merge(FileKindPackageSource, FileKindTest, FileKindXTest, FileKindFiletest) {
			// check if we already queued this dep
			if _, ok := queuedByPkgPath[imp]; ok {
				continue
			}

			// check if we have it in config cache
			if cached, ok := conf.Cache[imp]; ok {
				markForVisit(cached)
				continue
			}

			// check if we have it in extra deps patterns
			if extra, ok := extraMap[imp]; ok {
				markForVisit(extra)
				continue
			}

			// check if this is a stdlib and queue it
			if gnolang.IsStdlib(imp) {
				dir := filepath.Join(gnoroot, "gnovm", "stdlibs", filepath.FromSlash(imp))
				dirInfo, err := os.Stat(dir)
				if err == nil && !dirInfo.IsDir() {
					err = fmt.Errorf("%q is not a directory", dir)
				}
				if err != nil {
					pkg.Errors = append(pkg.Errors, err)
					delete(queuedByPkgPath, imp) // stop trying to get this lib, we can't
					continue
				}

				pkg := readPkgDir(dir, imp, fset)
				markForVisit(pkg)
				continue
			}

			if conf.SelfContained {
				pkg.Errors = append(pkg.Errors, fmt.Errorf("self-contained: package %q not found", imp))
				delete(queuedByPkgPath, imp) // stop trying to get this lib, we can't
				continue
			}

			dir := gnomod.PackageDir("", module.Version{Path: imp})
			if err := downloadPackage(conf, imp, dir); err != nil {
				pkg.Errors = append(pkg.Errors, err)
				delete(queuedByPkgPath, imp) // stop trying to get this lib, we can't
				continue
			}
			markForVisit(readPkgDir(dir, imp, fset))
		}

		loaded = append(loaded, pkg)
	}

	for _, pkg := range loaded {
		// TODO: this could be optimized
		pkg.Deps, err = listDeps(pkg.ImportPath, queuedByPkgPath)
		if err != nil {
			pkg.Errors = append(pkg.Errors, err)
		}
	}

	return loaded, nil
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

func fifoNext[T any](slice *[]T) (T, bool) {
	if len(*slice) == 0 {
		return *new(T), false
	}

	elem := (*slice)[0]
	*slice = (*slice)[1:]
	return elem, true
}

func setAdd[T comparable](set map[T]struct{}, val T) bool {
	if _, ok := set[val]; ok {
		return false
	}

	set[val] = struct{}{}
	return true
}
