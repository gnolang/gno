package main

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEstimateEnableAgainstARealChain runs the gas estimate gpao does before
// every approval against an actual node.
//
// Everything else about the estimate is tested against fakes: gasWantedFor and
// classifySimulate take values, not a chain. What none of them can tell us is
// whether simulating a MsgEnablePackage works at all -- the probe is signed for
// the chain's full Block.MaxGas, the ante refuses a gas-wanted above that
// rather than clamping, and the simulate has to execute the enable including
// the package's init(). Any of those could be wrong in a way no unit test sees.
//
// One key is both creator and approver. That conflates two roles the design
// keeps apart, but the estimate does not depend on the distinction and using
// one key keeps the genesis wiring to the point.
func TestEstimateEnableAgainstARealChain(t *testing.T) {
	const pkgPath = "gno.land/r/test/estimated"

	gnoroot := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(gnoroot)
	cfg.SkipGenesisSigVerification = true

	signer, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 0)
	require.NoError(t, err)
	info, err := signer.Info()
	require.NoError(t, err)
	who := info.GetAddress()

	// The chain runs "inert" and trusts this one key to approve.
	ggs := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
	ggs.Balances = []gnoland.Balance{{
		Address: who,
		Amount:  std.NewCoins(std.NewCoin("ugnot", 100_000_000_000)),
	}}
	ggs.VM.Params.CodeSubmissionPolicy = "inert"
	ggs.VM.Params.PkgApprovers = []crypto.Address{who}
	cfg.Genesis.AppState = ggs

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	defer node.Stop()

	rpc, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)
	client := gnoclient.Client{Signer: signer, RPCClient: rpc}

	// Submit, so something is actually parked to enable.
	mpkg := &std.MemPackage{
		Name: "estimated",
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "estimated.gno", Body: "package estimated\n\nvar n int\n\nfunc init() { n = 1 }\n\nfunc N(cur realm) int { return n }\n"},
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
		},
	}
	addTx := std.Tx{
		Msgs: []std.Msg{vm.MsgAddPackage{Creator: who, Package: mpkg}},
		Fee:  std.NewFee(20_000_000, std.MustParseCoin("1000000ugnot")),
	}
	signedAdd, err := client.SignTx(addTx, 0, 0)
	require.NoError(t, err)
	addRes, err := client.BroadcastTxCommit(signedAdd)
	require.NoError(t, err)
	require.True(t, addRes.CheckTx.IsOK(), "submit checkTx: %v", addRes.CheckTx.Error)
	require.True(t, addRes.DeliverTx.IsOK(), "submit deliverTx: %v", addRes.DeliverTx.Error)

	// The ceiling gpao would use, read the way gpao reads it.
	o := &oracle{client: client}
	maxGas, answered := o.queryBlockMaxGas(t.Context())
	require.True(t, answered)
	require.Positive(t, maxGas)

	simulateEnable := func(t *testing.T, hash string) (*abci.ResponseDeliverTx, error) {
		t.Helper()
		probe, err := client.SignTx(std.Tx{
			Msgs: []std.Msg{vm.MsgEnablePackage{Approver: who, PkgPath: pkgPath, PkgHash: hash}},
			Fee:  std.NewFee(maxGas, std.MustParseCoin("1000000ugnot")),
		}, 0, 0)
		require.NoError(t, err)
		return client.SimulateResult(probe)
	}

	t.Run("an enable that would succeed estimates", func(t *testing.T) {
		res, simErr := simulateEnable(t, vm.PackageContentHash(mpkg))
		require.Equal(t, verdictReady, classifySimulate(res, simErr),
			"a probe signed at the chain's own Block.MaxGas must be accepted and run")
		assert.Positive(t, res.GasUsed, "and report what the enable actually costs")

		want := gasWantedFor(res.GasUsed, 20_000_000, maxGas)
		assert.LessOrEqual(t, want, maxGas, "the sized gas-wanted must be sendable")
		assert.GreaterOrEqual(t, want, res.GasUsed, "and cover the measured cost")
	})

	t.Run("an enable that would fail is seen before broadcasting", func(t *testing.T) {
		// The payoff: gpao declines instead of paying a fee to be told this in
		// a block. A wrong hash is the cheapest way to make the enable fail for
		// a reason the chain reports rather than the ante.
		res, simErr := simulateEnable(t, "0000000000000000000000000000000000000000000000000000000000000000")
		assert.Equal(t, verdictWillFail, classifySimulate(res, simErr),
			"simulate must surface a failing enable as a verdict, not as an unreachable node")
	})
}
