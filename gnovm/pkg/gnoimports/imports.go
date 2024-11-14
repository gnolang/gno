package gnoimports

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
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
		imports, err := FileImports(filepath.Join(path, filename))
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

// FileImports returns the list of gno imports in a given file.
func FileImports(fname string) ([]string, error) {
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
