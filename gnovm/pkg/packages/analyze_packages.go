package packages

import (
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

func readPackages(matches []*pkgMatch) []*Package {
	pkgs := make([]*Package, 0, len(matches))
	for _, pkgMatch := range matches {
		pkg := readPkg(pkgMatch.Dir, "")
		pkg.Match = pkgMatch.Match
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func readPkg(pkgDir string, importPath string) *Package {
	pkg := &Package{
		Dir:        pkgDir,
		Files:      make(FilesMap),
		ImportPath: importPath,
	}

	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		pkg.Error = errors.Join(pkg.Error, err)
		return pkg
	}

	fset := token.NewFileSet()

	mempkg := gnovm.MemPackage{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		base := entry.Name()
		fpath := filepath.Join(pkgDir, base)

		if !strings.HasSuffix(base, ".gno") {
			continue
		}

		bodyBytes, err := os.ReadFile(fpath)
		if err != nil {
			pkg.Error = errors.Join(pkg.Error, err)
			continue
		}
		body := string(bodyBytes)

		fileKind, err := GetFileKind(base, body, fset)
		if err != nil {
			pkg.Error = errors.Join(pkg.Error, err)
			continue
		}

		// ignore files with invalid package clause
		_, err = parser.ParseFile(fset, fpath, nil, parser.PackageClauseOnly)
		if err != nil {
			pkg.Error = errors.Join(pkg.Error, err)
			continue
		}

		mempkg.Files = append(mempkg.Files, &gnovm.MemFile{Name: base, Body: body})
		pkg.Files[fileKind] = append(pkg.Files[fileKind], base)
	}

	pkg.Imports, err = Imports(&mempkg, fset)
	if err != nil {
		pkg.Error = errors.Join(pkg.Error, err)
	}

	// we use the ReadMemPkg utils because we want name resolution like the vm
	nameFiles := pkg.Files.Merge(FileKindPackageSource, FileKindTest, FileKindXTest)
	absFiles := make([]string, 0, len(nameFiles))
	for _, f := range nameFiles {
		absFiles = append(absFiles, filepath.Join(pkg.Dir, f))
	}
	minMempkg, err := gnolang.ReadMemPackageFromList(absFiles, "")
	if err != nil {
		pkg.Error = errors.Join(pkg.Error, err)
	} else {
		pkg.Name = minMempkg.Name
	}

	// TODO: check if stdlib

	pkg.Root, err = gnomod.FindRootDir(pkgDir)
	if errors.Is(err, gnomod.ErrGnoModNotFound) {
		return pkg
	}
	if err != nil {
		pkg.Error = errors.Join(pkg.Error, err)
		return pkg
	}

	mod, err := gnomod.ParseGnoMod(filepath.Join(pkg.Root, "gno.mod"))
	if err != nil {
		pkg.Error = errors.Join(pkg.Error, err)
		return pkg
	}

	pkg.Draft = mod.Draft

	if mod.Module == nil {
		return pkg
	}

	pkg.ModPath = mod.Module.Mod.Path
	subPath, err := filepath.Rel(pkg.Root, pkgDir)
	if err != nil {
		pkg.Error = errors.Join(pkg.Error, err)
		return pkg
	}

	pkg.ImportPath = path.Join(pkg.ModPath, filepath.ToSlash(subPath))

	return pkg
}
