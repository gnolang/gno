package packages

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
)

// DownloadDeps recursively fetches the imports of a local package while following a given gno.mod replace directives
func DownloadDeps(out io.Writer, fetcher pkgdownload.PackageFetcher, pkgDir string, gnoMod *gnomod.File, visited map[string]struct{}) error {
	if fetcher == nil {
		return errors.New("fetcher is nil")
	}

	pkg, err := gnolang.ReadMemPackage(pkgDir, gnoMod.Module)
	if err != nil {
		return fmt.Errorf("read package at %q: %w", pkgDir, err)
	}
	importsMap, err := Imports(pkg, nil)
	if err != nil {
		return fmt.Errorf("read imports at %q: %w", pkgDir, err)
	}
	imports := importsMap.Merge(FileKindPackageSource, FileKindTest, FileKindXTest)

	for _, imp := range imports {
		resolvedPkgPath := gnoMod.Resolve(imp.PkgPath)

		if !isRemotePkgPath(resolvedPkgPath) {
			continue
		}

		// Cycle + redundancy check: Have we already started processing this dependency?
		if _, exists := visited[resolvedPkgPath]; exists {
			continue // Skip dependencies already being processed or finished in this run.
		}
		// Mark this dependency as visited *before* recursive call.
		visited[resolvedPkgPath] = struct{}{}

		depDir := PackageDir(resolvedPkgPath)

		if err := downloadPackage(out, fetcher, resolvedPkgPath, depDir); err != nil {
			return fmt.Errorf("download import %q of %q: %w", resolvedPkgPath, pkgDir, err)
		}

		if err := DownloadDeps(out, fetcher, depDir, gnoMod, visited); err != nil {
			return err
		}
	}

	return nil
}

// Download downloads a remote gno package by pkg path and store it at dst
func downloadPackage(out io.Writer, fetcher pkgdownload.PackageFetcher, pkgPath string, dst string) error {
	modFilePath := filepath.Join(dst, "gno.mod")

	if _, err := os.Stat(modFilePath); err == nil {
		// modfile exists in modcache, do nothing
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat downloaded module %q at %q: %w", pkgPath, dst, err)
	}

	fmt.Fprintf(out, "gno: downloading %s\n", pkgPath)

	if err := pkgdownload.Download(pkgPath, dst, fetcher); err != nil {
		return err
	}

	// We need to write a marker file for each downloaded package.
	// For example: if you first download gno.land/r/foo/bar then download gno.land/r/foo,
	// we need to know that gno.land/r/foo is not downloaded yet.
	// We do this by checking for the presence of gno.land/r/foo/gno.mod
	if err := os.WriteFile(modFilePath, []byte("module "+pkgPath+"\n"), 0o644); err != nil {
		return fmt.Errorf("write modfile at %q: %w", modFilePath, err)
	}

	return nil
}

// isRemotePkgPath determines whether s is a remote pkg path, i.e.: not a filepath nor a standard library
func isRemotePkgPath(s string) bool {
	return !strings.HasPrefix(s, ".") && !filepath.IsAbs(s) && !gnolang.IsStdlib(s)
}
