// Package pkgdownload provides interfaces and utility functions to download gno packages files.
package pkgdownload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("resolve absolute path for dst %q: %w", dst, err)
	}

	for _, file := range files {
		fileDst := filepath.Join(dst, file.Name)
		absFileDst, err := filepath.Abs(fileDst)
		if err != nil {
			return fmt.Errorf("resolve absolute path for file %q: %w", fileDst, err)
		}
		// Ensure the resolved file path is contained within dst.
		// This prevents path traversal via file names containing ".." segments.
		if !strings.HasPrefix(absFileDst, absDst+string(filepath.Separator)) {
			return fmt.Errorf("path traversal detected: file name %q resolves outside destination directory", file.Name)
		}
		if err := os.WriteFile(fileDst, []byte(file.Body), 0o644); err != nil {
			return fmt.Errorf("write file at %q: %w", fileDst, err)
		}
	}

	return nil
}
