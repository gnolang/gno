package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
)

func main() {
	// find stdlibs
	libs := []string{}
	gnoRoot := gnoenv.RootDir()
	stdlibsDir := filepath.Join(gnoRoot, "gnovm", "stdlibs")
	fs.WalkDir(os.DirFS(stdlibsDir), ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}
		if path == "." {
			return nil
		}
		libs = append(libs, path)
		return nil
	})

	// read stdlibs
	pl := gnomod.PkgList{}
	for _, lib := range libs {
		memPkg := gnolang.MustReadMemPackage(filepath.Join(stdlibsDir, lib), lib)
		pkg, xpkg, err := splitMemPackage(memPkg)
		if err != nil {
			panic(fmt.Errorf("split %q: %w", lib, err))
		}
		{
			imports, err := packages.Imports(pkg)
			if err != nil {
				panic(fmt.Errorf("read %q: %w", lib, err))
			}
			pl = append(pl, gnomod.Pkg{
				Dir:     "",
				Name:    lib,
				Imports: imports,
			})
		}
		if !xpkg.IsEmpty() {
			imports, err := packages.Imports(xpkg)
			if err != nil {
				panic(fmt.Errorf("read %q: %w", lib, err))
			}
			pl = append(pl, gnomod.Pkg{
				Dir:     "",
				Name:    lib + "_test",
				Imports: imports,
			})
		}
	}

	// detect import cycles
	_, err := pl.Sort()
	if err != nil {
		panic(err)
	}
}

func splitMemPackage(pkg *gnovm.MemPackage) (*gnovm.MemPackage, *gnovm.MemPackage, error) {
	corePkg := gnovm.MemPackage{
		Name: pkg.Name,
		Path: pkg.Path,
	}
	xtestPkg := gnovm.MemPackage{
		Name: pkg.Name + "_test",
		Path: pkg.Path,
	}

	for _, file := range pkg.Files {
		if !strings.HasSuffix(file.Name, "_test.gno") {
			corePkg.Files = append(corePkg.Files, file)
			continue
		}
		pkgName, err := packages.FilePackageName(file.Name, file.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("get package name in file %q: %w", file.Name, err)
		}
		if !strings.HasSuffix(pkgName, "_test") {
			corePkg.Files = append(corePkg.Files, file)
			continue
		}
		xtestPkg.Files = append(xtestPkg.Files, file)
	}

	return &corePkg, &xtestPkg, nil
}
