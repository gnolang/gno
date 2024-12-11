package packages

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm"
)

type ImportsMap map[FileKind][]string

// Merge merges imports, it removes duplicates and sorts the result
func (imap ImportsMap) Merge(kinds ...FileKind) []string {
	res := []string{}
	seen := map[string]struct{}{}

	for _, kind := range kinds {
		for _, im := range imap[kind] {
			if _, ok := seen[im]; ok {
				continue
			}
			seen[im] = struct{}{}

			res = append(res, im)
		}
	}

	sort.Strings(res)
	return res
}

// Imports returns the list of gno imports from a [gnovm.MemPackage].
func Imports(pkg *gnovm.MemPackage) (ImportsMap, error) {
	res := make(ImportsMap)
	seen := make(map[FileKind]map[string]struct{})

	for _, file := range pkg.Files {
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}

		fileKind, err := GetFileKind(file.Name, file.Body)
		if err != nil {
			return nil, err
		}

		imports, _, err := FileImports(file.Name, file.Body)
		if err != nil {
			return nil, err
		}
		for _, im := range imports {
			if im.Error != nil {
				return nil, err
			}
			if _, ok := seen[fileKind][im.PkgPath]; ok {
				continue
			}
			res[fileKind] = append(res[fileKind], im.PkgPath)
			if _, ok := seen[fileKind]; !ok {
				seen[fileKind] = make(map[string]struct{})
			}
			seen[fileKind][im.PkgPath] = struct{}{}
		}
	}

	for _, imports := range res {
		sort.Strings(imports)
	}

	return res, nil
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
