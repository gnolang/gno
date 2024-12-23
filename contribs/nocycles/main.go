package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
)

func main() {
	// set to true to dump the final pkg list
	verbose := true

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
		importsMap, err := packages.Imports(memPkg, nil)
		if err != nil {
			panic(fmt.Errorf("read %q: %w", memPkg.Name, err))
		}
		pl = append(pl, gnomod.Pkg{
			Dir:     "",
			Name:    lib,
			Imports: fileImportsToStrings(importsMap.Merge(packages.FileKindPackageSource, packages.FileKindTest)),
		})
		xTestImports := fileImportsToStrings(importsMap.Merge(packages.FileKindXTest))
		if len(xTestImports) > 0 {
			pl = append(pl, gnomod.Pkg{
				Dir:     "",
				Name:    "_xtest_" + lib,
				Imports: xTestImports,
			})
		}
	}

	// load all examples
	examples, err := gnomod.ListPkgs(filepath.Join(gnoRoot, "examples"))
	if err != nil {
		panic(fmt.Errorf("load examples: %w", err))
	}
	for _, example := range examples {
		if example.Draft {
			continue
		}

		pkgPath := example.Name
		memPkg := gnolang.MustReadMemPackage(example.Dir, example.Name)

		if memPkg.IsEmpty() {
			continue
		}

		importsMap, err := packages.Imports(memPkg, nil)
		if err != nil {
			panic(fmt.Errorf("read %q: %w", pkgPath, err))
		}

		pl = append(pl, gnomod.Pkg{
			Dir:     example.Dir,
			Name:    pkgPath,
			Imports: fileImportsToStrings(importsMap.Merge(packages.FileKindPackageSource, packages.FileKindTest)),
		})

		xTestImports := fileImportsToStrings(importsMap.Merge(packages.FileKindXTest))
		if len(xTestImports) > 0 {
			pl = append(pl, gnomod.Pkg{
				Dir:     example.Dir,
				Name:    "_xtest_" + pkgPath,
				Imports: xTestImports,
			})
		}
	}

	// detect import cycles
	if _, err := pl.Sort(); err != nil {
		panic(err)
	}

	if verbose {
		for _, p := range pl {
			fmt.Println(p.Name)
		}
	}
}

func fileImportsToStrings(fis []packages.FileImport) []string {
	res := make([]string, len(fis))
	for i, fi := range fis {
		res[i] = fi.PkgPath
	}
	return res
}
