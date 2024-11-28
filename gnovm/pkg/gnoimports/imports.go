package gnoimports

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// PackageImports returns the list of gno imports from a given path.
// Note: It ignores subdirs. Since right now we are still deciding on
// how to handle subdirs.
// See:
// - https://github.com/gnolang/gno/issues/1024
// - https://github.com/gnolang/gno/issues/852
func PackageImports(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	allImports := make([]string, 0)
	seen := make(map[string]struct{})
	for _, e := range entries {
		filename := e.Name()
		if !strings.HasSuffix(filename, ".gno") {
			continue
		}
		if strings.HasSuffix(filename, "_filetest.gno") {
			continue
		}
		filePath := filepath.Join(path, filename)
		imports, _, err := FileImportsFromPath(filePath)
		if err != nil {
			return nil, err
		}
		for _, im := range imports {
			if im.Error != nil {
				return nil, err
			}
			if _, ok := seen[im.PkgPath]; ok {
				continue
			}
			allImports = append(allImports, im.PkgPath)
			seen[im.PkgPath] = struct{}{}
		}
	}
	sort.Strings(allImports)

	return allImports, nil
}

type FileImport struct {
	PkgPath string
	Spec    *ast.ImportSpec
	Error   error
}

// FileImports returns the list of gno imports in the given file src.
// The given filename is only used when recording position information.
func FileImports(filename string, src []byte) ([]*FileImport, *token.FileSet, error) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, filename, src, parser.ImportsOnly)
	if err != nil {
		return nil, nil, err
	}
	res := make([]*FileImport, len(f.Imports))
	for i, im := range f.Imports {
		fi := FileImport{Spec: im}
		importPath, err := strconv.Unquote(im.Path.Value)
		if err != nil {
			fi.Error = fmt.Errorf("%v: unexpected invalid import path: %v", fs.Position(im.Pos()).String(), im.Path.Value)
		} else {
			fi.PkgPath = importPath
		}
		res[i] = &fi
	}
	return res, fs, nil
}

// FileImportsFromPath reads the file at filePath and returns the list of gno imports in it.
func FileImportsFromPath(filePath string) ([]*FileImport, *token.FileSet, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	return FileImports(filePath, data)
}
