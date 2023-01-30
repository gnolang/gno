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

var (
	source        = "test3.gno.land:36657" // "127.0.0.1:26657"
	queryPathFile = "vm/qfile"
)

// ReadModFile reads, parses and validates the mod file at gnomod.
func ReadModFile(absModPath string) (f *File, err error) {
	data, err := os.ReadFile(absModPath)
	if err != nil {
		return nil, fmt.Errorf("reading gno.mod: %s", err)
	}

	f, err = Parse(absModPath, data)
	if err != nil {
		return nil, fmt.Errorf("parsing gno.mod: %s", err)
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
	res, err := QueryChain(queryPathFile, []byte(pkgPath))
	if err != nil {
		return fmt.Errorf("makeReq gno.mod: %s", err)
	}

	dirPath, fileName := std.SplitFilepath(pkgPath)
	if fileName == "" {
		// Is Dir
		// Create Dir if not exists
		dirPath := filepath.Join(basePath, dirPath)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, 0o755); err != nil {
				return fmt.Errorf("creating pkg dir: %s", err)
			}
		}

		files := strings.Split(string(res.Data), "\n")
		for _, file := range files {
			if err := writePackage(basePath, filepath.Join(pkgPath, file)); err != nil {
				return fmt.Errorf("writing mod files: %s", err)
			}
		}
	} else {
		// Is File
		// Precompile
		filePath := filepath.Join(basePath, pkgPath)
		_, targetFilename := ResolvePrecompileName(filePath)
		precompileRes, err := gnolang.Precompile(string(res.Data), "", fileName)
		if err != nil {
			return fmt.Errorf("precompiling modules: %s", err)
		}

		err = os.WriteFile(filepath.Join(basePath, dirPath, targetFilename), []byte(precompileRes.Translated), 0o644)
		if err != nil {
			return fmt.Errorf("writing mod files: %s", err)
		}
	}

	return nil
}

func ResolvePrecompileName(gnoFilePath string) (tags, targetFilename string) {
	nameNoExtension := strings.TrimSuffix(filepath.Base(gnoFilePath), ".gno")
	switch {
	case strings.HasSuffix(gnoFilePath, "_filetest.gno"):
		tags = "gno,filetest"
		targetFilename = "." + nameNoExtension + ".gno.gen.go"
	case strings.HasSuffix(gnoFilePath, "_test.gno"):
		tags = "gno,test"
		targetFilename = "." + nameNoExtension + ".gno.gen_test.go"
	default:
		tags = "gno"
		targetFilename = nameNoExtension + ".gno.gen.go"
	}
	return
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
