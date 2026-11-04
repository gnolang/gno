package main

import (
	"bytes"
	"context"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// importCache holds the genesis packages fetched for earlier verifications, so
// a later one asks the node only for what none of them fetched. The daemon owns
// it and hands each child its share in an importSeed.
//
// Only a public package added at genesis is kept: that is the chain's own
// content, of a size no submitter controls, and AddPackage refuses an existing
// public path, so what a node served once it serves for the life of the chain.
// A package added since is asked for again every time, since anyone can get
// one approved and imported and keeping it would let them grow this cache
// without bound. So is a private package, which can be redeployed under its
// path, and an absent one, which can be enabled.
type importCache map[string]*std.MemPackage

// genesisImports returns the import cache as the last verification left it.
// Verifier goroutine only.
func (o *oracle) genesisImports() importCache {
	return o.imports
}

// dropImportsFromAnotherChain empties the import cache when block 1 is not the
// one it was filled under: a chain started again under the same ID has a
// genesis of its own, whatever its tip did meanwhile. A node that does not give
// block 1 empties it too, since nothing then says which chain it came from.
// Verifier goroutine only.
func (o *oracle) dropImportsFromAnotherChain(ctx context.Context) {
	var first []byte
	if res, err := o.client.RPCClient.BlockchainInfo(ctx, 1, 1); err == nil &&
		len(res.BlockMetas) == 1 && res.BlockMetas[0] != nil {
		first = res.BlockMetas[0].BlockID.Hash
	}
	if first == nil || !bytes.Equal(first, o.importsChain) {
		o.imports = importCache{}
	}
	o.importsChain = first
}

// closure walks mpkg's chain import closure, from the cache where it holds a
// package and from g otherwise, keeping what g fetches. It returns the packages
// found and the paths the node answered it does not serve. The walk ends at
// g's first transport fault: the child asks the node for whatever it is not
// handed, and the child is what judges a miss.
func (c importCache) closure(g *rpcGetter, mpkg *std.MemPackage) (found []*std.MemPackage, absent []string) {
	_ = walkImports(mpkg, func(path, _ string) (*std.MemPackage, error) {
		if !gno.IsUserlib(path) {
			return nil, nil
		}
		dep, ok := c[path]
		if !ok {
			dep = g.GetMemPackage(path)
			if g.transportErr != nil {
				return nil, g.transportErr
			}
			if dep == nil {
				absent = append(absent, path)
				return nil, nil
			}
			// The keeper stamps the submitting block's height over whatever
			// [addpkg] the submitter wrote, and genesis runs at height 0. It
			// also parsed the gnomod to store the package, so a parse error
			// here is not expected; such a package is only fetched again.
			if gm, err := gnomod.ParseMemPackage(dep); err == nil && gm.AddPkg.Height == 0 && !gm.Private {
				c[path] = dep
			}
		}
		found = append(found, dep)
		return dep, nil
	})
	return found, absent
}
