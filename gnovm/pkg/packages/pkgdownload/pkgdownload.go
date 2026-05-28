// Package pkgdownload provides interfaces and utility functions to download gno packages files.
package pkgdownload

import (
	"fmt"
	"os"
	"path/filepath"
)

// Download downloads the package identified by `pkgPath` in the directory at
// `dst` using the provided [PackageFetcher]. The directory at `dst` is
// created if it does not exist. Layout follows MemFile.DiskSubdir, so the
// resulting on-disk package round-trips through ReadMemPackage cleanly.
func Download(pkgPath string, dst string, fetcher PackageFetcher) error {
	files, err := fetcher.FetchPackage(pkgPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o744); err != nil {
		return err
	}
	for _, file := range files {
		fdir := filepath.Join(dst, file.DiskSubdir())
		if err := os.MkdirAll(fdir, 0o755); err != nil {
			return fmt.Errorf("mkdir for %q: %w", fdir, err)
		}
		fileDst := filepath.Join(fdir, file.Name)
		if err := os.WriteFile(fileDst, []byte(file.Body), 0o644); err != nil {
			return fmt.Errorf("write file at %q: %w", fileDst, err)
		}
	}
	return nil
}
