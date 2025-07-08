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
)

func loadMatches(out io.Writer, fetcher pkgdownload.PackageFetcher, matches []*pkgMatch, known PkgList, fset *token.FileSet) (PkgList, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}

	pkgs := make([]*Package, 0, len(matches))
	for _, pkgMatch := range matches {
		if known.GetByDir(pkgMatch.Dir) != nil {
			continue
		}

		pkg := loadSinglePkg(out, fetcher, pkgMatch.Dir, fset)
		pkg.Errors = dedupErrors(pkg.Errors)
		pkg.Match = pkgMatch.Match
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func loadSinglePkg(out io.Writer, fetcher pkgdownload.PackageFetcher, pkgDir string, fset *token.FileSet) *Package {
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

	stdlibDir := filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs")
	if strings.HasPrefix(pkg.Dir, stdlibDir) {
		// get package path from dir
		rel, err := filepath.Rel(stdlibDir, pkg.Dir)
		if err != nil {
			// return partial package if can't find lib pkgpath
			pkg.Errors = append(pkg.Errors, FromErr(err, fset, pkg.Dir, false)...)
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
			return pkg
		}
		pkg.Ignore = mod.Ignore
		pkg.ImportPath = mod.Module
	}

	mpkg, err := gnolang.ReadMemPackage(pkg.Dir, pkg.ImportPath, gnolang.MPAnyAll)
	if err != nil {
		pkg.Errors = append(pkg.Errors, FromErr(err, fset, pkg.Dir, true)...)
		return pkg
	}

	pkg.Name = mpkg.Name

	// XXX: gnowork.toml files are included in mempkgs, should we ignore them?

	// XXX: files are ignored if ReadMemPackage fails,
	// since ReadMemPackage is restrictive we should probably consider files another way

	for _, file := range mpkg.Files {
		fpath := filepath.Join(pkg.Dir, file.Name)

		fileKind, err := GetFileKind(file.Name, file.Body, fset)
		if err != nil {
			pkg.Errors = append(pkg.Errors, FromErr(err, fset, fpath, false)...)
			continue
		}
		pkg.Files[fileKind] = append(pkg.Files[fileKind], file.Name)
	}

	imps, err := Imports(mpkg, fset)
	if err != nil {
		pkg.Errors = append(pkg.Errors, FromErr(err, fset, pkg.Dir, true)...)
		return pkg
	}
	pkg.ImportsSpecs = imps
	pkg.Imports = imps.ToStrings()

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
