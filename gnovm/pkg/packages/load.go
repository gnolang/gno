package packages

import (
	"errors"
	"fmt"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/gnovm/tests/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type LoadConfig struct {
	Fetcher             pkgdownload.PackageFetcher // package fetcher used to load dependencies not present in patterns. Could be wrapped to support fetching from examples and/or an in-memory cache.
	Deps                bool                       // load dependencies
	AllowEmpty          bool                       // don't return error when no packages are loaded
	Fset                *token.FileSet             // external fset to help with pretty errors
	Out                 io.Writer                  // used to print info
	Test                bool                       // load test dependencies
	WorkspaceRoot       string                     // disable workspace root detection if set
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
	return nil
}

func Load(conf LoadConfig, patterns ...string) (PkgList, error) {
	if err := conf.applyDefaults(); err != nil {
		return nil, err
	}

	// XXX: allow loading only stdlibs without a workspace (like go allow loading stdlibs without a go.mod)

	if conf.WorkspaceRoot == "" {
		root, err := findLoaderRootDir()
		if err != nil {
			return nil, err
		}
		conf.WorkspaceRoot = root
	} else if !filepath.IsAbs(conf.WorkspaceRoot) {
		var err error
		conf.WorkspaceRoot, err = filepath.Abs(conf.WorkspaceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute workspace root: %w", err)
		}
	}

	// fmt.Fprintf(conf.Out, "gno: workspace root is %q\n", conf.WorkspaceRoot)

	// sanity assert
	if !filepath.IsAbs(conf.WorkspaceRoot) {
		panic(fmt.Errorf("workspace root should be absolute at this point, got %q", conf.WorkspaceRoot))
	}

	localDeps := discoverPkgsForLocalDeps(append([]string{conf.WorkspaceRoot}, conf.ExtraWorkspaceRoots...))

	expanded, err := expandPatterns(conf.WorkspaceRoot, conf.Out, patterns...)
	if err != nil {
		return nil, err
	}

	pkgs, err := loadMatches(conf.Out, conf.Fetcher, expanded, nil, conf.Fset)
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

		// fmt.Fprintf(conf.Out, "gno: visiting deps of %q at %q\n", pkg.ImportPath, pkg.Dir)

		// don't load test deps if test flag is not set
		importKinds := []FileKind{FileKindPackageSource}
		if conf.Test {
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

func findLoaderRootDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	root, err := FindRootDir(wd)
	switch {
	case err == nil:
		return root, nil
	default:
		return "", err
	}
}

// ErrGnoworkNotFound is returned by [FindRootDir] when, even after traversing
// up to the root directory, a gno.mod file could not be found.
var ErrGnoworkNotFound = errors.New("gnowork.toml file not found in current or any parent directory")

// FindRootDir determines the root directory of the project.
// The given path must be absolute.
func FindRootDir(absPath string) (string, error) {
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

func (p *Package) MemPkg() (*std.MemPackage, error) {
	// XXX: use gnolang.ReadMemPackageFromList

	files := []*std.MemFile{}
	for _, cat := range p.Files {
		for _, f := range cat {
			body, err := os.ReadFile(filepath.Join(p.Dir, f))
			if err != nil {
				return nil, err
			}
			files = append(files, &std.MemFile{
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

func discoverPkgsForLocalDeps(roots []string) map[string]string {
	// we swallow errors in this routine as we want the most packages we can get

	byPkgPath := make(map[string]string)
	byDir := make(map[string]string)

	for _, root := range roots {
		_ = fs.WalkDir(os.DirFS(root), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
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

			case "gnowork.toml":
				// stop if sub-workspace

				if dir != root {
					// if we found a pkg in this dir, ignore it
					pkgPath, ok := byDir[dir]
					if ok {
						delete(byDir, dir)
						delete(byPkgPath, pkgPath)
					}

					// skip subtree
					return fs.SkipDir
				}
			}

			return nil
		})
	}

	return byPkgPath
}
