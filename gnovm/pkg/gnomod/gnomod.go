package gnomod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/std"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

const queryPathFile = "vm/qfile"

// GetGnoModPath returns the path for gno modules
func GetGnoModPath() string {
	return filepath.Join(client.HomeDir(), "pkg", "mod")
}

func writePackage(remote, basePath, pkgPath string) (requirements []string, err error) {
	res, err := queryChain(remote, queryPathFile, []byte(pkgPath))
	if err != nil {
		return nil, fmt.Errorf("querychain: %w", err)
	}

	dirPath, fileName := std.SplitFilepath(pkgPath)
	if fileName == "" {
		// Is Dir
		// Create Dir if not exists
		dirPath := filepath.Join(basePath, dirPath)
		if _, err = os.Stat(dirPath); os.IsNotExist(err) {
			if err = os.MkdirAll(dirPath, 0o755); err != nil {
				return nil, fmt.Errorf("mkdir %q: %w", dirPath, err)
			}
		}

		files := strings.Split(string(res.Data), "\n")
		for _, file := range files {
			reqs, err := writePackage(remote, basePath, filepath.Join(pkgPath, file))
			if err != nil {
				return nil, fmt.Errorf("writepackage: %w", err)
			}
			requirements = append(requirements, reqs...)
		}
	} else {
		// Is File
		// Precompile
		filePath := filepath.Join(basePath, pkgPath)
		targetFilename, _ := gnolang.GetPrecompileFilenameAndTags(filePath)
		precompileRes, err := gnolang.Precompile(string(res.Data), "", fileName)
		if err != nil {
			return nil, fmt.Errorf("precompile: %w", err)
		}

		for _, i := range precompileRes.Imports {
			requirements = append(requirements, i.Path.Value)
		}

		fileNameWithPath := filepath.Join(basePath, dirPath, targetFilename)
		err = os.WriteFile(fileNameWithPath, []byte(precompileRes.Translated), 0o644)
		if err != nil {
			return nil, fmt.Errorf("writefile %q: %w", fileNameWithPath, err)
		}
	}

	return removeDuplicateStr(requirements), nil
}

// GnoToGoMod make necessary modifications in the gno.mod
// and return go.mod file.
func GnoToGoMod(f File) (*File, error) {
	gnoModPath := GetGnoModPath()

	if strings.HasPrefix(f.Module.Mod.Path, gnolang.GnoRealmPkgsPrefixBefore) ||
		strings.HasPrefix(f.Module.Mod.Path, gnolang.GnoPackagePrefixBefore) {
		f.Module.Mod.Path = gnolang.ImportPrefix + "/examples/" + f.Module.Mod.Path
	}

	for i := range f.Require {
		mod, replaced := isReplaced(f.Require[i].Mod, f.Replace)
		if replaced {
			if modfile.IsDirectoryPath(mod.Path) {
				continue
			}
		}
		path := f.Require[i].Mod.Path
		if strings.HasPrefix(f.Require[i].Mod.Path, gnolang.GnoRealmPkgsPrefixBefore) ||
			strings.HasPrefix(f.Require[i].Mod.Path, gnolang.GnoPackagePrefixBefore) {
			f.Require[i].Mod.Path = gnolang.ImportPrefix + "/examples/" + f.Require[i].Mod.Path
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

	// By this stage every replacement should be replace by dir.
	// If not replaced by dir, remove it.
	//
	// e.g:
	//
	// ```
	// require (
	//	gno.land/p/demo/avl v1.2.3
	// )
	//
	// replace (
	//	gno.land/p/demo/avl v1.2.3  => gno.land/p/demo/avl v3.2.1
	// )
	// ```
	//
	// In above case we will fetch `gno.land/p/demo/avl v3.2.1` and
	// replace will look something like:
	//
	// ```
	// replace (
	//	gno.land/p/demo/avl v1.2.3  => gno.land/p/demo/avl v3.2.1
	//	gno.land/p/demo/avl v3.2.1  => /path/to/avl/version/v3.2.1
	// )
	// ```
	//
	// Remove `gno.land/p/demo/avl v1.2.3  => gno.land/p/demo/avl v3.2.1`.
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
		hasNoVersion := r.Old.Path == module.Path && r.Old.Version == ""
		hasExactVersion := r.Old == module
		if hasNoVersion || hasExactVersion {
			return &r.New, true
		}
	}
	return nil, false
}

func removeDuplicateStr(str []string) (res []string) {
	m := make(map[string]struct{}, len(str))
	for _, s := range str {
		if _, ok := m[s]; !ok {
			m[s] = struct{}{}
			res = append(res, s)
		}
	}
	return
}
