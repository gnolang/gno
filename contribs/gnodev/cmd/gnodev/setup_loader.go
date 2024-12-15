package main

import (
	"fmt"
	"log/slog"
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

		res = packages.NewRemoteResolver(rpc)
	case "root":
		res = packages.NewFSResolver(location)
	case "pkgdir":
		path, ok := guessPathGnoMod(location)
		if !ok {
			return fmt.Errorf("unable to read module path from gno.mod in %q", location)
		}

		res = packages.NewLocalResolver(path, location)
	default:
		return fmt.Errorf("invalid resolver name: %q", name)
	}

	*va = append(*va, res)
	return nil
}

func setupPackagesResolver(logger *slog.Logger, cfg *devCfg, dirs ...string) (packages.Resolver, []string) {
	// Add root resolvers
	exampleRoot := filepath.Join(cfg.root, "examples")

	var localResolvers []packages.Resolver
	var paths []string
	for _, dir := range dirs {
		path := guessPath(cfg, dir)
		localResolvers = append(localResolvers, packages.NewLocalResolver(path, dir))
		paths = append(paths, path)

		logger.Info("guessing directory path", "path", path, "dir", dir)
	}

	resolver := packages.ChainResolvers(
		packages.ChainResolvers(localResolvers...), // Resolve local directories
		packages.ChainResolvers(cfg.resolvers...),  // Use user's custom resolvers
		packages.NewFSResolver(exampleRoot),        // Ultimately use fs resolver
	)

	// Enrich resolver with middleware
	return packages.MiddlewareResolver(resolver,
		packages.CacheMiddleware(func(pkg *packages.Package) bool {
			return pkg.Kind == packages.PackageKindRemote // Cache only remote package
		}),
		packages.FilterPathMiddleware("stdlib", isStdPath), // Filter stdlib package from resolving
		packages.PackageCheckerMiddleware(logger),          // Pre-check syntax to avoid bothering the node reloading on invalid files
		packages.LogMiddleware(logger),                     // Log any request
	), paths
}

func guessPathGnoMod(dir string) (path string, ok bool) {
	modfile, err := gnomod.ParseAt(dir)
	if err == nil {
		return modfile.Module.Mod.Path, true
	}

	return "", false
}

var reInvalidChar = regexp.MustCompile(`[^\w_-]`)

func guessPath(cfg *devCfg, dir string) (path string) {
	if path, ok := guessPathGnoMod(dir); ok {
		return path
	}

	rname := reInvalidChar.ReplaceAllString(filepath.Base(dir), "-")
	return filepath.Join(cfg.chainDomain, "/r/dev/", rname)
}

func isStdPath(path string) bool {
	if i := strings.IndexRune(path, '/'); i > 0 {
		if j := strings.IndexRune(path[:i], '.'); j >= 0 {
			return false
		}
	}

	return true
}

// func guessPathFromRoots(dir string, roots ...string) (path string, ok bool) {
// 	for _, root := range roots {
// 		if !strings.HasPrefix(dir, root) {
// 			continue
// 		}

// 		return strings.TrimPrefix(dir, root), true
// 	}

// 	return "", false
// }
