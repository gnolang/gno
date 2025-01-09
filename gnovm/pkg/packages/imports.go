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

func Imports(pkg *gnovm.MemPackage, fset *token.FileSet) (ImportsMap, error) {
	specs, err := ImportsSpecs(pkg, fset)
	if err != nil {
		return nil, err
	}
	return ImportsMapFromSpecs(specs, fset), nil
}

// Imports returns the list of gno imports from a [gnovm.MemPackage].
// fset is optional.
func ImportsSpecs(pkg *gnovm.MemPackage, fset *token.FileSet) (ImportsSpecsMap, error) {
	res := make(ImportsSpecsMap, 16)
	seen := make(map[FileKind]map[string]struct{}, 16)

	for _, file := range pkg.Files {
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}

		fileKind, err := GetFileKind(file.Name, file.Body, fset)
		if err != nil {
			return nil, err
		}
		imports, err := FileImportsSpecs(file.Name, file.Body, fset)
		if err != nil {
			return nil, err
		}
		for _, im := range imports {
			if _, ok := seen[fileKind][im.Path.Value]; ok {
				continue
			}
			res[fileKind] = append(res[fileKind], im)
			if _, ok := seen[fileKind]; !ok {
				seen[fileKind] = make(map[string]struct{}, 16)
			}
			seen[fileKind][im.Path.Value] = struct{}{}
		}
	}

	for _, imports := range res {
		sortImportsSpecs(imports)
	}

	return res, nil
}

func FileImportsSpecs(filename string, src string, fset *token.FileSet) ([]*ast.ImportSpec, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}
	f, err := parser.ParseFile(fset, filename, src, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	res := make([]*ast.ImportSpec, len(f.Imports))
	for i, im := range f.Imports {
		_, err := strconv.Unquote(im.Path.Value)
		if err != nil {
			// should not happen - parser.ParseFile should already ensure we get
			// a valid string literal here.
			panic(fmt.Errorf("%v: unexpected invalid import path: %v", fset.Position(im.Pos()).String(), im.Path.Value))
		}

		res[i] = im
	}
	return res, nil
}

type ImportsMap map[FileKind][]string

func ImportsMapFromSpecs(specs ImportsSpecsMap, fset *token.FileSet) ImportsMap {
	res := make(ImportsMap, len(specs))
	for k, v := range specs {
		c := make([]string, 0, len(v))
		for _, x := range v {
			pkgPath, err := strconv.Unquote(x.Path.Value)
			if err != nil {
				// should not happen - parser.ParseFile should already ensure we get
				// a valid string literal here.
				panic(fmt.Errorf("%v: unexpected invalid import path: %v", fset.Position(x.Pos()).String(), x.Path.Value))
			}
			c = append(c, pkgPath)
		}
		res[k] = c
	}
	return res
}

// Merge merges imports, it removes duplicates and sorts the result
func (imap ImportsMap) Merge(kinds ...FileKind) []string {
	res := make([]string, 0, 16)
	seen := make(map[string]struct{}, 16)

	for _, kind := range kinds {
		for _, im := range imap[kind] {
			if _, ok := seen[im]; ok {
				continue
			}
			seen[im] = struct{}{}

			res = append(res, im)
		}
	}

	sortImports(res)
	return res
}

type ImportsSpecsMap map[FileKind][]*ast.ImportSpec

// Merge merges imports, it removes duplicates and sorts the result
func (imap ImportsSpecsMap) Merge(kinds ...FileKind) []*ast.ImportSpec {
	res := make([]*ast.ImportSpec, 0, 16)
	seen := make(map[string]struct{}, 16)

	for _, kind := range kinds {
		for _, im := range imap[kind] {
			if _, ok := seen[im.Path.Value]; ok {
				continue
			}
			seen[im.Path.Value] = struct{}{}

			res = append(res, im)
		}
	}

	sortImportsSpecs(res)
	return res
}

func sortImports(imports []string) {
	sort.Slice(imports, func(i, j int) bool {
		return imports[i] < imports[j]
	})
}

func sortImportsSpecs(imports []*ast.ImportSpec) {
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Path.Value < imports[j].Path.Value
	})
}
