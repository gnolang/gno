package gnolang

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// ReadPkgListFromDir() lists all gno packages in the given dir directory.
func ReadPkgListFromDir(dir string) (gnomod.PkgList, error) {
	var pkgs []gnomod.Pkg

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		gmfPath := filepath.Join(path, "gno.mod")
		data, err := os.ReadFile(gmfPath)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}

		gmf, err := gnomod.ParseBytes(gmfPath, data)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		gmf.Sanitize()
		if err := gmf.Validate(); err != nil {
			return fmt.Errorf("failed to validate gno.mod in %s: %w", gmfPath, err)
		}

		pkg, err := ReadMemPackage(path, gmf.Module.Mod.Path)
		if err != nil {
			// ignore package files on error
			pkg = &std.MemPackage{}
		}

		importsMap, err := packages.Imports(pkg, nil)
		if err != nil {
			// ignore imports on error
			importsMap = nil
		}
		importsRaw := importsMap.Merge(packages.FileKindPackageSource, packages.FileKindTest, packages.FileKindXTest)

		imports := make([]string, 0, len(importsRaw))
		for _, imp := range importsRaw {
			// remove self and standard libraries from imports
			if imp.PkgPath != gmf.Module.Mod.Path &&
				!IsStdlib(imp.PkgPath) {
				imports = append(imports, imp.PkgPath)
			}
		}

		pkgs = append(pkgs, gnomod.Pkg{
			Dir:     path,
			Name:    gmf.Module.Mod.Path,
			Draft:   gmf.Draft,
			Imports: imports,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}
