package gnoland

import (
	"sync"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// anteParams is what the ante wrapper reads from the params store for every
// transaction: the auth params, the vm params, and the gas config derived from
// them. Decoding them amino-unmarshals every field and bech32-decodes every
// allowlist entry, which made up over a third of a CheckTx.
type anteParams struct {
	auth   auth.Params
	vm     vm.Params
	gasCfg store.GasConfig
}

// checkTxParams memoizes anteParams for CheckTx, which includes mempool
// recheck, keyed on the block height of the CheckTx context.
//
// Why the height is a sufficient key: the CheckTx context reads the check
// state, which BaseApp replaces only in InitChain, Commit and on load, each
// time with the header of the block just committed. CheckTx runs the ante
// handler and never a message handler, so nothing writes params between two
// commits. A params change made by a transaction in block H is therefore
// visible to CheckTx exactly from the check state at height H, which misses
// this cache because the height moved.
//
// DeliverTx and simulate always load, so a params change inside a block is
// seen by the next transaction of that block, as before. Height 0 (genesis)
// always loads too.
//
// The cached values are shared between callers and must be treated as
// read-only.
type checkTxParams struct {
	load func(sdk.Context) anteParams

	mu     sync.Mutex
	height int64
	cached *anteParams
}

func (c *checkTxParams) get(ctx sdk.Context) anteParams {
	if ctx.Mode() != sdk.RunTxModeCheck || ctx.BlockHeight() <= 0 {
		return c.load(ctx)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cached != nil && c.height == ctx.BlockHeight() {
		return *c.cached
	}
	p := c.load(ctx)
	c.height, c.cached = ctx.BlockHeight(), &p
	return p
}
