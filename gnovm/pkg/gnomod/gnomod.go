package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/importer"
	"github.com/gnolang/gno/gnovm/pkg/transpiler"
	"github.com/gnolang/gno/tm2/pkg/std"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

const queryPathFile = "vm/qfile"

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

func writePackage(remote, basePath, pkgPath string) (requirements []string, err error) {
	res, err := queryChain(remote, queryPathFile, []byte(pkgPath))
	if err != nil {
		return nil, fmt.Errorf("querychain (%s): %w", pkgPath, err)
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
		// Transpile and write generated go file
		if strings.HasSuffix(fileName, ".gno") {
			filePath := filepath.Join(basePath, pkgPath)
			targetFilename, _ := transpiler.GetTranspileFilenameAndTags(filePath)
			transpileRes, err := transpiler.Transpile(string(res.Data), "", fileName)
			if err != nil {
				return nil, fmt.Errorf("transpile: %w", err)
			}

			for _, i := range transpileRes.Imports {
				requirements = append(requirements, i.Path.Value)
			}

			targetFileNameWithPath := filepath.Join(basePath, dirPath, targetFilename)
			err = os.WriteFile(targetFileNameWithPath, []byte(transpileRes.Translated), 0o644)
			if err != nil {
				return nil, fmt.Errorf("writefile %q: %w", targetFileNameWithPath, err)
			}
		}

		// Write file
		fileNameWithPath := filepath.Join(basePath, dirPath, fileName)
		err = os.WriteFile(fileNameWithPath, res.Data, 0o644)
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

	if strings.HasPrefix(f.Module.Mod.Path, transpiler.GnoRealmPkgsPrefixBefore) ||
		strings.HasPrefix(f.Module.Mod.Path, transpiler.GnoPackagePrefixBefore) {
		f.AddModuleStmt(transpiler.ImportPrefix + "/examples/" + f.Module.Mod.Path)
	}

	for i := range f.Require {
		mod, replaced := isReplaced(f.Require[i].Mod, f.Replace)
		if replaced {
			if modfile.IsDirectoryPath(mod.Path) {
				continue
			}
		}
		path := f.Require[i].Mod.Path
		if strings.HasPrefix(f.Require[i].Mod.Path, transpiler.GnoRealmPkgsPrefixBefore) ||
			strings.HasPrefix(f.Require[i].Mod.Path, transpiler.GnoPackagePrefixBefore) {
			// Add dependency with a modified import path
			f.AddRequire(transpiler.ImportPrefix+"/examples/"+f.Require[i].Mod.Path, f.Require[i].Mod.Version)
		}
		f.AddReplace(f.Require[i].Mod.Path, f.Require[i].Mod.Version, filepath.Join(gnoModPath, path), "")
		// Remove the old require since the new dependency was added above
		f.DropRequire(f.Require[i].Mod.Path)
	}

	// Remove replacements that are not replaced by directories.
	//
	// Explanation:
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
	for _, r := range f.Replace {
		if !modfile.IsDirectoryPath(r.New.Path) {
			f.DropReplace(r.Old.Path, r.Old.Version)
		}
	}

	return &f, nil
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
			if file.IsDir() || !importer.IsGnoFile(file.Name(), "!*_filetest.gno") {
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
