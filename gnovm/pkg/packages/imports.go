package packages

import (
	"fmt"
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
			if _, ok := seen[fileKind][im]; ok {
				continue
			}
			res[fileKind] = append(res[fileKind], im)
			if _, ok := seen[fileKind]; !ok {
				seen[fileKind] = make(map[string]struct{}, 16)
			}
			seen[fileKind][im] = struct{}{}
		}
	}

	for _, imports := range res {
		sortImports(imports)
	}

	return res, nil
}

// FileImports returns the list of gno imports in the given file src.
// The given filename is only used when recording position information.
func FileImports(filename string, src string, fset *token.FileSet) ([]string, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}
	f, err := parser.ParseFile(fset, filename, src, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	res := make([]string, len(f.Imports))
	for i, im := range f.Imports {
		importPath, err := strconv.Unquote(im.Path.Value)
		if err != nil {
			// should not happen - parser.ParseFile should already ensure we get
			// a valid string literal here.
			panic(fmt.Errorf("%v: unexpected invalid import path: %v", fset.Position(im.Pos()).String(), im.Path.Value))
		}

		res[i] = importPath
	}
	return res, nil
}

type ImportsMap map[FileKind][]string

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

func sortImports(imports []string) {
	sort.Slice(imports, func(i, j int) bool {
		return imports[i] < imports[j]
	})
}

func FilePackageName(filename string, src string) (string, error) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, filename, src, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return f.Name.Name, nil
}
