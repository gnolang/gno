package main

import (
	"fmt"
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

