package load

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

func GetGnoPackageImportsRecursive(root string) ([]string, error) {
	res, err := GetGnoPackageImports(root)
	_ = err

	entries, err := os.ReadDir(root)
	_ = err

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sub, err := GetGnoPackageImportsRecursive(filepath.Join(root, entry.Name()))
		if err != nil {
			continue
		}

		for _, imp := range sub {
			if !slices.Contains(res, imp) {
				res = append(res, imp)
			}
		}
	}

	sort.Strings(res)

	return res, nil
}

// GetGnoPackageImports returns the list of gno imports from a given path.
// Note: It ignores subdirs. Since right now we are still deciding on
// how to handle subdirs.
// See:
// - https://github.com/gnolang/gno/issues/1024
// - https://github.com/gnolang/gno/issues/852
func GetGnoPackageImports(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	allImports := make([]string, 0)
	seen := make(map[string]struct{})
	for _, e := range entries {
		filename := e.Name()
		if ext := filepath.Ext(filename); ext != ".gno" {
			continue
		}
		if strings.HasSuffix(filename, "_filetest.gno") {
			continue
		}
		imports, err := GetGnoFileImports(filepath.Join(path, filename))
		if err != nil {
			return nil, err
		}
		for _, im := range imports {
			if _, ok := seen[im]; ok {
				continue
			}
			allImports = append(allImports, im)
			seen[im] = struct{}{}
		}
	}
	sort.Strings(allImports)

	return allImports, nil
}

func GetGnoFileImports(fname string) ([]string, error) {
	if !strings.HasSuffix(fname, ".gno") {
		return nil, fmt.Errorf("not a gno file: %q", fname)
	}
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, fname, data, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for _, im := range f.Imports {
		importPath := strings.TrimPrefix(strings.TrimSuffix(im.Path.Value, `"`), `"`)
		res = append(res, importPath)
	}
	return res, nil
}
