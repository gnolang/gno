package packages

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
)

// XXX: duplicate with gno cmd

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
