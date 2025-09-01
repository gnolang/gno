package main

import (
	"fmt"
	"log/slog"
	gopath "path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

type varResolver []packages.Resolver

func (va varResolver) String() string {
	resolvers := packages.ChainedResolver(va)
	return resolvers.Name()
}

func (va *varResolver) Set(value string) error {
	name, location, found := strings.Cut(value, "=")
	if !found {
		return fmt.Errorf("invalid resolver format %q, should be `<name>=<location>`", value)
	}

	var res packages.Resolver
	switch name {
	case "remote":
		rpc, err := client.NewHTTPClient(location)
		if err != nil {
			return fmt.Errorf("invalid resolver remote: %q", location)
		}

		res = packages.NewRemoteResolver(location, rpc)
	case "root": // process everything from a root directory
		res = packages.NewRootResolver(location)
	case "local": // process a single directory
		path, ok := guessPathGnoMod(location)
		if !ok {
			return fmt.Errorf("unable to read module path from gnomod.toml in %q", location)
		}

		res = packages.NewLocalResolver(path, location)
	default:
		return fmt.Errorf("invalid resolver name: %q", name)
	}

	*va = append(*va, res)
	return nil
}

func setupPackagesResolver(logger *slog.Logger, cfg *AppConfig, dirs ...string) (packages.Resolver, []string) {
	// Add root resolvers
	localResolvers := make([]packages.Resolver, len(dirs))

	var paths []string
	for i, dir := range dirs {
		path := guessPath(cfg, dir)
		resolver := packages.NewLocalResolver(path, dir)

		if resolver.IsValid() {
			logger.Info("guessing directory path", "path", path, "dir", dir)
			paths = append(paths, path) // append local path
		} else {
			logger.Warn("no gno package found", "dir", dir)
		}

		localResolvers[i] = resolver
	}

	resolver := packages.ChainResolvers(
		packages.ChainResolvers(localResolvers...), // Resolve local directories
		packages.ChainResolvers(cfg.resolvers...),  // Use user's custom resolvers
	)

	// Enrich resolver with middleware
	return packages.MiddlewareResolver(resolver,
		packages.CacheMiddleware(func(pkg *packages.Package) bool {
			return pkg.Kind == packages.PackageKindRemote // Only cache remote package
		}),
		packages.FilterStdlibs,                    // Filter stdlib package from resolving
		packages.PackageCheckerMiddleware(logger), // Pre-check syntax to avoid bothering the node reloading on invalid files
		packages.LogMiddleware(logger),            // Log request
	), paths
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
