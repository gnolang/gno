package packages

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gofrs/flock"
)

func PackageDir(importPath string) string {
	return filepath.Join(gnomod.ModCachePath(), filepath.FromSlash(importPath))
}

// LockCache ensure the modcache dir exists, attempts to lock it and returns a filelock.
func LockCache(modCachePath string) (*flock.Flock, error) {
	if err := os.MkdirAll(modCachePath, 0o774); err != nil {
		return nil, fmt.Errorf("ensure modcache dir exists: %w", err)
	}

	flpath := filepath.Join(modCachePath, ".lock")
	fl := flock.New(flpath)
	// XXX: use TryLockContext to support concurrency instead of erroring out
	locked, err := fl.TryLock()
	if err != nil {
		return nil, fmt.Errorf("lock modcache: %w", err)
	}
	if !locked {
		return nil, fmt.Errorf("modcache already locked")
	}

	return fl, nil
}

// DownloadPackageToCache downloads a remote gno package by pkg path and store it in the modcache
func DownloadPackageToCache(out io.Writer, pkgPath string, fetcher pkgdownload.PackageFetcher) error {
	modCachePath := gnomod.ModCachePath()

	fl, err := LockCache(modCachePath)
	if err != nil {
		return err
	}
	defer fl.Unlock()

	markersDir := filepath.Join(modCachePath, ".markers")
	if err := os.MkdirAll(markersDir, 0o744); err != nil {
		return fmt.Errorf("ensure .markers dir exists: %w", err)
	}
	markerFile := filepath.Join(markersDir, gnolang.DerivePkgBech32Addr(pkgPath).String())

	if _, err := os.Stat(markerFile); err == nil {
		// package exists in modcache, do nothing
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat marker file for package %q at %q: %w", pkgPath, markerFile, err)
	}

	fmt.Fprintf(out, "gno: downloading %s\n", pkgPath)

	dst := filepath.Join(modCachePath, filepath.FromSlash(pkgPath))
	if err := pkgdownload.Download(pkgPath, dst, fetcher); err != nil {
		return err
	}

	// mark package as downloaded
	if err := os.WriteFile(markerFile, nil, 0o644); err != nil {
		return fmt.Errorf("write marker file: %w", err)
	}

	return nil
}
