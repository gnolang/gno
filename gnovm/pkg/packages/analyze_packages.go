package packages

import (
	"errors"
	"fmt"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

func readPackages(matches []*pkgMatch, fset *token.FileSet) ([]*Package, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}
	pkgs := make([]*Package, 0, len(matches))
	for _, pkgMatch := range matches {
		var pkg *Package
		var err error
		if pkgMatch.Dir == "command-line-arguments" {
			pkg, err = readCLAPkg(pkgMatch.Match, fset)
			if err != nil {
				return nil, err
			}
		} else {
			pkg = readPkgDir(pkgMatch.Dir, "", fset)
		}
		pkg.Match = pkgMatch.Match
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func readCLAPkg(patterns []string, fset *token.FileSet) (*Package, error) {
	pkg := &Package{
		ImportPath: "command-line-arguments",
		Files:      make(FilesMap),
		Imports:    make(ImportsMap),
	}
	var err error

	files := []string{}
	for _, match := range patterns {
		dir, base := filepath.Split(match)
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		if pkg.Dir == "" {
			pkg.Dir = dir
		} else if dir != pkg.Dir {
			return nil, fmt.Errorf("named files must all be in one directory; have %s and %s", pkg.Dir, dir)
		}

		files = append(files, base)
	}

	return readPkgFiles(pkg, files, fset), nil
}

func readPkgDir(pkgDir string, importPath string, fset *token.FileSet) *Package {
	pkg := &Package{
		Dir:        pkgDir,
		Files:      make(FilesMap),
		Imports:    make(ImportsMap),
		ImportPath: importPath,
	}

	if pkg.ImportPath == "" {
		stdlibsPath := filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs")
		if strings.HasPrefix(filepath.Clean(pkg.Dir), stdlibsPath) {
			libPath, err := filepath.Rel(stdlibsPath, pkg.Dir)
			if err != nil {
				pkg.Errors = append(pkg.Errors, err)
				return pkg
			}
			pkg.ImportPath = libPath
		}
	}

	files := []string{}
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		pkg.Errors = append(pkg.Errors, err)
		return pkg
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		base := entry.Name()

		if !strings.HasSuffix(base, ".gno") {
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

	mempkg := gnovm.MemPackage{}

	for _, base := range files {
		fpath := filepath.Join(pkg.Dir, base)

		bodyBytes, err := os.ReadFile(fpath)
		if err != nil {
			pkg.Errors = append(pkg.Errors, err)
			continue
		}
		body := string(bodyBytes)

		fileKind, err := GetFileKind(base, body, fset)
		if err != nil {
			pkg.Errors = append(pkg.Errors, err)
			continue
		}

		mempkg.Files = append(mempkg.Files, &gnovm.MemFile{Name: base, Body: body})
		pkg.Files[fileKind] = append(pkg.Files[fileKind], base)
	}

	var err error
	pkg.Imports, err = Imports(&mempkg, fset)
	if err != nil {
		pkg.Errors = append(pkg.Errors, err)
	}

	// we use the ReadMemPkg utils to get the package name because we want name resolution like the vm
	nameFiles := pkg.Files.Merge(FileKindPackageSource, FileKindTest, FileKindXTest)
	absFiles := make([]string, 0, len(nameFiles))
	for _, f := range nameFiles {
		absFiles = append(absFiles, filepath.Join(pkg.Dir, f))
	}
	minMempkg, err := gnolang.ReadMemPackageFromList(absFiles, "")
	if err != nil {
		pkg.Errors = append(pkg.Errors, err)
	} else {
		pkg.Name = minMempkg.Name
	}

	// TODO: check if stdlib

	if pkg.ImportPath == "command-line-arguments" {
		return pkg
	}

	pkg.Root, err = gnomod.FindRootDir(pkg.Dir)
	if errors.Is(err, gnomod.ErrGnoModNotFound) {
		return pkg
	}
	if err != nil {
		pkg.Errors = append(pkg.Errors, err)
		return pkg
	}

	mod, err := gnomod.ParseGnoMod(filepath.Join(pkg.Root, "gno.mod"))
	if err != nil {
		pkg.Errors = append(pkg.Errors, err)
		return pkg
	}

	pkg.Draft = mod.Draft

	if mod.Module == nil {
		return pkg
	}

	pkg.ModPath = mod.Module.Mod.Path
	subPath, err := filepath.Rel(pkg.Root, pkg.Dir)
	if err != nil {
		pkg.Errors = append(pkg.Errors, err)
		return pkg
	}

	pkg.ImportPath = path.Join(pkg.ModPath, filepath.ToSlash(subPath))

	return pkg
}
