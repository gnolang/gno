package sdk

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
)

const deferTestChain = "test-chain"

// newDeferTestApp builds a BaseApp over what gno.land mounts, a bptree main
// store and a plain base store. Its msgCounter handler sets k<n> = v<n> in
// both stores and answers custom queries for a main-store key; its
// msgCounter2 handler iterates the base store, as does its ante handler in
// CheckTx while checkTxIterates (if not nil) is set.
func newDeferTestApp(t *testing.T, db dbm.DB, checkTxIterates *atomic.Bool, options ...func(*BaseApp)) *BaseApp {
	t.Helper()

	app := NewBaseApp(t.Name(), defaultLogger(), db, baseKey, mainKey, options...)
	app.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, nil)
	app.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, nil)
	app.SetAnteHandler(func(ctx Context, tx Tx, simulate bool) (Context, Result, bool) {
		if ctx.Mode() == RunTxModeCheck && checkTxIterates != nil && checkTxIterates.Load() {
			it := ctx.Store(baseKey).Iterator(nil, nil, nil)
			it.Close()
		}
		return ctx, Result{}, false
	})
	app.Router().AddRoute(routeMsgCounter, testHandler{
		process: func(ctx Context, msg Msg) Result {
			n := msg.(msgCounter).Counter
			ctx.Store(mainKey).Set(nil, fmt.Appendf(nil, "k%d", n), fmt.Appendf(nil, "v%d", n))
			ctx.Store(baseKey).Set(nil, fmt.Appendf(nil, "k%d", n), fmt.Appendf(nil, "v%d", n))
			return Result{}
		},
		query: func(ctx Context, req abci.RequestQuery) abci.ResponseQuery {
			return abci.ResponseQuery{Value: ctx.Store(mainKey).Get(nil, req.Data)}
		},
	})
	app.Router().AddRoute(routeMsgCounter2, newTestHandler(func(ctx Context, msg Msg) Result {
		it := ctx.Store(baseKey).Iterator(nil, nil, nil)
		it.Close()
		return Result{}
	}))
	require.NoError(t, app.LoadLatestVersion())
	return app
}

func deferTestTx(t *testing.T, msgs ...Msg) []byte {
	t.Helper()

	bz, err := amino.Marshal(std.Tx{Msgs: msgs})
	require.NoError(t, err)
	return bz
}

// deferTestBlock runs block n with one transaction that sets k<n>, and commits.
func deferTestBlock(t *testing.T, app *BaseApp, n int64) {
	t.Helper()

	app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: deferTestChain, Height: n}})
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: deferTestTx(t, msgCounter{Counter: n})})
	require.Nil(t, res.Error)
	app.EndBlock(abci.RequestEndBlock{Height: n})
	app.Commit()
}

func queryMain(app *BaseApp, key string) abci.ResponseQuery {
	return app.Query(abci.RequestQuery{Path: ".store/main/key", Data: []byte(key)})
}

func queryCustom(app *BaseApp, key string) abci.ResponseQuery {
	return app.Query(abci.RequestQuery{Path: routeMsgCounter + "/key", Data: []byte(key)})
}

// panicError returns what f panics with, if an error; nil if f returns.
func panicError(f func()) (err error) {
	defer func() { err, _ = recover().(error) }()
	f()
	return nil
}

// TestDeferCommitsNeedsTheOption pins that BaseApp defers only when built
// with AllowDeferredCommits, and stops deferring at DeferCommits(false).
func TestDeferCommitsNeedsTheOption(t *testing.T) {
	t.Parallel()

	without := newDeferTestApp(t, memdb.NewMemDB(), nil)
	require.False(t, without.DeferCommits(true), "deferred without AllowDeferredCommits")

	app := newDeferTestApp(t, memdb.NewMemDB(), nil, AllowDeferredCommits())
	app.InitChain(abci.RequestInitChain{ChainID: deferTestChain})
	require.True(t, app.DeferCommits(true))
	deferTestBlock(t, app, 1)
	require.Panics(t, func() { app.DeferCommits(false) }, "deferral turned off with a commit held")
	require.NoError(t, app.WriteCommits())
	require.False(t, app.DeferCommits(false))
	deferTestBlock(t, app, 2)
	require.Equal(t, int64(2), app.Info(abci.RequestInfo{}).LastBlockHeight, "not deferring: every commit is written")
}

// TestDeferredCommitsAnswerFromDisk pins that while commits are held, Info,
// store and custom queries and Simulate answer from the last version on
// disk, and that WriteCommits moves them to the held version; a node
// restarted before that, as after a crash, is at the written version.
func TestDeferredCommitsAnswerFromDisk(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	app := newDeferTestApp(t, db, nil, AllowDeferredCommits())
	app.InitChain(abci.RequestInitChain{ChainID: deferTestChain})
	deferTestBlock(t, app, 1)

	require.True(t, app.DeferCommits(true))
	deferTestBlock(t, app, 2)
	deferTestBlock(t, app, 3)

	require.Equal(t, int64(3), app.LastBlockHeight())
	require.Equal(t, int64(1), app.Info(abci.RequestInfo{}).LastBlockHeight)
	res := queryMain(app, "k1")
	require.Nil(t, res.Error)
	require.Equal(t, []byte("v1"), res.Value)
	require.Equal(t, int64(1), res.Height)
	require.Empty(t, queryMain(app, "k3").Value, "a query answered from a held commit")
	require.Equal(t, []byte("v1"), queryCustom(app, "k1").Value)
	require.Empty(t, queryCustom(app, "k3").Value, "a custom query answered from a held commit")
	above := app.Query(abci.RequestQuery{Path: ".store/main/key", Data: []byte("k3"), Height: 3})
	require.NotNil(t, above.Error, "a query above the version on disk")
	sim := app.Simulate(deferTestTx(t, msgCounter{Counter: 9}))
	require.Nil(t, sim.Error)
	require.Equal(t, int64(1), app.getLastBlockHeader().GetHeight(), "a held header published")

	restarted := newDeferTestApp(t, db, nil, AllowDeferredCommits())
	require.Equal(t, int64(1), restarted.LastBlockHeight())
	require.Equal(t, int64(1), restarted.Info(abci.RequestInfo{}).LastBlockHeight)

	require.NoError(t, app.WriteCommits())
	require.Equal(t, int64(3), app.Info(abci.RequestInfo{}).LastBlockHeight)
	require.Equal(t, []byte("v3"), queryMain(app, "k3").Value)
	require.Equal(t, int64(3), app.getLastBlockHeader().GetHeight())

	restarted = newDeferTestApp(t, db, nil, AllowDeferredCommits())
	require.Equal(t, app.LastCommitID(), restarted.LastCommitID())
}

// TestDeferredCommitsFreshNodeRefusesQueries pins that a node with nothing on
// disk answers no query and no Simulate while its commits are held: custom
// queries would read an empty version, Simulate the live state.
func TestDeferredCommitsFreshNodeRefusesQueries(t *testing.T) {
	t.Parallel()

	app := newDeferTestApp(t, memdb.NewMemDB(), nil, AllowDeferredCommits())
	app.InitChain(abci.RequestInitChain{ChainID: deferTestChain})
	require.True(t, app.DeferCommits(true))
	deferTestBlock(t, app, 1)

	require.NotNil(t, queryMain(app, "k1").Error)
	require.NotNil(t, queryCustom(app, "k1").Error)
	require.NotNil(t, app.Simulate(deferTestTx(t, msgCounter{Counter: 9})).Error)

	require.NoError(t, app.WriteCommits())
	require.Equal(t, []byte("v1"), queryMain(app, "k1").Value)
	require.Equal(t, []byte("v1"), queryCustom(app, "k1").Value)
	require.Nil(t, app.Simulate(deferTestTx(t, msgCounter{Counter: 9})).Error)
}

// TestDeferredCommitsSimulateAfterGenesis pins that Simulate between the
// genesis commit and the first flush point answers from the genesis version
// on disk: InitChain's header has no height, so Simulate falls back to the
// version, which must be the one on disk, not the one held.
func TestDeferredCommitsSimulateAfterGenesis(t *testing.T) {
	t.Parallel()

	app := newDeferTestApp(t, memdb.NewMemDB(), nil, AllowDeferredCommits())
	app.InitChain(abci.RequestInitChain{ChainID: deferTestChain})
	app.Commit()
	require.True(t, app.DeferCommits(true))
	deferTestBlock(t, app, 1)
	require.Zero(t, app.getLastBlockHeader().GetHeight())
	require.Nil(t, app.Simulate(deferTestTx(t, msgCounter{Counter: 9})).Error)
}

// TestNeedsFlushPoint pins its two reasons, the halt height reached by the
// version in memory (the one on disk stays below it) and a required flush
// point (until written), and that HeldBytes reports what is held.
func TestNeedsFlushPoint(t *testing.T) {
	t.Parallel()

	app := newDeferTestApp(t, memdb.NewMemDB(), nil, AllowDeferredCommits())
	app.InitChain(abci.RequestInitChain{ChainID: deferTestChain})
	app.SetHaltHeight(3)
	require.True(t, app.DeferCommits(true))

	deferTestBlock(t, app, 1)
	deferTestBlock(t, app, 2)
	require.False(t, app.NeedsFlushPoint())
	deferTestBlock(t, app, 3)
	require.True(t, app.NeedsFlushPoint(), "halt height reached in memory")
	require.NoError(t, app.WriteCommits())
	app.SetHaltHeight(0)
	require.False(t, app.NeedsFlushPoint())

	app.RequireFlushPoint()
	require.True(t, app.NeedsFlushPoint())
	require.NoError(t, app.WriteCommits())
	require.False(t, app.NeedsFlushPoint(), "a required flush point stays required after the write")

	require.Zero(t, app.HeldBytes())
	deferTestBlock(t, app, 4)
	require.Positive(t, app.HeldBytes())
	require.False(t, app.NeedsFlushPoint(), "held bytes are the node's budget, not a reason of the app's")
}

// TestDeferredCommitsRefusedIteratorEndsTheBlock pins that an iterator over
// the database inside a block, while commits are held, ends the block with a
// panic even though runTx recovers it, while in CheckTx it only fails the
// transaction, and while not deferring it works.
func TestDeferredCommitsRefusedIteratorEndsTheBlock(t *testing.T) {
	t.Parallel()

	var checkTxIterates atomic.Bool
	app := newDeferTestApp(t, memdb.NewMemDB(), &checkTxIterates, AllowDeferredCommits())
	app.InitChain(abci.RequestInitChain{ChainID: deferTestChain})
	iterating := deferTestTx(t, msgCounter2{Counter: 1})

	app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: deferTestChain, Height: 1}})
	require.Nil(t, app.DeliverTx(abci.RequestDeliverTx{Tx: iterating}).Error, "iterating while not deferring")
	app.EndBlock(abci.RequestEndBlock{Height: 1})
	app.Commit()

	require.True(t, app.DeferCommits(true))
	app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: deferTestChain, Height: 2}})
	require.ErrorIs(t, panicError(func() { app.DeliverTx(abci.RequestDeliverTx{Tx: iterating}) }), dbm.ErrIteratorOverHeldWrites)

	// CheckTx runs only the ante handler; one that iterates fails its
	// transaction, and leaves no record for the block's next call.
	checkTxIterates.Store(true)
	require.NotNil(t, app.CheckTx(abci.RequestCheckTx{Tx: deferTestTx(t, msgCounter{Counter: 2})}).Error)
	require.NotPanics(t, func() {
		require.Nil(t, app.DeliverTx(abci.RequestDeliverTx{Tx: deferTestTx(t, msgCounter{Counter: 2})}).Error)
	})
}

// TestDeferredCommitsRefusedIteratorInBlockers pins that BeginBlock and
// EndBlock end the block when a blocker opened a refused iterator while
// commits are held, even if the blocker recovers the panic, but not for a
// refusal CheckTx recorded before them, and that the same blockers run
// unharmed while not deferring.
func TestDeferredCommitsRefusedIteratorInBlockers(t *testing.T) {
	t.Parallel()

	var iterate, checkTxIterates atomic.Bool
	swallow := func(ctx Context) {
		if !iterate.Load() {
			return
		}
		defer func() { recover() }() // as the VM's recovers do
		it := ctx.Store(baseKey).Iterator(nil, nil, nil)
		it.Close()
	}
	blockers := func(app *BaseApp) {
		app.SetBeginBlocker(func(ctx Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
			swallow(ctx)
			return abci.ResponseBeginBlock{}
		})
		app.SetEndBlocker(func(ctx Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
			swallow(ctx)
			return abci.ResponseEndBlock{}
		})
	}
	app := newDeferTestApp(t, memdb.NewMemDB(), &checkTxIterates, AllowDeferredCommits(), blockers)
	app.InitChain(abci.RequestInitChain{ChainID: deferTestChain})
	begin := func(h int64) func() {
		return func() {
			app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: deferTestChain, Height: h}})
		}
	}
	end := func(h int64) func() {
		return func() { app.EndBlock(abci.RequestEndBlock{Height: h}) }
	}

	iterate.Store(true)
	require.NotPanics(t, begin(1), "BeginBlock while not deferring")
	require.NotPanics(t, end(1), "EndBlock while not deferring")
	app.Commit()

	require.True(t, app.DeferCommits(true))
	require.ErrorIs(t, panicError(begin(2)), dbm.ErrIteratorOverHeldWrites)
	require.ErrorIs(t, panicError(end(2)), dbm.ErrIteratorOverHeldWrites)

	iterate.Store(false)
	checkTxIterates.Store(true)
	checkTx := func() {
		require.NotNil(t, app.CheckTx(abci.RequestCheckTx{Tx: deferTestTx(t, msgCounter{Counter: 3})}).Error)
	}
	checkTx()
	require.NotPanics(t, begin(2), "BeginBlock ended by CheckTx's refusal")
	checkTx()
	require.NotPanics(t, end(2), "EndBlock ended by CheckTx's refusal")
}

// TestDeferredCommitsCloseWritesNothing pins that closing the app drops its
// held commits, as a crash would: the app never writes on its own.
func TestDeferredCommitsCloseWritesNothing(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	app := newDeferTestApp(t, db, nil, AllowDeferredCommits())
	app.InitChain(abci.RequestInitChain{ChainID: deferTestChain})
	deferTestBlock(t, app, 1)
	require.True(t, app.DeferCommits(true))
	deferTestBlock(t, app, 2)
	require.NoError(t, app.Close())

	reopened := newDeferTestApp(t, db, nil, AllowDeferredCommits())
	require.Equal(t, int64(1), reopened.LastBlockHeight(), "Close wrote a held commit")
}
