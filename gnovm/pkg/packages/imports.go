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

// Imports returns the list of gno imports from a [gnovm.MemPackage].
// fset is optional.
func Imports(pkg *gnovm.MemPackage, fset *token.FileSet) (ImportsMap, error) {
	res := make(ImportsMap, 16)
	seen := make(map[FileKind]map[string]struct{}, 16)

	for _, file := range pkg.Files {
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}

		fileKind, err := GetFileKind(file.Name, file.Body, fset)
		if err != nil {
			return nil, err
		}
		imports, err := FileImports(file.Name, file.Body, fset)
		if err != nil {
			return nil, err
		}
		for _, im := range imports {
			if _, ok := seen[fileKind][im.PkgPath]; ok {
				continue
			}
			res[fileKind] = append(res[fileKind], im)
			if _, ok := seen[fileKind]; !ok {
				seen[fileKind] = make(map[string]struct{}, 16)
			}
			seen[fileKind][im.PkgPath] = struct{}{}
		}
	}

	for _, imports := range res {
		sortImports(imports)
	}

	return res, nil
}

// FileImport represents an import
type FileImport struct {
	PkgPath string
	Spec    *ast.ImportSpec
}

// FileImports returns the list of gno imports in the given file src.
// The given filename is only used when recording position information.
func FileImports(filename string, src string, fset *token.FileSet) ([]FileImport, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}
	f, err := parser.ParseFile(fset, filename, src, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	res := make([]FileImport, len(f.Imports))
	for i, im := range f.Imports {
		importPath, err := strconv.Unquote(im.Path.Value)
		if err != nil {
			// should not happen - parser.ParseFile should already ensure we get
			// a valid string literal here.
			panic(fmt.Errorf("%v: unexpected invalid import path: %v", fset.Position(im.Pos()).String(), im.Path.Value))
		}

		res[i] = FileImport{
			PkgPath: importPath,
			Spec:    im,
		}
	}
	return res, nil
}

type ImportsMap map[FileKind][]FileImport

// Merge merges imports, it removes duplicates and sorts the result
func (imap ImportsMap) Merge(kinds ...FileKind) []FileImport {
	res := make([]FileImport, 0, 16)
	seen := make(map[string]struct{}, 16)

	for _, kind := range kinds {
		for _, im := range imap[kind] {
			if _, ok := seen[im.PkgPath]; ok {
				continue
			}
			seen[im.PkgPath] = struct{}{}

			res = append(res, im)
		}
	}

	sortImports(res)
	return res
}

func sortImports(imports []FileImport) {
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].PkgPath < imports[j].PkgPath
	})
}
