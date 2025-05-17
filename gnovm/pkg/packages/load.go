package packages

import (
	"errors"
	"fmt"
	"go/token"
	"io"
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
	"github.com/gnolang/gno/tm2/pkg/std"
	"golang.org/x/mod/module"
)

type LoadConfig struct {
	Fetcher    pkgdownload.PackageFetcher // package fetcher used to load dependencies not present in patterns. Could be wrapped to support fetching from examples and/or an in-memory cache.
	Deps       bool                       // load dependencies
	AllowEmpty bool                       // don't return error when no packages are loaded
	Fset       *token.FileSet             // external fset to help with pretty errors
	Out        io.Writer                  // used to print info
	Test       bool                       // load test dependencies
}

// XXX: get from ssot
var injectedTestingLibs = []string{"encoding/json", "fmt", "internal/os_test", "os"}

func IsInjectedTestingStdlib(pkgPath string) bool {
	return slices.Contains(injectedTestingLibs, pkgPath)
}

func (conf *LoadConfig) applyDefaults() error {
	if conf.Out == nil {
		conf.Out = io.Discard
	}
	if conf.Fetcher == nil {
		conf.Fetcher = rpcpkgfetcher.New(nil)
	}
	if conf.Fset == nil {
		conf.Fset = token.NewFileSet()
	}
	return nil
}

func Load(conf *LoadConfig, patterns ...string) (PkgList, error) {
	if conf == nil {
		conf = &LoadConfig{}
	}
	if err := conf.applyDefaults(); err != nil {
		return nil, err
	}

	root, err := findLoaderRootDir()
	if err != nil {
		return nil, err
	}

	// try to get root module if it exists, to resolve local packages
	// XXX: use a fetcher middleware?
	rootModule, err := gnomod.ParseAt(root)
	if err != nil {
		rootModule = nil
	}

	// this output is not present in go but could be useful since we don't follow the same rules to find root
	// fmt.Fprintf(conf.Out, "gno: loading patterns %s\n", strings.Join(patterns, ", "))
	// fmt.Fprintf(conf.Out, "gno: using %q as root\n", root)

	expanded, err := expandPatterns(root, conf.Out, patterns...)
	if err != nil {
		return nil, err
	}

	pkgs, err := readPackages(conf.Out, conf.Fetcher, expanded, nil, conf.Fset)
	if err != nil {
		return nil, err
	}

	if !conf.AllowEmpty && len(pkgs) == 0 {
		return nil, errors.New("no packages")
	}

	if !conf.Deps {
		return pkgs, nil
	}

	// resolve deps

	// mark all pattern packages for visit
	toVisit := []*Package(pkgs)

	resolvedByPkgPath := NewPackagesMap(pkgs...)
	markDepForVisit := func(pkg *Package) {
		resolvedByPkgPath.Add(pkg) // will only add if not already added
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

		// fmt.Fprintf(conf.Out, "gno: visiting deps of %q at %q\n", pkg.ImportPath, pkg.Dir)

		// don't load test deps if test flag is not set
		importKinds := []FileKind{FileKindPackageSource}
		if conf.Test {
			importKinds = append(importKinds, FileKindTest, FileKindXTest, FileKindFiletest)
		}

		for _, imp := range pkg.ImportsSpecs.Merge(importKinds...) {
			if IsInjectedTestingStdlib(imp.PkgPath) {
				continue
			}

			// check if we already resolved this dep
			if _, ok := resolvedByPkgPath[imp.PkgPath]; ok {
				continue
			}

			// fmt.Fprintf(conf.Out, "gno: resolving dep %q\n", imp.PkgPath)

			// check if this is a stdlib and load it from gnoroot if available
			// XXX: use a fetcher middleware?
			if gnolang.IsStdlib(imp.PkgPath) {
				// fmt.Fprintf(conf.Out, "gno: loading stdlib %q\n", imp.PkgPath)

				dir := filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs", filepath.FromSlash(imp.PkgPath))
				dirInfo, err := os.Stat(dir)
				if err != nil || !dirInfo.IsDir() {
					err := &Error{
						Pos: filepath.Join(filepath.FromSlash(pkg.Dir), conf.Fset.Position(imp.Spec.Pos()).String()),
						Msg: fmt.Sprintf("package %s is not in std (%s)", imp.PkgPath, dir),
					}
					pkg.Errors = append(pkg.Errors, err)
				}
				markDepForVisit(readPkgDir(conf.Out, conf.Fetcher, dir, imp.PkgPath, conf.Fset))
				continue
			}

			// check if this package is present in local context
			// XXX: use a fetcher middleware?
			if rootModule != nil && rootModule.Module != nil && strings.HasPrefix(imp.PkgPath, rootModule.Module.Mod.Path) {
				pkgSubPath := strings.TrimPrefix(imp.PkgPath, rootModule.Module.Mod.Path)
				pkgDir := filepath.Join(root, filepath.FromSlash(pkgSubPath))
				if info, err := os.Stat(pkgDir); err == nil && info.IsDir() {
					// fmt.Fprintf(conf.Out, "gno: loading local package %q at %q\n", imp.PkgPath, pkgDir)

					// XXX: check that this dir has gno pkg files
					markDepForVisit(readPkgDir(conf.Out, nil, pkgDir, imp.PkgPath, conf.Fset))
					continue
				}
			}

			// fmt.Fprintf(conf.Out, "gno: fetching %q\n", imp.PkgPath)

			dir := gnomod.PackageDir("", module.Version{Path: imp.PkgPath})
			if err := downloadPackage(conf.Out, conf.Fetcher, imp.PkgPath, dir); err != nil {
				pkg.Errors = append(pkg.Errors, &Error{
					Pos: pkg.Dir,
					Msg: fmt.Sprintf("download %q: %v", imp.PkgPath, err),
				})
				continue
			}
			markDepForVisit(readPkgDir(conf.Out, conf.Fetcher, dir, imp.PkgPath, conf.Fset))
		}

		loaded = append(loaded, pkg)
	}

	return loaded, nil
}

func findLoaderRootDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	root, err := gnomod.FindRootDir(wd)
	switch {
	case err == nil:
		return root, nil
	case errors.Is(err, gnomod.ErrGnoModNotFound):
		return wd, err
	default:
		return "", err
	}
}

func (p *Package) MemPkg() (*std.MemPackage, error) {
	// XXX: use gnolang.ReadMemPackageFromList

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
	sort.Slice(files, func(i int, j int) bool {
		return files[i].Name < files[j].Name
	})
	return &std.MemPackage{
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
