package packages

import (
	"fmt"
	"go/token"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func loadMatches(out io.Writer, fetcher pkgdownload.PackageFetcher, matches []*pkgMatch, ofs *overlayFS, fset *token.FileSet) (PkgList, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}

	pkgs := make([]*Package, 0, len(matches))
	for _, pkgMatch := range matches {
		pkg := loadSinglePkg(out, fetcher, pkgMatch.Dir, ofs, fset)
		pkg.Match = pkgMatch.Match
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func loadSinglePkg(out io.Writer, fetcher pkgdownload.PackageFetcher, pkgDir string, ofs *overlayFS, fset *token.FileSet) (pkg *Package) {
	defer func() {
		if pkg == nil {
			return
		}
		pkg.Errors = dedupErrors(pkg.Errors)
	}()

	pkg = &Package{
		Dir:          pkgDir,
		Files:        FilesMap{},
		Imports:      map[FileKind][]string{},
		ImportsSpecs: ImportsMap{},
	}

	mptype := gnolang.MPUserAll

	// get package from modcache if the dir is in it
	modCachePath := gnomod.ModCachePath()
	if strings.HasPrefix(filepath.Clean(pkg.Dir), modCachePath) {
		pkgPath, err := filepath.Rel(modCachePath, pkg.Dir)
		if err != nil {
			pkg.Errors = append(pkg.Errors, &Error{
				Pos: pkg.Dir,
				Msg: fmt.Errorf("failed to derive pkgpath from cache dir path: %w", err).Error(),
			})
			return pkg
		}
		pkg.ImportPath = pkgPath
		if err := DownloadPackageToCache(out, pkgPath, fetcher); err != nil {
			pkg.Errors = append(pkg.Errors, &Error{
				Pos: pkg.Dir,
				Msg: err.Error(),
			})
			return pkg
		}
	}

	// derive import path
	stdlibDir := filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs")
	testStdlibDir := filepath.Join(gnoenv.RootDir(), "gnovm", "tests", "stdlibs")
	switch {
	case strings.HasPrefix(pkg.Dir, stdlibDir):
		// get package path from dir
		rel, err := filepath.Rel(stdlibDir, pkg.Dir)
		if err != nil {
			// return partial package if can't find lib pkgpath
			pkg.Errors = append(pkg.Errors, fromErr(err, pkg.Dir, false)...)
			return pkg
		}
		pkg.ImportPath = filepath.ToSlash(rel)
		mptype = gnolang.MPStdlibAll
	case strings.HasPrefix(pkg.Dir, testStdlibDir):
		// get package path from dir
		rel, err := filepath.Rel(testStdlibDir, pkg.Dir)
		if err != nil {
			// return partial package if can't find lib pkgpath
			pkg.Errors = append(pkg.Errors, fromErr(err, pkg.Dir, false)...)
			return pkg
		}
		pkg.ImportPath = filepath.ToSlash(rel)
		mptype = gnolang.MPStdlibAll
	default:
		// attempt to load gnomod.toml if package is not stdlib
		// get import path and flags from gnomod
		modfpath := filepath.Join(pkg.Dir, "gnomod.toml")
		bz, err := fs.ReadFile(ofs, modfpath)
		if err != nil {
			// return partial package if invalid gnomod
			pkg.Errors = append(pkg.Errors, &Error{Msg: "missing gnomod", Pos: pkg.Dir})
			return pkg
		}
		mod, err := gnomod.ParseBytes(modfpath, bz)
		if err != nil {
			// return partial package if invalid gnomod
			pkg.Errors = append(pkg.Errors, fromErr(err, pkg.Dir, false)...)
			return pkg
		}
		pkg.Ignore = mod.Ignore
		pkg.ImportPath = mod.Module
	}

	mpkg, err := func() (mpkg *std.MemPackage, err error) {
		// need to recover since ReadMemPackage is panicking again
		// XXX: use a version of ReadMemPackage that doesn't panic
		defer func() {
			pret := recover()
			switch cret := pret.(type) {
			case nil:
				// do nothing
			case error:
				err = fmt.Errorf("read mempackage: %w", cret)
			default:
				err = fmt.Errorf("read mempackage: %v", cret)
			}
		}()
		return gnolang.ReadFSMemPackage(ofs, pkg.Dir, pkg.ImportPath, mptype)
	}()
	if err != nil {
		pkg.Errors = append(pkg.Errors, fromErr(err, pkg.Dir, true)...)
		return pkg
	}

	pkg.Name = mpkg.Name

	// XXX: gnowork.toml files are included in mempkgs, should we ignore them?

	// XXX: files are ignored if ReadMemPackage fails,
	// since ReadMemPackage is restrictive we should probably consider files another way

	for _, file := range mpkg.Files {
		fileKind := GetFileKind(file.Name, file.Body, fset)
		pkg.Files[fileKind] = append(pkg.Files[fileKind], file.Name)
	}

	imps, err := Imports(mpkg, fset)
	if err != nil {
		pkg.Errors = append(pkg.Errors, fromErr(err, pkg.Dir, true)...)
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
