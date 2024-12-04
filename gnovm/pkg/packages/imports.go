package packages

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm"
)

// Imports returns the list of gno imports from a [gnovm.MemPackage].
func Imports(pkg *gnovm.MemPackage) ([]string, error) {
	allImports := make([]string, 0)
	seen := make(map[string]struct{})
	for _, file := range pkg.Files {
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}
		if strings.HasSuffix(file.Name, "_filetest.gno") {
			continue
		}
		imports, _, err := FileImports(file.Name, file.Body)
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
func FileImports(filename string, src string) ([]*FileImport, *token.FileSet, error) {
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
	return FileImports(filePath, string(data))
}
