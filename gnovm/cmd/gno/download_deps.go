package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/mod/module"
)

// downloadDeps recursively fetches the imports of a local package while following a given gno.mod replace directives
func downloadDeps(io commands.IO, pkgDir string, gnoMod *gnomod.File, fetcher pkgdownload.PackageFetcher, visited map[string]struct{}) error {
	if fetcher == nil {
		return errors.New("fetcher is nil")
	}

	pkg, err := gnolang.ReadMemPackage(pkgDir, gnoMod.Module.Mod.Path)
	if err != nil {
		return fmt.Errorf("read package at %q: %w", pkgDir, err)
	}
	importsMap, err := packages.Imports(pkg, nil)
	if err != nil {
		return fmt.Errorf("read imports at %q: %w", pkgDir, err)
	}
	imports := importsMap.Merge(packages.FileKindPackageSource, packages.FileKindTest, packages.FileKindXTest)

	for _, pkgPath := range imports {
		resolved := gnoMod.Resolve(module.Version{Path: pkgPath.PkgPath})
		resolvedPkgPath := resolved.Path

		if !isRemotePkgPath(resolvedPkgPath) {
			continue
		}

		// Cycle + redundancy check: Have we already started processing this dependency?
		if _, exists := visited[resolvedPkgPath]; exists {
			continue // Skip dependencies already being processed or finished in this run.
		}
		// Mark this dependency as visited *before* recursive call.
		visited[resolvedPkgPath] = struct{}{}

		depDir := gnomod.PackageDir("", module.Version{Path: resolvedPkgPath})

		if err := downloadPackage(io, resolvedPkgPath, depDir, fetcher); err != nil {
			return fmt.Errorf("download import %q of %q: %w", resolvedPkgPath, pkgDir, err)
		}

		if err := downloadDeps(io, depDir, gnoMod, fetcher, visited); err != nil {
			return err
		}
	}

	return nil
}

// downloadPackage downloads a remote gno package by pkg path and store it at dst
func downloadPackage(io commands.IO, pkgPath string, dst string, fetcher pkgdownload.PackageFetcher) error {
	modFilePath := filepath.Join(dst, "gno.mod")

	if _, err := os.Stat(modFilePath); err == nil {
		// modfile exists in modcache, do nothing
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat downloaded module %q at %q: %w", pkgPath, dst, err)
	}

	io.ErrPrintfln("gno: downloading %s", pkgPath)

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
