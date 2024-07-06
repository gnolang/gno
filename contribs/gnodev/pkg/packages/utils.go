package packages

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm"
)

func isGnoFile(name string) bool {
	return filepath.Ext(name) == ".gno" && !strings.HasPrefix(name, ".")
}

func isTestFile(name string) bool {
	return strings.HasSuffix(name, "_filetest.gno") || strings.HasSuffix(name, "_test.gno")
}

func ReadPackageFromDir(fset *token.FileSet, path, dir string) (*Package, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir %q: %w", dir, err)
	}

	var name string
	memFiles := []*gnovm.MemFile{}
	for _, file := range files {
		fname := file.Name()
		if !isGnoFile(fname) || isTestFile(fname) {
			continue
		}

		filepath := filepath.Join(dir, fname)
		body, err := os.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("unable to read file %q: %w", filepath, err)
		}

		f, err := parser.ParseFile(fset, fname, body, parser.PackageClauseOnly)
		if err != nil {
			return nil, fmt.Errorf("unable to parse file %q: %w", fname, err)
		}

		if name != "" && name != f.Name.Name {
			return nil, fmt.Errorf("conflict package name between %q and %q", name, f.Name.Name)
		}

		name = f.Name.Name
		memFiles = append(memFiles, &gnovm.MemFile{
			Name: fname,
			Body: string(body),
		})
	}

	return &Package{
		MemPackage: gnovm.MemPackage{
			Name:  name,
			Path:  path,
			Files: memFiles,
		},
		Location: dir,
		Kind:     PackageKindFS,
	}, nil
}
