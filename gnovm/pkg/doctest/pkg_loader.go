package doctest

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

type dynPackageLoader struct {
	stdlibDir string
	cache     map[string]*std.MemPackage
}

func newDynPackageLoader(stdlibDir string) *dynPackageLoader {
	return &dynPackageLoader{
		stdlibDir: stdlibDir,
		cache:     make(map[string]*std.MemPackage),
	}
}

func (d *dynPackageLoader) GetMemPackage(path string) *std.MemPackage {
	if pkg, ok := d.cache[path]; ok {
		return pkg
	}

	pkg, err := d.loadPackage(path)
	if err != nil {
		return nil
	}

	d.cache[path] = pkg
	return pkg
}

func (d *dynPackageLoader) loadPackage(path string) (*std.MemPackage, error) {
	pkgDir := filepath.Join(d.stdlibDir, path)
	files, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, err
	}

	memFiles := []*std.MemFile{}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".gno") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(pkgDir, file.Name()))
		if err != nil {
			return nil, err
		}

		memFiles = append(memFiles, &std.MemFile{
			Name: file.Name(),
			Body: string(content),
		})
	}

	pkgName := ""
	if len(memFiles) > 0 {
		pkgName = extractPackageName(memFiles[0].Body)
	}

	return &std.MemPackage{
		Name:  pkgName,
		Path:  path,
		Files: memFiles,
	}, nil
}

func extractPackageName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}
