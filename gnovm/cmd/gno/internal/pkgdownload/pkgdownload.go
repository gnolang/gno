// Package pkgdownload provides interfaces and utility functions to download gno packages files.
package pkgdownload

import (
	"fmt"
	"os"
	"path/filepath"
)

// Download downloads the package identified by `pkgPath` in the directory at `dst` using the provided [PackageFetcher].
// The directory at `dst` is created if it does not exists.
func Download(pkgPath string, dst string, fetcher PackageFetcher) error {
	files, err := fetcher.FetchPackage(pkgPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0o744); err != nil {
		return err
	}

	for _, file := range files {
		fileDst := filepath.Join(dst, file.Name)
		if err := os.WriteFile(fileDst, []byte(file.Body), 0o644); err != nil {
			return fmt.Errorf("write file at %q: %w", fileDst, err)
		}
	}

	return nil
}
