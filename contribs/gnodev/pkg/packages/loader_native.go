package packages

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	vmpackages "github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
)

type NativeLoader struct {
	index           *PathIndex
	logger          *slog.Logger
	gnoRoot         string
	extraWorkspaces []string
	remoteOverrides map[string]string // domain -> rpc URL
}

var _ Loader = (*NativeLoader)(nil)

type NativeLoaderConfig struct {
	Logger          *slog.Logger
	GnoRoot         string
	ExtraWorkspaces []string
	RemoteOverrides map[string]string
}

// logWriter wraps a logger as an io.Writer
type logWriter struct {
	logger *slog.Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logger.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}

func NewNativeLoader(cfg NativeLoaderConfig) *NativeLoader {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	gnoRoot := cfg.GnoRoot
	if gnoRoot == "" {
		gnoRoot = gnoenv.RootDir()
	}
	remoteOverrides := cfg.RemoteOverrides
	if remoteOverrides == nil {
		remoteOverrides = make(map[string]string)
	}
	return &NativeLoader{
		index:           NewPathIndex(),
		logger:          logger,
		gnoRoot:         gnoRoot,
		extraWorkspaces: cfg.ExtraWorkspaces,
		remoteOverrides: remoteOverrides,
	}
}

func (l *NativeLoader) Name() string {
	return "native"
}

func (l *NativeLoader) Load(patterns ...string) ([]*Package, error) {
	// Configure gnovm packages loader
	cfg := vmpackages.LoadConfig{
		Deps:                true,
		AllowEmpty:          true,
		Test:                true, // Load test file dependencies
		GnoRoot:             l.gnoRoot,
		ExtraWorkspaceRoots: l.extraWorkspaces,
		Out:                 &logWriter{l.logger},
		Fetcher:             rpcpkgfetcher.New(l.remoteOverrides),
	}

	// Call native loader
	pkgList, err := vmpackages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("native load failed: %w", err)
	}

	// Sort packages by dependencies
	sortedPkgs, err := pkgList.Sort()
	if err != nil {
		return nil, fmt.Errorf("failed to sort packages: %w", err)
	}

	// Filter out ignored packages
	sortedPkgs = sortedPkgs.GetNonIgnoredPkgs()

	// Convert to gnodev packages and populate index
	result := make([]*Package, 0, len(sortedPkgs))
	for _, vmPkg := range sortedPkgs {
		// Skip packages with errors
		if len(vmPkg.Errors) > 0 {
			for _, e := range vmPkg.Errors {
				l.logger.Warn("package error", "path", vmPkg.ImportPath, "error", e.Error())
			}
			continue
		}

		pkg := FromGnoVMPackage(vmPkg)
		l.index.Add(pkg)
		result = append(result, pkg)
	}

	l.logger.Info("packages loaded", "count", len(result))
	return result, nil
}

func (l *NativeLoader) Resolve(importPath string) (*Package, error) {
	// First check the pre-populated index
	if pkg, ok := l.index.GetByPath(importPath); ok {
		l.logger.Debug("resolved from index", "path", importPath)
		return pkg, nil
	}

	// Check if it's a stdlib
	if gnolang.IsStdlib(importPath) {
		dir := filepath.Join(l.gnoRoot, "gnovm", "stdlibs", filepath.FromSlash(importPath))
		if _, err := os.Stat(dir); err == nil {
			pkg := &Package{
				ImportPath: importPath,
				Dir:        dir,
				Kind:       PackageKindFS,
			}
			l.index.Add(pkg)
			l.logger.Debug("resolved stdlib", "path", importPath)
			return pkg, nil
		}
		return nil, ErrResolverPackageNotFound
	}

	// Not found in index - return error
	// Note: We don't fall back to Load() here because it requires workspace context
	// The index should be pre-populated via DiscoverPackages() for lazy loading to work
	l.logger.Debug("package not found in index", "path", importPath)
	return nil, ErrResolverPackageNotFound
}

// DiscoverPackages scans workspace roots and populates the index with all
// discoverable packages. This enables Resolve() to work for lazy loading
// by pre-mapping import paths to filesystem directories.
func (l *NativeLoader) DiscoverPackages() error {
	roots := l.extraWorkspaces

	l.logger.Debug("starting package discovery", "roots", roots)

	for _, root := range roots {
		root = filepath.Clean(root)
		if _, err := os.Stat(root); os.IsNotExist(err) {
			l.logger.Debug("workspace root does not exist, skipping", "root", root)
			continue
		}

		l.logger.Debug("walking workspace root", "root", root)

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // skip errors
			}

			// Skip non-files
			if d.IsDir() {
				// Skip sub-workspaces (they have their own gnowork.toml)
				if path != root {
					subwork := filepath.Join(path, "gnowork.toml")
					if _, err := os.Stat(subwork); err == nil {
						return fs.SkipDir
					}
				}
				return nil
			}

			// Look for gnomod.toml or gno.mod files
			name := d.Name()
			if name != "gnomod.toml" && name != "gno.mod" {
				return nil
			}

			dir := filepath.Dir(path)

			// Skip if we already have this directory indexed
			if _, ok := l.index.GetByDir(dir); ok {
				return nil
			}

			// Parse the gnomod to get the module path
			gm, err := gnomod.ParseDir(dir)
			if err != nil {
				l.logger.Info("skipping invalid gnomod", "dir", dir, "err", err)
				return nil // skip invalid
			}

			pkg := &Package{
				ImportPath: gm.Module,
				Dir:        dir,
				Kind:       PackageKindFS,
			}
			l.index.Add(pkg)

			return nil
		})
		if err != nil {
			l.logger.Warn("error walking workspace", "root", root, "err", err)
		}
	}

	l.logger.Info("packages discovered", "count", l.index.Len())
	return nil
}

// GetIndex returns the path index for external use (e.g., file watching)
func (l *NativeLoader) GetIndex() *PathIndex {
	return l.index
}

// FromGnoVMPackage converts a gnovm package to gnodev package
func FromGnoVMPackage(pkg *vmpackages.Package) *Package {
	kind := PackageKindFS

	// Determine if remote by checking if it's in modcache
	modCachePath := gnomod.ModCachePath()
	if strings.HasPrefix(filepath.Clean(pkg.Dir), modCachePath) {
		kind = PackageKindRemote
	}

	return &Package{
		ImportPath: pkg.ImportPath,
		Dir:        pkg.Dir,
		Kind:       kind,
		Name:       pkg.Name,
	}
}
