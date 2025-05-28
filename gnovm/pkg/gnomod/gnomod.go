package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"golang.org/x/mod/module"
)

// ModCachePath returns the path for gno modules
func ModCachePath() string {
	return filepath.Join(gnoenv.HomeDir(), "pkg", "mod")
}

// PackageDir resolves a given module.Version to the path on the filesystem.
// If root is empty, it is defaulted to the value of [ModCachePath].
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
	modfile.SetModule(modPath)
	modfile.WriteFile(filepath.Join(rootDir, "gno.mod"))

	return nil
}
