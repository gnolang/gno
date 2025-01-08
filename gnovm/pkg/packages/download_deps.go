package packages

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"golang.org/x/mod/module"
)

// DownloadDeps recursively fetches the imports of a local package while following a given gno.mod replace directives
func DownloadDeps(conf *LoadConfig, pkgDir string, gnoMod *gnomod.File) error {
	if conf.Fetcher == nil {
		return errors.New("fetcher is nil")
	}

	pkg, err := gnolang.ReadMemPackage(pkgDir, gnoMod.Module.Mod.Path)
	if err != nil {
		return fmt.Errorf("read package at %q: %w", pkgDir, err)
	}
	importsMap, err := Imports(pkg, nil)
	if err != nil {
		return fmt.Errorf("read imports at %q: %w", pkgDir, err)
	}
	imports := importsMap.Merge(FileKindPackageSource, FileKindTest, FileKindXTest)

	for _, pkgPath := range imports {
		resolved := gnoMod.Resolve(module.Version{Path: pkgPath})
		resolvedPkgPath := resolved.Path

		if !isRemotePkgPath(resolvedPkgPath) {
			continue
		}

		depDir := gnomod.PackageDir("", module.Version{Path: resolvedPkgPath})

		if err := downloadPackage(conf, resolvedPkgPath, depDir); err != nil {
			return fmt.Errorf("download import %q of %q: %w", resolvedPkgPath, pkgDir, err)
		}

		if err := DownloadDeps(conf, depDir, gnoMod); err != nil {
			return err
		}
	}

	return nil
}

// Download downloads a remote gno package by pkg path and store it at dst
func downloadPackage(conf *LoadConfig, pkgPath string, dst string) error {
	modFilePath := filepath.Join(dst, "gno.mod")

	if _, err := os.Stat(modFilePath); err == nil {
		// modfile exists in modcache, do nothing
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat downloaded module %q at %q: %w", pkgPath, dst, err)
	}

	conf.IO.ErrPrintfln("gno: downloading %s", pkgPath)

	if err := pkgdownload.Download(pkgPath, dst, conf.Fetcher); err != nil {
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
