package packages

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// Imports returns the list of gno imports from a [std.MemPackage].
// fset is optional.
func Imports(pkg *std.MemPackage, fset *token.FileSet) (ImportsMap, error) {
	res := make(ImportsMap, 16)
	seen := make(map[FileKind]map[string]struct{}, 16)

	for _, file := range pkg.Files {
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}

		imports, err := FileImports(file.Name, file.Body, fset)
		if err != nil {
			return nil, err
		}

		fileKind := GetFileKind(file.Name, file.Body, fset)

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

// FileImport represents an import and it's location in the source file
type FileImport struct {
	PkgPath string
	Spec    *ast.ImportSpec
}

// FileImports returns the list of gno imports in the given file src.
// The given filename is only used when recording position information.
func FileImports(filename string, src string, fset *token.FileSet) ([]*FileImport, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}
	f, err := parser.ParseFile(fset, filename, src, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	res := make([]*FileImport, len(f.Imports))
	for i, spec := range f.Imports {
		pkgPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			// should not happen - parser.ParseFile should already ensure we get
			// a valid string literal here.
			panic(fmt.Errorf("%v: unexpected invalid import path: %v", fset.Position(spec.Pos()).String(), spec.Path.Value))
		}

		res[i] = &FileImport{
			PkgPath: pkgPath,
			Spec:    spec,
		}
	}
	return res, nil
}

type ImportsMap map[FileKind][]*FileImport

func (imap ImportsMap) ToStrings() map[FileKind][]string {
	res := make(map[FileKind][]string, len(imap))
	for k, v := range imap {
		c := make([]string, 0, len(v))
		for _, x := range v {
			c = append(c, x.PkgPath)
		}
		res[k] = c
	}
	return res
}

// Merge merges imports, it removes duplicates and sorts the result
func (imap ImportsMap) Merge(kinds ...FileKind) []*FileImport {
	if len(kinds) == 0 {
		kinds = GnoFileKinds()
	}

	res := make([]*FileImport, 0, 16)
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

func sortImports(imports []*FileImport) {
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].PkgPath < imports[j].PkgPath
	})
}
