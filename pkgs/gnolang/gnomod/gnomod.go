package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/std"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

var source = "test3.gno.land:36657" // "127.0.0.1:26657"
const queryPathFile = "vm/qfile"

// ReadModFile reads, parses and validates the mod file at gnomod.
func ReadModFile(absModPath string) (f *File, err error) {
	data, err := os.ReadFile(absModPath)
	if err != nil {
		return nil, fmt.Errorf("readfile %q: %w", absModPath, err)
	}

	f, err = Parse(absModPath, data)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return f, err
}

func IsModFileExist(absModPath string) bool {
	_, err := os.Stat(absModPath)
	return err == nil
}

// GetGnoModPath returns the path for gno modules
func GetGnoModPath() (string, error) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return "", errors.New("GOPATH not found")
	}

	return filepath.Join(goPath, "pkg/gnomod"), nil
}

func writePackage(basePath, pkgPath string) error {
	res, err := queryChain(queryPathFile, []byte(pkgPath))
	if err != nil {
		return fmt.Errorf("querychain: %w", err)
	}

	dirPath, fileName := std.SplitFilepath(pkgPath)
	if fileName == "" {
		// Is Dir
		// Create Dir if not exists
		dirPath := filepath.Join(basePath, dirPath)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, 0o755); err != nil {
				return fmt.Errorf("mkdir %q: %w", dirPath, err)
			}
		}

		files := strings.Split(string(res.Data), "\n")
		for _, file := range files {
			if err := writePackage(basePath, filepath.Join(pkgPath, file)); err != nil {
				return fmt.Errorf("writepackage: %w", err)
			}
		}
	} else {
		// Is File
		// Precompile
		filePath := filepath.Join(basePath, pkgPath)
		targetFilename, _ := gnolang.GetPrecompileFilenameAndTags(filePath)
		precompileRes, err := gnolang.Precompile(string(res.Data), "", fileName)
		if err != nil {
			return fmt.Errorf("precompile: %w", err)
		}

		fileNameWithPath := filepath.Join(basePath, dirPath, targetFilename)
		err = os.WriteFile(fileNameWithPath, []byte(precompileRes.Translated), 0o644)
		if err != nil {
			return fmt.Errorf("writefile %q: %w", fileNameWithPath, err)
		}
	}

	return nil
}

// Sanitize make necessary modifications in the gno.mod
// before writing it to go.mod file.
func Sanitize(f *File) error {
	gnoModPath, err := GetGnoModPath()
	if err != nil {
		return err
	}

	if strings.HasPrefix(f.Module.Mod.Path, "gno.land/r/") ||
		strings.HasPrefix(f.Module.Mod.Path, "gno.land/p/demo/") {
		f.Module.Mod.Path = "github.com/gnolang/gno/examples/" + f.Module.Mod.Path
	}

	for i := range f.Require {
		path := f.Require[i].Mod.Path
		if strings.HasPrefix(f.Require[i].Mod.Path, "gno.land/r/") ||
			strings.HasPrefix(f.Require[i].Mod.Path, "gno.land/p/demo/") {
			f.Require[i].Mod.Path = "github.com/gnolang/gno/examples/" + f.Require[i].Mod.Path
		}

		f.Replace = append(f.Replace, &modfile.Replace{
			Old: module.Version{
				Path:    f.Require[i].Mod.Path,
				Version: f.Require[i].Mod.Version,
			},
			New: module.Version{
				Path: filepath.Join(gnoModPath, path),
			},
		})
	}

	return nil
}
