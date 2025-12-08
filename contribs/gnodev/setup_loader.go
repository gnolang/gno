package main

import (
	"log/slog"
	gopath "path"
	"path/filepath"
	"regexp"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

func setupPackagesLoader(logger *slog.Logger, cfg *AppConfig, dirs ...string) (packages.Loader, []string) {
	opts := []packages.NativeLoaderOption{
		packages.WithLogger(logger),
		packages.WithGnoRoot(cfg.root),
	}

	// Add extra workspaces (e.g., examples directory)
	examplesDir := filepath.Join(cfg.root, "examples")
	opts = append(opts, packages.WithExtraWorkspaces(examplesDir))

	// Add remote overrides from cfg.resolvers
	remoteOverrides := make(map[string]string)
	for _, r := range cfg.resolvers {
		// The resolver format is "remote=<url>" - we parse domain from the URL
		// For now, we skip this as we're removing remote resolvers
		// but we can add it back if needed
		_ = r
	}
	if len(remoteOverrides) > 0 {
		opts = append(opts, packages.WithRemoteOverrides(remoteOverrides))
	}

	loader := packages.NewNativeLoader(opts...)

	// Determine local paths from directories
	var paths []string
	for _, dir := range dirs {
		path := guessPath(cfg, dir)
		logger.Info("guessing directory path", "path", path, "dir", dir)
		paths = append(paths, path)
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

var reInvalidChar = regexp.MustCompile(`[^\w_-]`)

func guessPath(cfg *AppConfig, dir string) (path string) {
	if path, ok := guessPathGnoMod(dir); ok {
		return path
	}

	rname := reInvalidChar.ReplaceAllString(filepath.Base(dir), "-")
	return gopath.Join(cfg.chainDomain, "/r/dev/", rname)
}
