// Package examplespkgfetcher provides an implementation of [pkgdownload.PackageFetcher]
// to fetch packages from the examples folder at GNOROOT
package examplespkgfetcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type ExamplesPackageFetcher struct {
	examplesDir string
}

var _ pkgdownload.PackageFetcher = (*ExamplesPackageFetcher)(nil)

func New(examplesDir string) pkgdownload.PackageFetcher {
	if examplesDir == "" {
		examplesDir = filepath.Join(gnoenv.RootDir(), "examples")
	}
	return &ExamplesPackageFetcher{examplesDir: examplesDir}
}

// FetchPackage implements [pkgdownload.PackageFetcher].
func (e *ExamplesPackageFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	pkgDir := filepath.Join(e.examplesDir, filepath.FromSlash(pkgPath))

	entries, err := os.ReadDir(pkgDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("query files list for pkg %q: package %q is not available", pkgPath, pkgPath)
	} else if err != nil {
		return nil, err
	}

	res := []*std.MemFile{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		filePath := filepath.Join(pkgDir, name)

		body, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read file at %q: %w", filePath, err)
		}

		res = append(res, &std.MemFile{Name: name, Body: string(body)})
	}

	return res, nil
}
