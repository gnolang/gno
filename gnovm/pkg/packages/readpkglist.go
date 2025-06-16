package packages

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// ========================================
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
		for _, fname := range []string{"gnomod.toml", "gno.mod"} {
			modPath := filepath.Join(path, fname)
			data, err := os.ReadFile(modPath)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return err
			}

			mod, err := gnomod.ParseBytes(modPath, data)
			if err != nil {
				return fmt.Errorf("parse: %w", err)
			}
			mod.Sanitize()
			if err := mod.Validate(); err != nil {
				return fmt.Errorf("failed to validate gnomod.toml in %s: %w", modPath, err)
			}

			pkg, err := gnolang.ReadMemPackage(path, mod.Module)
			if err != nil {
				// ignore package files on error
				pkg = &std.MemPackage{}
			}

			importsMap, err := Imports(pkg, nil)
			if err != nil {
				// ignore imports on error
				importsMap = nil
			}
			importsRaw := importsMap.Merge(FileKindPackageSource, FileKindTest, FileKindXTest)

			imports := make([]string, 0, len(importsRaw))
			for _, imp := range importsRaw {
				// remove self and standard libraries from imports
				if imp.PkgPath != mod.Module &&
					!gnolang.IsStdlib(imp.PkgPath) {
					imports = append(imports, imp.PkgPath)
				}
			}

			pkgs = append(pkgs, gnomod.Pkg{
				Dir:     path,
				Name:    mod.Module,
				Draft:   mod.Draft,
				Imports: imports,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}
