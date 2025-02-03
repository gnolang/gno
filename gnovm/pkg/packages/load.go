package packages

import (
	"errors"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

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
	Cache         PkgList
	SelfContained bool
	AllowEmpty    bool
	DepsPatterns  []string
	Fset          *token.FileSet
}

var injectedTestingLibs = []string{"encoding/json", "fmt", "internal/os_test", "os"}

func IsInjectedTestingStdlib(pkgPath string) bool {
	return slices.Contains(injectedTestingLibs, pkgPath)
}

func (conf *LoadConfig) applyDefaults() {
	if conf.IO == nil {
		conf.IO = commands.NewTestIO()
	}
	if conf.Fetcher == nil {
		conf.Fetcher = rpcpkgfetcher.New(nil)
	}
	if conf.Fset == nil {
		conf.Fset = token.NewFileSet()
	}
}

func LoadWorkspace(conf *LoadConfig, dir string) (PkgList, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	workRoot, err := gnomod.FindRootDir(absDir)
	if err != nil {
		return nil, err
	}

	return Load(conf, filepath.Join(workRoot, "..."))
}

func Load(conf *LoadConfig, patterns ...string) (PkgList, error) {
	if conf == nil {
		conf = &LoadConfig{}
	}
	conf.applyDefaults()

	expanded, err := expandPatterns(conf, patterns...)
	if err != nil {
		return nil, err
	}

	pkgs, err := readPackages(expanded, nil, conf.Fset)
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

	extraPkgs, err := readPackages(extra, pkgs, conf.Fset)
	if err != nil {
		return nil, err
	}
	extraMap := NewPackagesMap(extraPkgs...)

	toVisit := []*Package(pkgs)
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

		for _, imp := range pkg.ImportsSpecs.Merge(FileKindPackageSource, FileKindTest, FileKindXTest, FileKindFiletest) {
			if IsInjectedTestingStdlib(imp.PkgPath) {
				continue
			}

			// check if we already queued this dep
			if _, ok := queuedByPkgPath[imp.PkgPath]; ok {
				continue
			}

			// check if we have it in config cache
			if cached := conf.Cache.Get(imp.PkgPath); cached != nil {
				markForVisit(cached)
				continue
			}

			// check if we have it in extra deps patterns
			if extra, ok := extraMap[imp.PkgPath]; ok {
				markForVisit(extra)
				continue
			}

			// check if this is a stdlib and load it from gnoroot if available
			if gnolang.IsStdlib(imp.PkgPath) {
				dir := filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs", filepath.FromSlash(imp.PkgPath))
				dirInfo, err := os.Stat(dir)
				if err != nil || !dirInfo.IsDir() {
					err := &Error{
						Pos: filepath.Join(filepath.FromSlash(pkg.Dir), conf.Fset.Position(imp.Spec.Pos()).String()),
						Msg: fmt.Sprintf("package %s is not in std (%s)", imp.PkgPath, dir),
					}
					pkg.Errors = append(pkg.Errors, err)
				}
				markForVisit(readPkgDir(dir, imp.PkgPath, conf.Fset))
				continue
			}

			if conf.SelfContained {
				pkg.Errors = append(pkg.Errors, &Error{
					Pos: pkg.Dir,
					Msg: fmt.Sprintf("package %q not found (self-contained)", imp.PkgPath),
				})
				continue
			}

			dir := gnomod.PackageDir("", module.Version{Path: imp.PkgPath})
			if err := downloadPackage(conf, imp.PkgPath, dir); err != nil {
				pkg.Errors = append(pkg.Errors, &Error{
					Pos: pkg.Dir,
					Msg: fmt.Sprintf("download %q: %v", imp.PkgPath, err),
				})
				continue
			}
			markForVisit(readPkgDir(dir, imp.PkgPath, conf.Fset))
		}

		loaded = append(loaded, pkg)
	}

	for _, pkg := range loaded {
		// TODO: this could be optimized
		var errs []*Error
		pkg.Deps, errs = listDeps(pkg, queuedByPkgPath)
		pkg.Errors = append(pkg.Errors, errs...)
	}

	return loaded, nil
}

func listDeps(target *Package, pkgs PackagesMap) ([]string, []*Error) {
	deps := []string{}
	err := listDepsRecursive(target, target, pkgs, &deps, make(map[string]struct{}))
	return deps, err
}

func listDepsRecursive(rootTarget *Package, target *Package, pkgs PackagesMap, deps *[]string, visited map[string]struct{}) []*Error {
	if _, ok := visited[target.ImportPath]; ok {
		return nil
	}
	visited[target.ImportPath] = struct{}{}
	var outErrs []*Error
	for _, imp := range target.ImportsSpecs.Merge(FileKindPackageSource, FileKindTest, FileKindXTest, FileKindFiletest) {
		if IsInjectedTestingStdlib(imp.PkgPath) {
			continue
		}
		dep := pkgs[imp.PkgPath]
		if dep == nil {
			return []*Error{{
				Pos: rootTarget.Dir,
				Msg: fmt.Sprintf("package %q not found", imp.PkgPath),
			}}
		}
		errs := listDepsRecursive(rootTarget, dep, pkgs, deps, visited)
		outErrs = append(outErrs, errs...)
	}
	if target != rootTarget {
		(*deps) = append(*deps, target.ImportPath)
	}
	return outErrs
}

func (p *Package) MemPkg() (*gnovm.MemPackage, error) {
	files := []*gnovm.MemFile{}
	for _, cat := range p.Files {
		for _, f := range cat {
			if !strings.HasSuffix(f, ".gno") {
				continue
			}
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
	sort.Slice(files, func(i int, j int) bool {
		return files[i].Name < files[j].Name
	})
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
