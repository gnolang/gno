package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// ModCachePath returns the path for gno modules
func ModCachePath() string {
	return filepath.Join(gnoenv.HomeDir(), "pkg", "mod")
}

// PackageDir resolves a given module.Version to the path on the filesystem.
// If root is dir, it is defaulted to the value of [ModCachePath].
func PackageDir(root string, v module.Version) string {
	if root == "" {
		root = ModCachePath()
	}
	return filepath.Join(root, filepath.FromSlash(v.Path))
}

func CreateGnoModFile(rootDir, modPath string) error {
	if !filepath.IsAbs(rootDir) {
		return fmt.Errorf("dir %q is not absolute", rootDir)
	}

	modFilePath := filepath.Join(rootDir, "gno.mod")
	if _, err := os.Stat(modFilePath); err == nil {
		return errors.New("gno.mod file already exists")
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
