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

const queryPathFile = "vm/qfile"

// GetGnoModPath returns the path for gno modules
func GetGnoModPath() (string, error) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return "", errors.New("GOPATH not found")
	}

	return filepath.Join(goPath, "pkg", "gnomod"), nil
}

func writePackage(remote, basePath, pkgPath string) error {
	res, err := queryChain(remote, queryPathFile, []byte(pkgPath))
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
			if err := writePackage(remote, basePath, filepath.Join(pkgPath, file)); err != nil {
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

// GnoToGoMod make necessary modifications in the gno.mod
// and return go.mod file.
func GnoToGoMod(f File) (*File, error) {
	gnoModPath, err := GetGnoModPath()
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(f.Module.Mod.Path, "gno.land/r/") ||
		strings.HasPrefix(f.Module.Mod.Path, "gno.land/p/demo/") {
		f.Module.Mod.Path = "github.com/gnolang/gno/examples/" + f.Module.Mod.Path
	}

	for i := range f.Require {
		mod, replaced := isReplaced(f.Require[i].Mod, f.Replace)
		if replaced {
			if modfile.IsDirectoryPath(mod.Path) {
				continue
			}
		}
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

	// Since we already fetched and replaced replacement modules
	// with `/pkg/gnomod/...` path.
	// Ignore leftovers.
	repl := make([]*modfile.Replace, 0, len(f.Replace))
	for _, r := range f.Replace {
		if !modfile.IsDirectoryPath(r.New.Path) {
			continue
		}
		repl = append(repl, r)
	}
	f.Replace = repl

	return &f, nil
}

func isReplaced(module module.Version, repl []*modfile.Replace) (*module.Version, bool) {
	for _, r := range repl {
		if (r.Old.Path == module.Path && r.Old.Version == "") || r.Old == module {
			return &r.New, true
		}
	}
	return nil, false
}
