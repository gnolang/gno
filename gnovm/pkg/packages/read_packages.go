package packages

import (
	"fmt"
	"go/token"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func readPackages(out io.Writer, fetcher pkgdownload.PackageFetcher, matches []*pkgMatch, known PkgList, fset *token.FileSet) (PkgList, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}
	pkgs := make([]*Package, 0, len(matches))
	for _, pkgMatch := range matches {
		var pkg *Package
		if known.GetByDir(pkgMatch.Dir) != nil {
			continue
		} else {
			pkg = readPkgDir(out, fetcher, pkgMatch.Dir, fset)
		}
		pkg.Match = pkgMatch.Match
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

// XXX: bad name since it might download the package
func readPkgDir(out io.Writer, fetcher pkgdownload.PackageFetcher, pkgDir string, fset *token.FileSet) *Package {
	pkg := &Package{
		Dir:          pkgDir,
		Files:        FilesMap{},
		Imports:      map[FileKind][]string{},
		ImportsSpecs: ImportsMap{},
	}

	stdlibsPath := filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs")
	if strings.HasPrefix(filepath.Clean(pkg.Dir), stdlibsPath) {
		libPath, err := filepath.Rel(stdlibsPath, pkg.Dir)
		if err != nil {
			pkg.Errors = append(pkg.Errors, &Error{
				Pos: pkg.Dir,
				Msg: err.Error(),
			})
			return pkg
		}
		pkg.ImportPath = filepath.ToSlash(libPath)
	}

	// FIXME: concurrency + don't overwrite
	modCachePath := gnomod.ModCachePath()
	if strings.HasPrefix(filepath.Clean(pkg.Dir), modCachePath) {
		pkgPath, err := filepath.Rel(modCachePath, pkg.Dir)
		if err != nil {
			pkg.Errors = append(pkg.Errors, &Error{
				Pos: pkg.Dir,
				Msg: fmt.Errorf("pkgpath from cache dir: %w", err).Error(),
			})
			return pkg
		}
		pkg.ImportPath = path.Clean(filepath.ToSlash(pkgPath))
		_, err = os.Stat(pkg.Dir)
		if err != nil {
			if os.IsNotExist(err) {
				if err := DownloadPackage(out, pkgPath, pkg.Dir, fetcher); err != nil {
					pkg.Errors = append(pkg.Errors, &Error{
						Pos: pkg.Dir,
						Msg: err.Error(),
					})
					return pkg
				}
			} else {
				pkg.Errors = append(pkg.Errors, &Error{
					Pos: pkg.Dir,
					Msg: fmt.Errorf("stat: %w", err).Error(),
				})
				return pkg
			}
		}
	}

	files := []string{}
	entries, err := os.ReadDir(pkg.Dir)
	if err != nil {
		pkg.Errors = append(pkg.Errors, &Error{
			Pos: pkg.Dir,
			Msg: err.Error(),
		})
		return pkg
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		base := entry.Name()

		if !strings.HasSuffix(base, ".gno") && base != "LICENSE" && base != "README.md" {
			continue
		}

		files = append(files, base)
	}

	return readPkgFiles(pkg, files, fset)
}

func readPkgFiles(pkg *Package, files []string, fset *token.FileSet) *Package {
	if fset == nil {
		fset = token.NewFileSet()
	}

	mempkg := std.MemPackage{}

	for _, base := range files {
		fpath := filepath.Join(pkg.Dir, base)

		bodyBytes, err := os.ReadFile(fpath)
		if err != nil {
			pkg.Errors = append(pkg.Errors, &Error{
				Pos: fpath,
				Msg: err.Error(),
			})
			continue
		}
		body := string(bodyBytes)

		fileKind, err := GetFileKind(base, body, fset)
		if err != nil {
			pkg.Errors = append(pkg.Errors, FromErr(err, fset, fpath, false)...)
			continue
		}

		mempkg.Files = append(mempkg.Files, &std.MemFile{Name: base, Body: body})
		pkg.Files[fileKind] = append(pkg.Files[fileKind], base)
	}

	stdlibDir := filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs")
	if strings.HasPrefix(pkg.Dir, stdlibDir) {
		// get package path from dir
		rel, err := filepath.Rel(stdlibDir, pkg.Dir)
		if err != nil {
			// return partial package if can't find lib pkgpath
			pkg.Errors = append(pkg.Errors, FromErr(err, fset, pkg.Dir, false)...)
			pkg.Errors = dedupErrors(pkg.Errors)
			return pkg
		}
		pkg.ImportPath = filepath.ToSlash(rel)
	} else {
		// attempt to load gnomod.toml if package is not stdlib
		// get import path and flags from gnomod
		mod, fname, err := gnomod.ParseDir(pkg.Dir)
		if err != nil {
			// return partial package if invalid gnomod
			pkg.Errors = append(pkg.Errors, FromErr(err, fset, filepath.Join(pkg.Dir, fname), false)...)
			pkg.Errors = dedupErrors(pkg.Errors)
			return pkg
		}
		pkg.Ignore = mod.Ignore
		pkg.ImportPath = mod.Module
	}

	// XXX: fset
	minMempkg, err := gnolang.ReadMemPackage(pkg.Dir, pkg.ImportPath, gnolang.MPAnyAll)
	if err != nil {
		pkg.Errors = append(pkg.Errors, FromErr(err, fset, pkg.Dir, true)...)
	} else {
		pkg.Name = minMempkg.Name
	}

	pkg.ImportsSpecs, err = Imports(&mempkg, fset)
	if err != nil {
		pkg.Errors = append(pkg.Errors, FromErr(err, fset, pkg.Dir, true)...)
	}
	pkg.Imports = pkg.ImportsSpecs.ToStrings()

	pkg.Errors = dedupErrors(pkg.Errors)
	return pkg
}

func dedupErrors(s []*Error) []*Error {
	seen := map[Error]struct{}{}
	res := []*Error{}
	for _, elem := range s {
		if elem == nil {
			continue
		}
		if _, ok := seen[*elem]; ok {
			continue
		}
		res = append(res, elem)
		seen[*elem] = struct{}{}
	}
	return res
}
