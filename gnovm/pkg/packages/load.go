package packages

import (
	"errors"
	"fmt"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/gnovm/tests/stdlibs"
)

type LoadConfig struct {
	Fetcher             pkgdownload.PackageFetcher // package fetcher used to load dependencies not present in patterns. Could be wrapped to support fetching from examples and/or an in-memory cache.
	Deps                bool                       // load dependencies
	AllowEmpty          bool                       // don't return error when no packages are loaded
	Fset                *token.FileSet             // external fset to help with pretty errors
	Out                 io.Writer                  // used to print info
	Test                bool                       // load test dependencies
	GnoRoot             string                     // used to override GNOROOT
	ExtraWorkspaceRoots []string                   // extra workspaces root used to find dependencies
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
	if conf.GnoRoot == "" {
		conf.GnoRoot = gnoenv.RootDir()
	}
	return nil
}

func Load(conf LoadConfig, patterns ...string) (PkgList, error) {
	if err := conf.applyDefaults(); err != nil {
		return nil, err
	}

	// XXX: allow loading only stdlibs without a workspace (like go allow loading stdlibs without a go.mod)

	loaderCtx, err := findLoaderContext()
	if err != nil {
		return nil, err
	}

	// sanity assert
	if !filepath.IsAbs(loaderCtx.Root) {
		panic(fmt.Errorf("context root should be absolute at this point, got %q", loaderCtx.Root))
	}

	expanded, err := expandPatterns(conf.GnoRoot, loaderCtx, conf.Out, patterns...)
	if err != nil {
		return nil, err
	}

	pkgs, err := loadMatches(conf.Out, conf.Fetcher, expanded, conf.Fset)
	if err != nil {
		return nil, err
	}

	if !conf.AllowEmpty && len(pkgs) == 0 {
		return nil, errors.New("no packages")
	}

	if !conf.Deps {
		return pkgs, nil
	}

	// load deps

	localDeps := discoverPkgsForLocalDeps(conf, loaderCtx)

	// mark all pattern packages for visit
	toVisit := []*Package(pkgs)

	resolvedByPkgPath := map[string]struct{}{}
	markDepForVisit := func(pkg *Package) {
		resolvedByPkgPath[pkg.ImportPath] = struct{}{} // will only add if not already added
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

		// load tests deps if test flag is set and the package is not a dep
		importKinds := []FileKind{FileKindPackageSource}
		if conf.Test && len(pkg.Match) != 0 {
			importKinds = append(importKinds, FileKindTest, FileKindXTest, FileKindFiletest)
		}

		for _, imp := range pkg.ImportsSpecs.Merge(importKinds...) {
			// ignore injected testing stdlibs
			if stdlibs.HasNativePkg(imp.PkgPath) {
				continue
			}

			// check if we already resolved this dep
			if _, ok := resolvedByPkgPath[imp.PkgPath]; ok {
				continue
			}

			// check if this is a stdlib and load it from gnoroot if available
			// XXX: use a fetcher middleware?
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
				markDepForVisit(loadSinglePkg(conf.Out, conf.Fetcher, dir, conf.Fset))
				continue
			}

			// check if this package is present in current workspace or extra workspace roots
			if dir, ok := localDeps[imp.PkgPath]; ok {
				markDepForVisit(loadSinglePkg(conf.Out, nil, dir, conf.Fset))
				continue
			}

			// attempt to download package
			dir := PackageDir(imp.PkgPath)
			markDepForVisit(loadSinglePkg(conf.Out, conf.Fetcher, dir, conf.Fset))
		}

		loaded = append(loaded, pkg)
	}

	return loaded, nil
}

type loaderContext struct {
	Root        string
	IsWorkspace bool
}

func findLoaderContext() (*loaderContext, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	{
		dir, err := findWorkspaceRootDir(wd)
		switch {
		case err == nil:
			return &loaderContext{Root: dir, IsWorkspace: true}, nil
		case errors.Is(err, ErrGnoworkNotFound):
			// continue
		default:
			return nil, err
		}
	}

	gnomodPath := filepath.Join(wd, "gnomod.toml")
	_, err = os.Stat(gnomodPath)
	switch {
	case err == nil:
		return &loaderContext{Root: wd}, nil
	case os.IsNotExist(err):
		return nil, ErrGnoContextNotFound
	default:
		return nil, err
	}
}

// ErrGnoworkNotFound is returned by [findRootDir] when, even after traversing
// up to the root directory, a gnowork.toml file could not be found.
var ErrGnoworkNotFound = errors.New("gnowork.toml file not found in current or any parent directory")

var ErrGnoContextNotFound = errors.New("gnowork.toml file not found in current or any parent directory and gnomod.toml doesn't exists in current directory")

// findWorkspaceRootDir determines the root directory of the project.
// The given path must be absolute.
func findWorkspaceRootDir(absPath string) (string, error) {
	if !filepath.IsAbs(absPath) {
		return "", errors.New("requires absolute path")
	}

	root := filepath.VolumeName(absPath) + string(filepath.Separator)

	for absPath != root {
		modPath := filepath.Join(absPath, "gnowork.toml")
		_, err := os.Stat(modPath)
		if errors.Is(err, os.ErrNotExist) {
			absPath = filepath.Dir(absPath)
			continue
		}
		if err != nil {
			return "", err
		}
		return absPath, nil
	}

	return "", ErrGnoworkNotFound
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

func discoverPkgsForLocalDeps(conf LoadConfig, loaderCtx *loaderContext) map[string]string {
	// we swallow errors in this routine as we want the most packages we can get

	roots := []string{}
	if loaderCtx.IsWorkspace {
		roots = append(roots, loaderCtx.Root)
	}
	roots = append(roots, conf.ExtraWorkspaceRoots...)

	byPkgPath := make(map[string]string)
	byDir := make(map[string]string)

	for _, root := range roots {
		root = filepath.Clean(root)

		_ = fs.WalkDir(os.DirFS(root), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				dir := filepath.Join(root, path)
				if dir == root {
					return nil
				}
				subwork := filepath.Join(dir, "gnowork.toml")
				_, err := os.Stat(subwork)
				switch {
				case os.IsNotExist(err):
					// not a sub-workspace, continue walking
					return nil
				case err != nil:
					return fmt.Errorf("check that dir is not a subworkspace: %w", err)
				default:
					return fs.SkipDir
				}
			}

			dir, base := filepath.Split(path)
			dir = filepath.Join(root, dir)

			switch base {
			case "gnomod.toml", "gno.mod":
				// XXX: maybe also match *.gno

				// skip this file if we already found a package in this dir
				if _, ok := byDir[dir]; ok {
					return nil
				}

				// find pkg path
				gm, err := gnomod.ParseDir(dir)
				if err != nil {
					// XXX: maybe store errors by dir to not silently ignore packages with invalid gnomod
					return nil
				}
				pkgPath := gm.Module

				// skip this file if we already found this package
				if _, ok := byPkgPath[pkgPath]; ok {
					return nil
				}

				// store ref
				byPkgPath[pkgPath] = dir
				byDir[dir] = pkgPath

				return nil
			}

			return nil
		})
	}

	return byPkgPath
}
