package gnoland

// Aggregate-load measurements for 0-fee sponsorship: behaviour under many
// invalid / non-sponsoring txs, and mempool / recheck behaviour under load.
//
// These drive the REAL clist mempool wired to the gno.land ABCI app through a
// local proxy client, so admission, mempool add/drop, and the recheck that
// fires on every block commit are the production paths, not a model.
//
//	go test ./pkg/gnoland/ -run 'TestSponsorship(LoadInvalidFlood|RecheckLoad)' -v

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/mempool"
	mpcfg "github.com/gnolang/gno/tm2/pkg/bft/mempool/config"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// sponsorRealm is a VALID sponsor: Do() calls PayGas then does bounded VM work,
// so a first-time CheckExecute is meaningfully more expensive than an ante-only
// recheck. That gap is the point of the recheck measurement.
const sponsorRealm = `package sponsor

import "chain/runtime"

func Do(cur realm) string {
	runtime.PayGas(5000000)
	x := 0
	for i := 0; i < 10000; i++ {
		x += i
	}
	return "ok"
}
`

// doomedRealm burns the whole window and never calls PayGas.
const doomedRealm = `package doomed

func Burn(cur realm) string {
	x := 0
	for i := 0; i < 1000000000; i++ {
		x += i
	}
	return "done"
}
`

type loadEnv struct {
	app     *sdk.BaseApp
	chainID string
	signers []ed25519.PrivKeyEd25519
	accNums []uint64
}

func setupLoadEnv(tb testing.TB, window int64, nSigners int) *loadEnv {
	tb.Helper()

	opts := TestAppOptions(memdb.NewMemDB())
	opts.AllowZeroFeeTxs = true
	app, err := NewAppWithOptions(opts)
	require.NoError(tb, err)
	bapp := app.(*sdk.BaseApp)

	deployer := crypto.AddressFromPreimage([]byte("load_deployer"))
	appState := DefaultGenState()
	appState.Balances = []Balance{
		{Address: deployer, Amount: []std.Coin{{Amount: 1e15, Denom: "ugnot"}}},
		// The sponsor realm holds the funds: it is what actually pays.
		{Address: gnolang.DerivePkgCryptoAddr("gno.land/r/bench/sponsor"), Amount: []std.Coin{{Amount: 1e14, Denom: "ugnot"}}},
	}

	signers := make([]ed25519.PrivKeyEd25519, nSigners)
	for i := range signers {
		signers[i] = ed25519.GenPrivKey()
		appState.Balances = append(appState.Balances, Balance{
			Address: signers[i].PubKey().Address(),
			Amount:  []std.Coin{{Amount: 1_000_000, Denom: "ugnot"}},
		})
	}
	appState.Txs = []TxWithMetadata{
		{Tx: mustDeployTx(deployer, "gno.land/r/bench/sponsor", "sponsor.gno", sponsorRealm)},
		{Tx: mustDeployTx(deployer, "gno.land/r/bench/doomed", "doomed.gno", doomedRealm)},
	}

	blockParams := defaultBlockParams()
	blockParams.MaxGasCreditPerTx = window

	chainID := "load"
	resp := bapp.InitChain(abci.RequestInitChain{
		Time:            time.Now(),
		ChainID:         chainID,
		ConsensusParams: &abci.ConsensusParams{Block: blockParams},
		AppState:        appState,
	})
	require.True(tb, resp.IsOK(), "InitChain: %v", resp)

	bapp.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: 1}})
	bapp.EndBlock(abci.RequestEndBlock{})
	bapp.Commit()

	// Genesis assigns account numbers in creation order, which is not the signer
	// index: query them rather than assuming.
	accNums := make([]uint64, nSigners)
	for i := range signers {
		q := bapp.Query(abci.RequestQuery{Path: "auth/accounts/" + signers[i].PubKey().Address().String()})
		require.True(tb, q.IsOK(), "account query: %v", q)
		var acc GnoAccount
		require.NoError(tb, amino.UnmarshalJSON(q.Data, &acc))
		accNums[i] = acc.GetAccountNumber()
	}

	return &loadEnv{app: bapp, chainID: chainID, signers: signers, accNums: accNums}
}

func (e *loadEnv) signedCall(tb testing.TB, signer int, pkgPath, fn string, gasWanted int64) []byte {
	tb.Helper()
	priv := e.signers[signer]
	tx := std.Tx{
		Msgs: []std.Msg{vm.NewMsgCall(priv.PubKey().Address(), nil, pkgPath, fn, nil)},
		Fee:  std.Fee{GasWanted: gasWanted, GasFee: std.Coin{Denom: "ugnot", Amount: 0}},
	}
	sb, err := tx.GetSignBytes(e.chainID, e.accNums[signer], 0)
	require.NoError(tb, err)
	sig, err := priv.Sign(sb)
	require.NoError(tb, err)
	tx.Signatures = []std.Signature{{PubKey: priv.PubKey(), Signature: sig}}
	return amino.MustMarshal(tx)
}

func (e *loadEnv) mempool(tb testing.TB) (*mempool.CListMempool, func()) {
	tb.Helper()
	cc := proxy.NewLocalClientCreator(e.app)
	appConn, err := cc.NewABCIClient()
	require.NoError(tb, err)
	require.NoError(tb, appConn.Start())
	cfg := mpcfg.TestMempoolConfig()
	cfg.Recheck = true
	return mempool.NewCListMempool(cfg, appConn, 1, 1_000_000), func() { appConn.Stop() }
}

// TestSponsorshipLoadInvalidFlood: a flood of 0-fee txs that burn the window and
// never call PayGas. Each is rejected, the mempool never grows, and each is
// replayable (rejection neither advances the sequence nor keeps the tx in the
// dedup cache). Reports the sustained rejection throughput.
func TestSponsorshipLoadInvalidFlood(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("-short")
	}

	const window = 10_000_000
	const n = 300

	env := setupLoadEnv(t, window, 1)
	mp, done := env.mempool(t)
	defer done()

	doomed := env.signedCall(t, 0, "gno.land/r/bench/doomed", "Burn", window)

	// Prove we measure VM work, not an ante bailout.
	probe := env.app.CheckTx(abci.RequestCheckTx{Tx: doomed})
	require.False(t, probe.IsOK(), "doomed tx must be rejected")
	pgas, pop := parseOOG(t, probe.Log)
	require.True(t, vmOps[pop], "must OOG in the VM (op=%s, gas=%d), not the ante", pop, pgas)

	start := time.Now()
	rejected := 0
	for range n {
		require.NoError(t, mp.CheckTx(doomed, func(res abci.Response) {
			if !res.(abci.ResponseCheckTx).IsOK() {
				rejected++
			}
		}))
	}
	elapsed := time.Since(start)

	require.Equal(t, n, rejected, "every doomed tx must be rejected")
	require.Equal(t, 0, mp.Size(), "rejected txs must NOT grow the mempool")

	t.Logf("invalid flood: %d doomed 0-fee txs rejected in %v (%v/tx, %.0f rejections/sec/node); mempool size=%d; gas burned/tx=%d",
		n, elapsed, elapsed/time.Duration(n), float64(n)/elapsed.Seconds(), mp.Size(), pgas)
}

// TestSponsorshipRecheckLoad: fill the mempool with valid pending sponsored txs,
// commit a block, and measure the recheck it triggers. A recheck must NOT re-run
// the VM, otherwise M pending 0-fee txs would cost M full VM executions on every
// block on every node.
func TestSponsorshipRecheckLoad(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("-short")
	}

	const window = 30_000_000 // large enough that Do's VM work completes rather than OOGing
	const m = 200

	env := setupLoadEnv(t, window, m)
	mp, done := env.mempool(t)
	defer done()

	// Pre-sign outside the timer so signing is not charged to admission.
	txs := make([][]byte, m)
	for i := range txs {
		txs[i] = env.signedCall(t, i, "gno.land/r/bench/sponsor", "Do", window)
	}

	admitStart := time.Now()
	for i := range m {
		var r abci.ResponseCheckTx
		require.NoError(t, mp.CheckTx(txs[i], func(res abci.Response) { r = res.(abci.ResponseCheckTx) }))
		require.True(t, r.IsOK(), "sponsored tx %d should be admitted: err=%v log=%.200s", i, r.Error, r.Log)
	}
	admit := time.Since(admitStart)
	require.Equal(t, m, mp.Size(), "all m sponsored txs should be pending")

	// A block commits WITHOUT these txs. This is the production ordering:
	// admission advances the sequence in checkState (CheckExecute persists it),
	// and Commit resets checkState so the pending txs re-validate on recheck.
	env.app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: env.chainID, Height: 2}})
	env.app.EndBlock(abci.RequestEndBlock{})
	env.app.Commit()

	recheckStart := time.Now()
	require.NoError(t, mp.Update(2, nil, nil, nil, 0))
	recheck := time.Since(recheckStart)

	require.Equal(t, m, mp.Size(), "valid txs survive recheck")

	admitPer, recheckPer := admit/time.Duration(m), recheck/time.Duration(m)
	t.Logf("recheck load: %d pending sponsored txs", m)
	t.Logf("  first-time admission (full VM): %v total, %v/tx", admit, admitPer)
	t.Logf("  recheck on block commit (ante-only): %v total, %v/tx", recheck, recheckPer)
	t.Logf("  ratio: %.2fx", float64(recheckPer)/float64(admitPer))

	// A regression to full-VM recheck would make recheck ~= admission, so a 2x
	// bound catches it while staying clear of timing noise.
	require.Less(t, recheck*2, admit,
		"recheck (%v) must be far cheaper than admission (%v)", recheck, admit)
}
