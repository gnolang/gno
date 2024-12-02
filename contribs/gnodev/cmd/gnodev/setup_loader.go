package main

import (
	"fmt"
	"go/scanner"
	"log/slog"
	"path/filepath"
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
			return fmt.Errorf("invalid resolver remote location: %q", location, name)
		}

		res = packages.Cache(packages.NewRemoteResolver(rpc))
	case "root":
		res = packages.NewRootResolver(location)
	// case "pkgdir":
	// 	res = packages.NewLo(location)
	default:
		return fmt.Errorf("invalid resolver name: %q", name)
	}

	*va = append(*va, res)
	return nil
}

func guessPathFromRoots(dir string, roots ...string) (path string, ok bool) {
	for _, root := range roots {
		if !strings.HasPrefix(dir, root) {
			continue
		}

		return strings.TrimPrefix(dir, root), true
	}

	return "", false
}

func guessPathGnoMod(dir string) (path string, ok bool) {
	modfile, err := gnomod.ParseAt(dir)
	if err == nil {
		return modfile.Module.Mod.Path, true

	}

	return "", false
}

func guessPath(cfg *devCfg, dir string) (path string, ok bool) {
	gnoroot := cfg.root
	if path, ok := guessPathGnoMod(dir); ok {
		return path, true
	}

	if path, ok = guessPathFromRoots(dir, gnoroot); ok {
		return path, ok
	}

	return "", false
}

func isStdPath(path string) bool {
	if i := strings.IndexRune(path, '/'); i > 0 {
		if j := strings.IndexRune(path[:i], '.'); i >= 0 {
			return false
		}
	}

	return true
}

func setupPackagesLoader(logger *slog.Logger, cfg *devCfg, path, dir string) (loader *packages.Loader) {
	gnoroot := cfg.root

	localresolver := packages.NewLocalResolver(path, dir)

	// Add root resolvers
	exampleRoot := filepath.Join(gnoroot, "examples")
	fsResolver := packages.NewRootResolver(exampleRoot)

	resolver := packages.ChainWithLogger(logger,
		localresolver,
		packages.ChainResolvers(cfg.resolvers...),
		fsResolver,
	)

	syntaxResolver := packages.SyntaxChecker(resolver, resolverErrorHandler(logger))
	return packages.NewLoaderWithFilter(isStdPath, syntaxResolver)
}

func resolverErrorHandler(logger *slog.Logger) packages.SyntaxErrorHandler {
	return func(path string, filename string, serr *scanner.Error) {
		logger.Error("syntax error",
			"path", path,
			"filename", filename,
			"err", serr.Error(),
		)
	}
}
