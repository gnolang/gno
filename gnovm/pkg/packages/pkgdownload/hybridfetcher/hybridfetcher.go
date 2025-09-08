package hybridfetcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/examplespkgfetcher"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// HybridFetcher tries to fetch packages from local directories first,
// then falls back to RPC download if not found locally.
type HybridFetcher struct {
	localFetchers []pkgdownload.PackageFetcher
	rpcFetcher    pkgdownload.PackageFetcher
	verbose       bool
}

// New creates a new HybridFetcher that checks local directories before falling back to RPC.
func New(remoteOverrides map[string]string, verbose bool) pkgdownload.PackageFetcher {
	hf := &HybridFetcher{
		localFetchers: []pkgdownload.PackageFetcher{},
		rpcFetcher:    rpcpkgfetcher.New(remoteOverrides),
		verbose:       verbose,
	}

	// Add examples fetcher
	hf.localFetchers = append(hf.localFetchers, examplespkgfetcher.New(""))

	// Add local workspace fetcher if GNOROOT is set
	if gnoRoot := gnoenv.RootDir(); gnoRoot != "" {
		// Check common local directories
		localDirs := []string{
			filepath.Join(gnoRoot, "examples"),
			filepath.Join(gnoRoot, "gno.land"),
		}

		for _, dir := range localDirs {
			if _, err := os.Stat(dir); err == nil {
				hf.localFetchers = append(hf.localFetchers, &localDirFetcher{baseDir: dir})
			}
		}
	}

	// Add current working directory fetcher
	if cwd, err := os.Getwd(); err == nil {
		hf.localFetchers = append(hf.localFetchers, &localDirFetcher{baseDir: cwd})
	}

	return hf
}

// FetchPackage tries local fetchers first, then falls back to RPC.
func (hf *HybridFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	if hf.verbose {
		fmt.Println("gno: trying to fetch", pkgPath, "locally")
	}

	// Try local fetchers first
	for _, fetcher := range hf.localFetchers {
		files, err := fetcher.FetchPackage(pkgPath)
		if err == nil {
			if hf.verbose {
				fmt.Println("gno: found", pkgPath, "locally")
			}
			return files, nil
		}
	}

	// Fall back to RPC
	if hf.verbose {
		fmt.Println("gno: downloading", pkgPath)
	}
	return hf.rpcFetcher.FetchPackage(pkgPath)
}

// localDirFetcher fetches packages from a local directory.
type localDirFetcher struct {
	baseDir string
}

func (lf *localDirFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	// Convert package path to potential local paths
	possiblePaths := []string{
		filepath.Join(lf.baseDir, pkgPath),
	}

	// Handle gno.land/p/* and gno.land/r/* paths
	if strings.HasPrefix(pkgPath, "gno.land/p/") && len(pkgPath) > 11 {
		possiblePaths = append(possiblePaths, filepath.Join(lf.baseDir, "p", pkgPath[11:]))
	} else if strings.HasPrefix(pkgPath, "gno.land/r/") && len(pkgPath) > 11 {
		possiblePaths = append(possiblePaths, filepath.Join(lf.baseDir, "r", pkgPath[11:]))
	}

	for _, localPath := range possiblePaths {
		if _, err := os.Stat(localPath); err != nil {
			continue
		}

		// Try to read the package
		memPkg, err := gnolang.ReadMemPackage(localPath, pkgPath, gnolang.MPAnyAll)
		if err != nil {
			continue
		}

		return memPkg.Files, nil
	}

	return nil, fmt.Errorf("package %s not found in %s", pkgPath, lf.baseDir)
}
