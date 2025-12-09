package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

func setupPackagesLoader(logger *slog.Logger, cfg *AppConfig, dirs ...string) (packages.Loader, []string) {
	opts := []packages.NativeLoaderOption{
		packages.WithLogger(logger),
		packages.WithGnoRoot(cfg.root),
	}

	// Add extra workspaces (e.g., examples directory and user-provided directories)
	examplesDir := filepath.Join(cfg.root, "examples")
	workspaces := append([]string{examplesDir}, dirs...)
	opts = append(opts, packages.WithExtraWorkspaces(workspaces...))

	// Warn about deprecated resolver flag
	if len(cfg.resolvers) > 0 {
		logger.Warn("the -resolver flag is deprecated and ignored; packages are now discovered via gnomod.toml and gnowork.toml")
	}

	loader := packages.NewNativeLoader(opts...)

	// Pre-populate the index for lazy loading support
	// This scans workspace roots and maps import paths to filesystem directories
	if err := loader.DiscoverPackages(); err != nil {
		logger.Warn("failed to discover packages", "err", err)
	}

	// Determine local paths from directories
	// - If dir has gnomod.toml -> it's a package, add its path
	// - If dir has gnowork.toml -> it's a workspace, use for discovery only
	// - Otherwise -> use for discovery only
	var paths []string
	for _, dir := range dirs {
		if path, ok := guessPathGnoMod(dir); ok {
			logger.Info("package directory detected", "path", path, "dir", dir)
			paths = append(paths, path)
		} else if isWorkspaceDir(dir) {
			logger.Debug("workspace directory detected, using for discovery only", "dir", dir)
		} else {
			logger.Debug("directory has no gnomod/gnowork, using for discovery only", "dir", dir)
		}
	}

	return loader, paths
}

func guessPathGnoMod(dir string) (path string, ok bool) {
	modfile, err := gnomod.ParseDir(dir)
	if err != nil {
		return "", false
	}
	return modfile.Module, true
}

func isWorkspaceDir(dir string) bool {
	workFile := filepath.Join(dir, "gnowork.toml")
	_, err := os.Stat(workFile)
	return err == nil
}
