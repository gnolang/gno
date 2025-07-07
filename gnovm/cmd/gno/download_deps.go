package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// downloadDeps recursively fetches the imports of a local package while following a given gno.mod replace directives
func downloadDeps(io commands.IO, pkgDir string, modfile *gnomod.File, fetcher pkgdownload.PackageFetcher, visited map[string]struct{}) error {
	if fetcher == nil {
		return errors.New("fetcher is nil")
	}

	pkg, err := gnolang.ReadMemPackage(pkgDir, modfile.Module, gnolang.MPUserAll)
	if err != nil {
		return fmt.Errorf("read package at %q: %w", pkgDir, err)
	}
	importsMap, err := packages.Imports(pkg, nil)
	if err != nil {
		return fmt.Errorf("read imports at %q: %w", pkgDir, err)
	}
	imports := importsMap.Merge(packages.FileKindPackageSource, packages.FileKindTest, packages.FileKindXTest)

	for _, imp := range imports {
		resolved := modfile.Resolve(imp.PkgPath)

		if !isRemotePkgPath(resolved) {
			continue
		}

		// Cycle + redundancy check: Have we already started processing this dependency?
		if _, exists := visited[resolved]; exists {
			continue // Skip dependencies already being processed or finished in this run.
		}
		// Mark this dependency as visited *before* recursive call.
		visited[resolved] = struct{}{}

		cachePath := gnomod.ModCachePath()
		depDir := filepath.Join(cachePath, filepath.FromSlash(resolved))

		if err := packages.DownloadPackage(io.Err(), resolved, depDir, fetcher); err != nil {
			return fmt.Errorf("download import %q of %q: %w", resolved, pkgDir, err)
		}

		if err := downloadDeps(io, depDir, modfile, fetcher, visited); err != nil {
			return err
		}
	}

	return nil
}

// isRemotePkgPath determines whether s is a remote pkg path, i.e.: not a filepath nor a standard library
func isRemotePkgPath(s string) bool {
	return !strings.HasPrefix(s, ".") && !filepath.IsAbs(s) && !gnolang.IsStdlib(s)
}
