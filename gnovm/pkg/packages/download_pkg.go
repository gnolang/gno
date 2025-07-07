package packages

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
)

// XXX: duplicate with gno cmd

// DownloadPackage downloads a remote gno package by pkg path and store it at dst
func DownloadPackage(out io.Writer, pkgPath string, dst string, fetcher pkgdownload.PackageFetcher) error {
	modFilePath := filepath.Join(dst, "gnomod.toml")

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

	return nil
}
