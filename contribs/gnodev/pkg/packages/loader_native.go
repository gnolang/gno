package packages

import (
	"fmt"
	"io"
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
	out             io.Writer
}

type NativeLoaderOption func(*NativeLoader)

func WithLogger(logger *slog.Logger) NativeLoaderOption {
	return func(l *NativeLoader) { l.logger = logger }
}

func WithGnoRoot(root string) NativeLoaderOption {
	return func(l *NativeLoader) { l.gnoRoot = root }
}

func WithExtraWorkspaces(roots ...string) NativeLoaderOption {
	return func(l *NativeLoader) { l.extraWorkspaces = roots }
}

func WithRemoteOverrides(overrides map[string]string) NativeLoaderOption {
	return func(l *NativeLoader) { l.remoteOverrides = overrides }
}

func WithOutput(out io.Writer) NativeLoaderOption {
	return func(l *NativeLoader) { l.out = out }
}

func NewNativeLoader(opts ...NativeLoaderOption) *NativeLoader {
	l := &NativeLoader{
		index:           NewPathIndex(),
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		gnoRoot:         gnoenv.RootDir(),
		remoteOverrides: make(map[string]string),
		out:             os.Stdout,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *NativeLoader) Name() string {
	return "native"
}

func (l *NativeLoader) Load(patterns ...string) ([]*Package, error) {
	// Configure gnovm packages loader
	cfg := vmpackages.LoadConfig{
		Deps:                true,
		AllowEmpty:          true,
		GnoRoot:             l.gnoRoot,
		ExtraWorkspaceRoots: l.extraWorkspaces,
		Out:                 l.out,
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
	// First check the index
	if pkg, ok := l.index.GetByPath(importPath); ok {
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
			return pkg, nil
		}
		return nil, fmt.Errorf("stdlib %s not found", importPath)
	}

	// Try to load via gnovm packages
	pkgs, err := l.Load(importPath)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, ErrResolverPackageNotFound
	}

	// Find the matching package
	for _, pkg := range pkgs {
		if pkg.ImportPath == importPath {
			return pkg, nil
		}
	}

	return nil, ErrResolverPackageNotFound
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
