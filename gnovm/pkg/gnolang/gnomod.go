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

const (
	// gno.mod files assumed in testing/default contexts.
	GnoModLatest  = `go 0.9` // when gno.mod is missing for stdlibs.
	GnoModTesting = `go 0.9` // when gno.mod is missing while testing.
	GnoModDefault = `go 0.0` // when gno.mod is missing in general.
	GnoVerDefault = `0.0`    // when gno version isn't specified.
)

// ========================================
// Parses and checks the gno.mod file from mpkg.
// To generate default ones, use:
// gnomod.ParseBytes(GnoModDefault)
//
// Results:
//   - mod: the gno.mod file, or nil if not found.
//   - err: wrapped error, or nil if file not found.
func ParseCheckGnoMod(mpkg *std.MemPackage) (mod *gnomod.File, err error) {
	if IsStdlib(mpkg.Path) {
		// stdlib/extern packages are assumed up to date.
		mod, _ = gnomod.ParseBytes("<stdlibs>/gno.mod", []byte(GnoModLatest))
	} else if mpkg.GetFile("gno.mod") == nil {
		// gno.mod doesn't exist.
		return nil, nil
	} else if mod, err = gnomod.ParseMemPackage(mpkg); err != nil {
		// error parsing gno.mod.
		err = fmt.Errorf("%s/gno.mod: parse error %q", mpkg.Path, err)
	} else if mod.Gno == nil {
		// 'gno 0.9' was never specified; just write 0.0.
		mod.SetGno(GnoVerDefault)
		// err = fmt.Errorf("%s/gno.mod: gno version unspecified", mpkg.Path)
	} else if mod.Gno.Version == GnoVersion {
		// current version, nothing to do.
	} else {
		panic("unsupported gno version " + mod.Gno.Version)
	}
	return
}

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
		modPath := filepath.Join(path, "gno.mod")
		data, err := os.ReadFile(modPath)
		if os.IsNotExist(err) {
			return nil
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
			return fmt.Errorf("failed to validate gno.mod in %s: %w", modPath, err)
		}

		pkg, err := ReadMemPackage(path, mod.Module.Mod.Path)
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
			if imp.PkgPath != mod.Module.Mod.Path &&
				!IsStdlib(imp.PkgPath) {
				imports = append(imports, imp.PkgPath)
			}
		}

		pkgs = append(pkgs, gnomod.Pkg{
			Dir:     path,
			Name:    mod.Module.Mod.Path,
			Draft:   mod.Draft,
			Imports: imports,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}
