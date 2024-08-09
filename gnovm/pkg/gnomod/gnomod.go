package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// GetGnoModPath returns the path for gno modules
func GetGnoModPath() string {
	return filepath.Join(gnoenv.HomeDir(), "pkg", "mod")
}

// PackageDir resolves a given module.Version to the path on the filesystem.
// If root is dir, it is defaulted to the value of [GetGnoModPath].
func PackageDir(root string, v module.Version) string {
	// This is also used internally exactly like filepath.Join; but we'll keep
	// the calls centralized to make sure we can change the path centrally should
	// we start including the module version in the path.

	if root == "" {
		root = GetGnoModPath()
	}
	return filepath.Join(root, v.Path)
}

func CreateGnoModFile(rootDir, modPath string) error {
	if !filepath.IsAbs(rootDir) {
		return fmt.Errorf("dir %q is not absolute", rootDir)
	}

	modFilePath := filepath.Join(rootDir, "gno.mod")
	if _, err := os.Stat(modFilePath); err == nil {
		return errors.New("gno.mod file already exists")
	}

	if modPath == "" {
		// Check .gno files for package name
		// and use it as modPath
		files, err := os.ReadDir(rootDir)
		if err != nil {
			return fmt.Errorf("read dir %q: %w", rootDir, err)
		}

		var pkgName gnolang.Name
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".gno") || strings.HasSuffix(file.Name(), "_filetest.gno") {
				continue
			}

			fpath := filepath.Join(rootDir, file.Name())
			bz, err := os.ReadFile(fpath)
			if err != nil {
				return fmt.Errorf("read file %q: %w", fpath, err)
			}

			pn := gnolang.PackageNameFromFileBody(file.Name(), string(bz))
			if strings.HasSuffix(string(pkgName), "_test") {
				pkgName = pkgName[:len(pkgName)-len("_test")]
			}
			if pkgName == "" {
				pkgName = pn
			}
			if pkgName != pn {
				return fmt.Errorf("package name mismatch: [%q] and [%q]", pkgName, pn)
			}
		}
		if pkgName == "" {
			return errors.New("cannot determine package name")
		}
		modPath = string(pkgName)
	}
	if err := module.CheckImportPath(modPath); err != nil {
		return err
	}

	modfile := new(File)
	modfile.AddModuleStmt(modPath)
	modfile.Write(filepath.Join(rootDir, "gno.mod"))

	return nil
}

func isReplaced(mod module.Version, repl []*modfile.Replace) (module.Version, bool) {
	for _, r := range repl {
		hasNoVersion := r.Old.Path == mod.Path && r.Old.Version == ""
		hasExactVersion := r.Old == mod
		if hasNoVersion || hasExactVersion {
			return r.New, true
		}
	}
	return module.Version{}, false
}
