package main

import (
	"fmt"
	"go/scanner"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
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

func setupPackagesLoader(logger *slog.Logger, cfg *devCfg, dir string) (*packages.Loader, error) {
	// Setup first resolver for the current package directory
	gnoroot := cfg.root
	localresolver, path := packages.GuessLocalResolverGnoMod(dir)
	if localresolver == nil {
		localresolver, path = packages.GuessLocalResolverFromRoots(dir, []string{gnoroot})
		if localresolver == nil {
			return nil, fmt.Errorf("unable to guess current package")
		}
	}

	// Add root resolvers
	exampleRoot := filepath.Join(gnoroot, "examples")
	fsResolver := packages.NewRootResolver(exampleRoot)

	resolver := packages.ChainWithLogger(logger,
		localresolver,
		packages.ChainResolvers(cfg.resolvers...),
		fsResolver,
	)

	return &packages.Loader{
		Paths:    []string{path},
		Resolver: packages.SyntaxChecker(resolver, resolverErrorHandler(logger)),
	}, nil
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
