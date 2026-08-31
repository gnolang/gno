package main

import (
	"io"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpentMovesOnlyWhenATransactionIsSent pins the accounting to the money.
//
// The ante handler deducts the fee for any transaction that reaches the chain,
// and for nothing else. An enable whose simulate says verdictWillFail returns
// before BroadcastTxCommit -- no transaction exists, nothing is charged -- so
// counting it against -max-spend makes the oracle report itself out of budget
// while holding every coin it started with, and refuse legitimate packages for
// want of funds it still has.
//
// The failing arm uses a package whose init() panics: it passes verification
// (the verifier preprocesses rather than executes) but EnablePackage runs
// init(), so the simulate probe fails -- the cheapest reproduction of "verified
// clean, enable would fail".
func TestSpentMovesOnlyWhenATransactionIsSent(t *testing.T) {
	gnoroot := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(gnoroot)
	cfg.SkipGenesisSigVerification = true

	signer, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 0)
	require.NoError(t, err)
	info, err := signer.Info()
	require.NoError(t, err)
	who := info.GetAddress()

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

	park := func(t *testing.T, mpkg *std.MemPackage) {
		t.Helper()
		tx := std.Tx{
			Msgs: []std.Msg{vm.MsgAddPackage{Creator: who, Package: mpkg}},
			Fee:  std.NewFee(20_000_000, std.MustParseCoin("1000000ugnot")),
		}
		signed, err := client.SignTx(tx, 0, 0)
		require.NoError(t, err)
		res, err := client.BroadcastTxCommit(signed)
		require.NoError(t, err)
		require.True(t, res.CheckTx.IsOK(), "park checkTx: %v", res.CheckTx.Error)
		require.True(t, res.DeliverTx.IsOK(), "park deliverTx: %v", res.DeliverTx.Error)
	}

	const badPath = "gno.land/r/test/tollbad"
	bad := &std.MemPackage{
		Name: "tollbad",
		Path: badPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(badPath)},
			{Name: "tollbad.gno", Body: "package tollbad\n\nfunc init() { panic(\"never enables\") }\n\nfunc F(cur realm) {}\n"},
		},
	}
	const goodPath = "gno.land/r/test/tollgood"
	good := &std.MemPackage{
		Name: "tollgood",
		Path: goodPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(goodPath)},
			{Name: "tollgood.gno", Body: "package tollgood\n\nfunc F(cur realm) string { return \"ok\" }\n"},
		},
	}
	park(t, bad)
	park(t, good)

	ocfg := config{
		remote:       remote,
		chainID:      cfg.Genesis.ChainID,
		mnemonic:     integration.DefaultAccount_Seed,
		gnoRoot:      gnoroot,
		gasFee:       defaultGasFee,
		gasWanted:    defaultGasWanted,
		verifyBudget: time.Minute,
	}
	o, err := newOracle(ocfg, testIO(t))
	require.NoError(t, err)
	o.blockMaxGas = o.queryBlockMaxGas(t.Context())

	// Verified clean, refused at simulate, nothing broadcast: the counter must
	// not move, because the money did not.
	o.handleCandidate(t.Context(), bad)
	badStatus := o.status.get(badPath)
	require.Equal(t, statusPending, badStatus.Status)
	require.Contains(t, badStatus.Reason, "simulate says the enable would fail",
		"the arm has to reach the simulate refusal, not stop earlier at verification")
	require.Zero(t, o.spent,
		"no transaction was broadcast and the ante charged nothing; the counter disagrees with the money")

	// Control arm: a broadcast approval costs exactly one fee, counted at the
	// send.
	o.handleCandidate(t.Context(), good)
	assert.Equal(t, o.enableFee, o.spent,
		"a broadcast approval must be counted, whether or not it succeeds -- the ante charges for it")
	assert.Equal(t, statusApproved, o.status.get(goodPath).Status)
}

// TestSpentIsRefundedWhenCheckTxRejects covers the other half of the same rule:
// the debit is taken before BroadcastTxCommit, so the one failure that costs
// nothing has to give it back.
//
// CheckTx runs the ante handler and aborts on the first refusal, before the fee
// is deducted, and the transaction never enters a block -- nothing was charged.
// A DeliverTx failure did run in a block and was charged, so it stays counted,
// and so does a transport error, which may have committed with the answer lost.
// Only the CheckTx arm is refunded, and an unrefunded one walks the counter away
// from the balance one rejection at a time.
//
// The rejection is provoked with a gas fee of 1ugnot against this chain's block
// gas price of 1ugnot per 1000 gas. Simulate skips two of the ante handler's
// checks -- the mempool fee and the signature -- but RequireSigForSimulate puts
// the signature back for MsgEnablePackage (tm2/pkg/sdk/auth/ante.go), leaving
// the fee as the only gate this transaction clears at simulate and fails at
// CheckTx: the arm under test, without contriving anything else.
func TestSpentIsRefundedWhenCheckTxRejects(t *testing.T) {
	gnoroot := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(gnoroot)
	cfg.SkipGenesisSigVerification = true

	signer, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 0)
	require.NoError(t, err)
	info, err := signer.Info()
	require.NoError(t, err)
	who := info.GetAddress()

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

	const pkgPath = "gno.land/r/test/tollrefund"
	mpkg := &std.MemPackage{
		Name: "tollrefund",
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
			{Name: "tollrefund.gno", Body: "package tollrefund\n\nfunc F(cur realm) string { return \"ok\" }\n"},
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
	require.True(t, addRes.CheckTx.IsOK(), "park checkTx: %v", addRes.CheckTx.Error)
	require.True(t, addRes.DeliverTx.IsOK(), "park deliverTx: %v", addRes.DeliverTx.Error)

	ocfg := config{
		remote:   remote,
		chainID:  cfg.Genesis.ChainID,
		mnemonic: integration.DefaultAccount_Seed,
		gnoRoot:  gnoroot,
		// Underpriced on purpose: one ugnot for a whole enable, on a chain that
		// asks one per 1000 gas.
		gasFee:       "1ugnot",
		gasWanted:    defaultGasWanted,
		verifyBudget: time.Minute,
	}
	o, err := newOracle(ocfg, testIO(t))
	require.NoError(t, err)
	o.blockMaxGas = o.queryBlockMaxGas(t.Context())

	before, _, err := client.QueryBalance(who)
	require.NoError(t, err)

	o.handleCandidate(t.Context(), mpkg)

	status := o.status.get(pkgPath)
	require.Equal(t, statusPending, status.Status)
	require.Contains(t, status.Reason, "broadcast:",
		"the arm has to reach BroadcastTxCommit; an earlier return leaves the counter alone for another reason")

	after, _, err := client.QueryBalance(who)
	require.NoError(t, err)
	require.Equal(t, before, after,
		"the refusal has to be CheckTx's: a DeliverTx failure would have run in a block and been charged")
	assert.Zero(t, o.spent,
		"the ante deducted nothing, so the debit taken before the broadcast has to be given back")
}

// TestBroadcastWasFree covers the discriminator neither test above can reach:
// which broadcast outcome gives the debit back.
//
// The node tests each exercise one arm, and both stay green if the rule is
// widened to "always refund" -- neither one distinguishes the outcomes it does
// not provoke. All four are named here, without a chain.
func TestBroadcastWasFree(t *testing.T) {
	refused := abci.StringError("insufficient fee")

	for _, tc := range []struct {
		name string
		res  *ctypes.ResultBroadcastTxCommit
		want bool
	}{
		{
			// The dangerous one. gnoclient.BroadcastTxCommit returns no result
			// for a lost answer and for every pre-mempool refusal alike, and a
			// refusal that never reached a block cannot be told apart from one
			// that may have committed, so the fee stays counted.
			name: "nothing came back",
			want: false,
		},
		{
			name: "CheckTx refused it before any block",
			res: &ctypes.ResultBroadcastTxCommit{
				CheckTx: abci.ResponseCheckTx{ResponseBase: abci.ResponseBase{Error: refused}},
			},
			want: true,
		},
		{
			name: "DeliverTx failed inside a block",
			res: &ctypes.ResultBroadcastTxCommit{
				DeliverTx: abci.ResponseDeliverTx{ResponseBase: abci.ResponseBase{Error: refused}},
			},
			want: false,
		},
		{
			name: "it committed",
			res:  &ctypes.ResultBroadcastTxCommit{Height: 42},
			want: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, broadcastWasFree(tc.res))
		})
	}
}

// testIO builds a quiet commands.IO with real (non-nil) writers, which the
// child-teeing path in verify() requires.
func testIO(t *testing.T) commands.IO {
	t.Helper()
	tio := commands.NewTestIO()
	tio.SetOut(commands.WriteNopCloser(io.Discard))
	tio.SetErr(commands.WriteNopCloser(io.Discard))
	return tio
}
