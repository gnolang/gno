// Package examplespkgfetcher provides an implementation of [pkgdownload.PackageFetcher]
// to fetch packages from the examples folder at GNOROOT
package examplespkgfetcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	filetestsDir := ""
	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "filetests" {
				// Process filetests dir below
				filetestsDir = filepath.Join(pkgDir, entry.Name())
			}
			continue
		}

		name := entry.Name()
		filePath := filepath.Join(pkgDir, name)

		body, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read file at %q: %w", filePath, err)
		}

		// Classify by suffix at the root: filetest files end in `_filetest.gno`
		// (legacy layout); otherwise KindUnknown lets downstream fall back to
		// per-name inference for tests vs sources.
		kind := std.KindUnknown
		if std.IsFiletestName(name) {
			kind = std.KindFiletest
		}
		res = append(res, &std.MemFile{Name: name, Body: string(body), Kind: kind})
	}
	if filetestsDir != "" {
		filetestsFiles, err := os.ReadDir(filetestsDir)
		if err != nil {
			return nil, err
		}
		for _, entry := range filetestsFiles {
			if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			// Any `.gno` file in filetests/ is a filetest. The subdir IS the
			// classification, surfaced as MemFile.Kind = KindFiletest below.
			if !strings.HasSuffix(entry.Name(), ".gno") {
				continue
			}
			filePath := filepath.Join(filetestsDir, entry.Name())
			body, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("read file at %q: %w", filePath, err)
			}
			// MemFile.Name is a flat basename; the filetests/ subdir is a
			// write-routing convention on disk, not part of the name.
			res = append(res, &std.MemFile{
				Name: entry.Name(),
				Body: string(body),
				Kind: std.KindFiletest,
			})
		}
	}

	return res, nil
}
