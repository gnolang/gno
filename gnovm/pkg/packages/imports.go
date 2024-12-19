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
func Imports(pkg *gnovm.MemPackage, fset *token.FileSet) ([]FileImport, error) {
	allImports := make([]FileImport, 0, 16)
	seen := make(map[string]struct{}, 16)
	for _, file := range pkg.Files {
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}
		if strings.HasSuffix(file.Name, "_filetest.gno") {
			continue
		}
		imports, err := FileImports(file.Name, file.Body, fset)
		if err != nil {
			return nil, err
		}
		for _, im := range imports {
			if _, ok := seen[im.PkgPath]; ok {
				continue
			}
			allImports = append(allImports, im)
			seen[im.PkgPath] = struct{}{}
		}
	}
	sort.Slice(allImports, func(i, j int) bool {
		return allImports[i].PkgPath < allImports[j].PkgPath
	})

	return allImports, nil
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
