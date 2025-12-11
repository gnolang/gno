package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

func setupPackagesLoader(logger *slog.Logger, cfg *AppConfig, dirs ...string) (packages.Loader, []string) {
	// Add extra workspaces (e.g., examples directory and user-provided directories)
	examplesDir := filepath.Join(cfg.root, "examples")
	workspaces := append([]string{examplesDir}, dirs...)

	// Warn about deprecated resolver flag
	if len(cfg.resolvers) > 0 {
		logger.Warn("the -resolver flag is deprecated and ignored; packages are now discovered via gnomod.toml and gnowork.toml")
	}

	loader := packages.NewNativeLoader(packages.NativeLoaderConfig{
		Logger:          logger,
		GnoRoot:         cfg.root,
		ExtraWorkspaces: workspaces,
	})

	// Pre-populate the index for lazy loading support
	// This scans workspace roots and maps import paths to filesystem directories
	if err := loader.DiscoverPackages(); err != nil {
		logger.Warn("failed to discover packages", "err", err)
	}

	// Determine paths to pre-load based on load mode
	var paths []string
	switch cfg.loadMode {
	case LoadModeAuto:
		// If in examples folder, use lazy mode (no pre-load)
		if isInExamplesDir(cfg.root, dirs) {
			logger.Info("running from examples folder, using lazy loading")
			break
		}

		examplesDir := filepath.Join(cfg.root, "examples")

		for _, dir := range dirs {
			absDir, err := filepath.Abs(dir)
			if err != nil {
				continue
			}

			if isWorkspaceDir(dir) {
				// Workspace detected: pre-load ALL packages within this workspace
				// by filtering the discovered index by directory prefix
				logger.Info("workspace detected, will pre-load all packages", "dir", dir)
				for _, pkg := range loader.GetIndex().List() {
					// Skip examples packages
					if strings.HasPrefix(pkg.Dir, examplesDir) {
						continue
					}
					// Only include packages under this workspace
					if strings.HasPrefix(pkg.Dir, absDir) {
						logger.Debug("workspace package detected", "path", pkg.ImportPath)
						paths = append(paths, pkg.ImportPath)
					}
				}
			} else if pkgPath, ok := guessPathGnoMod(dir); ok {
				// Single package detected
				logger.Info("package detected, will be pre-loaded", "path", pkgPath, "dir", dir)
				paths = append(paths, pkgPath)
			}
		}
	case LoadModeLazy:
		logger.Info("lazy mode: packages will be loaded on-demand")
	case LoadModeFull:
		// Pre-load all discovered packages
		for _, pkg := range loader.GetIndex().List() {
			paths = append(paths, pkg.ImportPath)
		}
		logger.Info("full mode: pre-loading all discovered packages", "count", len(paths))
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

func isInExamplesDir(gnoRoot string, dirs []string) bool {
	examplesDir := filepath.Join(gnoRoot, "examples")
	for _, dir := range dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absDir, examplesDir) {
			return true
		}
	}
	return false
}
