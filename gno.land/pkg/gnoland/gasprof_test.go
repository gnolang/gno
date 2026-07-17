package gnoland

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// Phase 3b: the dev-only .app/profiletx ABCI query runs a tx through Simulate
// with gas profiling and returns a pprof profile of its gas usage.
func TestApp_gasProfileQuery(t *testing.T) {
	t.Parallel()

	opts := TestAppOptions(memdb.NewMemDB())
	opts.EnableGasProfiler = true
	app, err := NewAppWithOptions(opts)
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)

	// Deploy a realm at genesis and fund the caller.
	addr := crypto.AddressFromPreimage([]byte("test1"))
	appState := DefaultGenState()
	appState.Balances = []Balance{{Address: addr, Amount: []std.Coin{{Amount: 1e15, Denom: "ugnot"}}}}
	appState.Txs = []TxWithMetadata{{Tx: std.Tx{
		Msgs: []std.Msg{vm.NewMsgAddPackage(addr, "gno.land/r/demo", []*std.MemFile{
			{Name: "demo.gno", Body: "package demo\nfunc Hello(cur realm) string { return `hello` }"},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/demo")},
		})},
		Fee:        std.Fee{GasWanted: 1e6, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}},
		Signatures: []std.Signature{{}},
	}}}
	resp := bapp.InitChain(abci.RequestInitChain{
		Time:            time.Now(),
		ChainID:         "dev",
		ConsensusParams: &abci.ConsensusParams{Block: defaultBlockParams()},
		Validators:      []abci.ValidatorUpdate{},
		AppState:        appState,
	})
	require.True(t, resp.IsOK(), "InitChain: %v", resp)
	bapp.Commit() // commit the genesis realm so Simulate (checkState) sees it

	tx := amino.MustMarshal(std.Tx{
		Msgs:       []std.Msg{vm.NewMsgCall(addr, nil, "gno.land/r/demo", "Hello", nil)},
		Fee:        std.Fee{GasWanted: 1_000_000, GasFee: std.Coin{Denom: "ugnot", Amount: 1_000_000}},
		Signatures: []std.Signature{{}},
	})

	qres := bapp.Query(abci.RequestQuery{Path: ".app/profiletx", Data: tx})
	require.True(t, qres.IsOK(), "profiletx query failed: %v", qres.Error)
	require.NotEmpty(t, qres.Value)
	require.Equal(t, "ok", qres.Log, "tx completed, profile is not partial")

	// The value is a gzipped pprof profile; decompressed, it names the profiled
	// function in its string table. This name — pkgpath.Func concatenated — does
	// NOT appear in the amino tx bytes (separate fields), so it genuinely proves
	// a call frame was recorded, not an echo of the input.
	gz, err := gzip.NewReader(bytes.NewReader(qres.Value))
	require.NoError(t, err)
	raw, err := io.ReadAll(gz)
	require.NoError(t, err)
	require.Contains(t, string(raw), "gno.land/r/demo.Hello")

	// A malformed tx through the ENABLED profiler must fail gracefully, not panic.
	bad := bapp.Query(abci.RequestQuery{Path: ".app/profiletx", Data: []byte("not-amino")})
	require.False(t, bad.IsOK(), "malformed tx must return an error")
	require.NotNil(t, bad.Error)
}

// Regression for the master merge of #5431: Simulate splits into a pre-first-
// commit fallback and a committed-height snapshot path, and the gas-profiler
// ctxFn must reach BOTH. TestApp_gasProfileQuery above only exercises the
// fallback (it queries before advancing past height 1). Here the chain is
// advanced past height 1 first, so .app/profiletx runs through the snapshot
// path a real node uses — where the profiler would silently record nothing if
// the ctxFn were applied to only one branch.
func TestApp_gasProfileQuery_pastFirstCommit(t *testing.T) {
	t.Parallel()

	opts := TestAppOptions(memdb.NewMemDB())
	opts.EnableGasProfiler = true
	app, err := NewAppWithOptions(opts)
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)

	key := getDummyKey(t)
	addr := key.PubKey().Address()
	chainID := "dev"

	appState := DefaultGenState()
	appState.Balances = []Balance{{Address: addr, Amount: []std.Coin{{Amount: 1e15, Denom: "ugnot"}}}}
	appState.Txs = []TxWithMetadata{{Tx: std.Tx{
		Msgs: []std.Msg{vm.NewMsgAddPackage(addr, "gno.land/r/demo", []*std.MemFile{
			{Name: "demo.gno", Body: "package demo\nfunc Hello(cur realm) string { return `hello` }"},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/demo")},
		})},
		Fee:        std.Fee{GasWanted: 1e6, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}},
		Signatures: []std.Signature{{}},
	}}}
	resp := bapp.InitChain(abci.RequestInitChain{
		Time:            time.Now(),
		ChainID:         chainID,
		ConsensusParams: &abci.ConsensusParams{Block: defaultBlockParams()},
		Validators:      []abci.ValidatorUpdate{},
		AppState:        appState,
	})
	require.True(t, resp.IsOK(), "InitChain: %v", resp)
	bapp.Commit()

	// Advance past height 1 so getLastBlockHeader() reports a committed height
	// and Simulate takes the snapshot branch (not the fallback).
	for h := int64(1); h <= 2; h++ {
		bapp.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: h, Time: time.Now()}})
		bapp.EndBlock(abci.RequestEndBlock{Height: h})
		bapp.Commit()
	}

	// A signed call tx: simulation skips crypto verification but still requires
	// a PubKey to be present, so an unsigned tx would fail the ante past height 1.
	tx := createAndSignTx(t, []std.Msg{vm.NewMsgCall(addr, nil, "gno.land/r/demo", "Hello", nil)}, chainID, key)
	txBz := amino.MustMarshal(tx)

	qres := bapp.Query(abci.RequestQuery{Path: ".app/profiletx", Data: txBz})
	require.True(t, qres.IsOK(), "profiletx via snapshot path failed: %v (log=%s)", qres.Error, qres.Log)
	require.NotEmpty(t, qres.Value, "snapshot path must still produce a profile")
	require.Equal(t, "ok", qres.Log)

	gz, err := gzip.NewReader(bytes.NewReader(qres.Value))
	require.NoError(t, err)
	raw, err := io.ReadAll(gz)
	require.NoError(t, err)
	require.Contains(t, string(raw), "gno.land/r/demo.Hello")
}

// Off by default: without EnableGasProfiler the .app/profiletx query is rejected.
func TestApp_gasProfileQuery_disabledByDefault(t *testing.T) {
	t.Parallel()

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)

	qres := bapp.Query(abci.RequestQuery{Path: ".app/profiletx", Data: []byte("x")})
	require.False(t, qres.IsOK(), "profiletx must be disabled unless EnableGasProfiler is set")
}
