package gnoland

// N-way concurrent-query stress test for the lock-free query connection.
//
// The queryMtx split (proxy.NewReadOnlyABCIClient) decoupled queries from
// consensus and mempool, but the query connection still held a per-call mutex,
// so queries remained serialised one-at-a-time among themselves. Removing that
// mutex turns "the query path is goroutine-safe" from a property nothing relied
// on into a load-bearing invariant: N simultaneous queries now execute inside
// one shared gnoStore/defaultStore, against the immutable DB snapshot, the
// SyncGoMap cacheNodes, the atomic last-block header, and per-tx forked
// allocators and caches.
//
// TestParallelQueries_NWaySimulate is the validation that invariant was missing.
// The pre-existing TestQueryRace_FastIndexParity pits ONE query against ONE
// committer, which the old queryMtx permitted; it can never observe two queries
// overlapping. This test asserts overlap directly: a peak-concurrency gauge must
// record at least two simulates in flight at the same instant, which fails on
// the mutexed code, and every simulate must return the same OK result under
// -race while the consensus connection commits blocks underneath.
//
// Run with -race; without it this only checks result stability and overlap.

import (
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bftCfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/config"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// concurrencyProbeApp counts how many calls are inside Application.Query at the
// same instant. It must wrap the application BENEATH the ABCI client, because
// that is where the client's per-call mutex is held: a gauge kept in the test's
// own goroutines would count callers merely BLOCKED on that mutex as in-flight
// and would report overlap even on fully serialised code.
type concurrencyProbeApp struct {
	abci.Application
	inFlight atomic.Int64
	peak     atomic.Int64
}

func (p *concurrencyProbeApp) Query(req abci.RequestQuery) abci.ResponseQuery {
	cur := p.inFlight.Add(1)
	for {
		old := p.peak.Load()
		if cur <= old || p.peak.CompareAndSwap(old, cur) {
			break
		}
	}
	defer p.inFlight.Add(-1)
	return p.Application.Query(req)
}

func TestParallelQueries_NWaySimulate(t *testing.T) {
	if testing.Short() {
		t.Skip("N-way concurrent query stress")
	}

	const (
		chainID = "dev"
		// queriers is the N in "N-way". Each gets its own account so no two
		// goroutines contend on a sequence number.
		queriers = 8
		// blocks committed on the consensus connection underneath the queries.
		blocks = 20
		// rounds of simultaneous simulate per querier.
		rounds = 25
	)

	// One account per querier (simulate only, sequence never advances), plus one
	// write-once sender per block for the consensus connection, plus a sink.
	simKeys := make([]ed25519.PrivKeyEd25519, queriers)
	senders := make([]ed25519.PrivKeyEd25519, blocks)
	balances := make([]Balance, 0, queriers+blocks+1)
	funded := std.Coins{std.NewCoin("ugnot", 100_000_000)}
	for i := range simKeys {
		simKeys[i] = ed25519.GenPrivKey()
		balances = append(balances, Balance{
			Address: simKeys[i].PubKey().Address(),
			Amount:  funded,
		})
	}
	for i := range senders {
		senders[i] = ed25519.GenPrivKey()
		balances = append(balances, Balance{
			Address: senders[i].PubKey().Address(),
			Amount:  funded,
		})
	}
	sink := ed25519.GenPrivKey().PubKey().Address()
	balances = append(balances, Balance{
		Address: sink,
		Amount:  std.Coins{std.NewCoin("ugnot", 1)},
	})

	appDir := t.TempDir()
	db, err := dbm.NewDB("gnolang", dbm.PebbleDBBackend, filepath.Join(appDir, bftCfg.DefaultDBDir))
	require.NoError(t, err)
	appCfg := config.DefaultAppConfig()
	genesisCfg := NewTestGenesisAppConfig()
	app, err := NewAppWithOptions(&AppOptions{
		DB:          db,
		Logger:      log.NewNoopLogger(),
		EventSwitch: events.NewEventSwitch(),
		InitChainerConfig: InitChainerConfig{
			GenesisTxResultHandler: PanicOnFailingTxResultHandler,
			StdlibDir:              filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs"),
		},
		MinGasPrices:               appCfg.MinGasPrices,
		SkipGenesisSigVerification: genesisCfg.SkipSigVerification,
		PruneStrategy:              appCfg.PruneStrategy,
	})
	require.NoError(t, err)

	// Production connection topology: a mutating consensus client and the
	// read-only query client the RPC layer holds.
	probe := &concurrencyProbeApp{Application: app}
	creator := proxy.NewLocalClientCreator(probe)
	cons, err := creator.NewABCIClient()
	require.NoError(t, err)
	require.NoError(t, cons.Start())
	defer cons.Stop()
	query, err := creator.NewReadOnlyABCIClient()
	require.NoError(t, err)
	require.NoError(t, query.Start())
	defer query.Stop()

	genState := DefaultGenState()
	genState.Balances = append(genState.Balances, balances...)
	_, err = cons.InitChainSync(abci.RequestInitChain{
		ChainID: chainID,
		Time:    time.Now(),
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{MaxGas: 100_000_000},
		},
		AppState: genState,
	})
	require.NoError(t, err)
	_, err = cons.CommitSync()
	require.NoError(t, err)

	// Only ever called from the test goroutine, before the queriers start, so it
	// may fail the test directly rather than plumb an error back.
	queryAccountNumber := func(addr crypto.Address) uint64 {
		bech32 := crypto.AddressToBech32(addr)
		res, err := query.QuerySync(abci.RequestQuery{Path: "auth/accounts/" + bech32})
		require.NoError(t, err, "query account %s", bech32)
		require.Falsef(t, res.IsErr() || len(res.Data) == 0 || string(res.Data) == "null",
			"account %s query failed: %v (log: %q)", bech32, res.Error, res.Log)
		var acct GnoAccount
		require.NoError(t, amino.UnmarshalJSON(res.Data, &acct), "decode account %s", bech32)
		return acct.GetAccountNumber()
	}

	sendAmount := std.Coins{std.NewCoin("ugnot", 1_000_000)}
	fee := std.NewFee(2_000_000, std.NewCoin("ugnot", 10_000_000))
	signedSend := func(key ed25519.PrivKeyEd25519, accNum, seq uint64) []byte {
		addr := key.PubKey().Address()
		tx := std.Tx{
			Msgs: []std.Msg{bank.NewMsgSend(addr, sink, sendAmount)},
			Fee:  fee,
		}
		signBytes, err := tx.GetSignBytes(chainID, accNum, seq)
		require.NoError(t, err)
		sig, err := key.Sign(signBytes)
		require.NoError(t, err)
		tx.Signatures = []std.Signature{{PubKey: key.PubKey(), Signature: sig}}
		return amino.MustMarshal(tx)
	}

	// Pre-sign every payload before any goroutine starts: the measured window
	// then contains only the query path, and signedSend's require calls stay on
	// the test goroutine (commitBlock runs on the block-loop goroutine, where
	// require -> t.FailNow would be illegal).
	simTxs := make([][]byte, queriers)
	for i, key := range simKeys {
		simTxs[i] = signedSend(key, queryAccountNumber(key.PubKey().Address()), 0)
	}
	blockTxs := make([][]byte, blocks)
	for i, key := range senders {
		blockTxs[i] = signedSend(key, queryAccountNumber(key.PubKey().Address()), 0)
	}

	// Commit one real block BEFORE any querier starts. BaseApp.Simulate only
	// takes the immutable-snapshot path once getLastBlockHeader() reports a
	// height >= 1; below that it falls back to the shared checkState context
	// (tm2/pkg/sdk/helpers.go:60-64), which is single-threaded by construction
	// and races on its gas meter and live stores. That fallback is out of scope
	// here — this test covers the production snapshot path.
	base := app.(*sdk.BaseApp)
	startHeight := base.LastBlockHeight() + 1
	// commitBlock commits block i of the run (its height follows from i) with the
	// i-th pre-signed tx. It returns errors instead of using require because the
	// block loop runs off the test goroutine in phase 1.
	commitBlock := func(i int) error {
		h := startHeight + int64(i)
		if _, err := cons.BeginBlockSync(abci.RequestBeginBlock{
			Header: &bft.Header{ChainID: chainID, Height: h, Time: time.Now()},
		}); err != nil {
			return err
		}
		resTx, err := cons.DeliverTxSync(abci.RequestDeliverTx{Tx: blockTxs[i]})
		if err != nil {
			return err
		}
		if !resTx.IsOK() {
			return fmt.Errorf("block %d: send failed: %w (log: %q)", h, resTx.Error, resTx.Log)
		}
		if _, err := cons.EndBlockSync(abci.RequestEndBlock{}); err != nil {
			return err
		}
		_, err = cons.CommitSync()
		return err
	}
	require.NoError(t, commitBlock(0))
	require.GreaterOrEqual(t, base.LastBlockHeight(), int64(1),
		"need a committed block so Simulate takes the snapshot path")

	// runQueriers fires `queriers` goroutines that each simulate their own
	// pre-signed tx `n` times, all released from one barrier so the calls
	// genuinely overlap. It returns the per-querier gas readings and the peak
	// number of simulates observed inside Application.Query simultaneously.
	// setup runs under the same barrier, for driving commits concurrently.
	runQueriers := func(n int, setup func()) (gas [][]int64, peak int64) {
		var (
			start sync.WaitGroup
			wg    sync.WaitGroup
		)
		probe.peak.Store(0) // inFlight is self-balancing; only the high-water mark carries over

		start.Add(1)
		gas = make([][]int64, queriers)
		errs := make([]error, queriers)

		for g := range queriers {
			wg.Go(func() {
				start.Wait()
				readings := make([]int64, 0, n)
				for range n {
					res, err := query.QuerySync(abci.RequestQuery{
						Path: ".app/simulate",
						Data: simTxs[g],
					})
					if err != nil {
						errs[g] = err
						return
					}
					if res.IsErr() {
						errs[g] = fmt.Errorf("simulate query: %w (log: %q)", res.Error, res.Log)
						return
					}
					var result sdk.Result
					if err := amino.Unmarshal(res.Value, &result); err != nil {
						errs[g] = err
						return
					}
					if result.IsErr() {
						errs[g] = fmt.Errorf("simulate result: %w (log: %q)", result.Error, result.Log)
						return
					}
					readings = append(readings, result.GasUsed)
				}
				gas[g] = readings
			})
		}
		if setup != nil {
			wg.Go(func() {
				start.Wait()
				setup()
			})
		}

		start.Done()
		wg.Wait()

		for g := range queriers {
			require.NoError(t, errs[g], "querier %d", g)
			require.Len(t, gas[g], n, "querier %d: missing rounds", g)
		}
		return gas, probe.peak.Load()
	}

	// Phase 1: queries overlap each other AND race a live committer. The -race
	// detector is the assertion here; gas is expected to move between rounds
	// because each simulate snapshots whatever height has been committed by
	// then, so it is only checked for success, not stability.
	var blockErr error
	_, peak := runQueriers(rounds, func() {
		for i := 1; i < blocks; i++ {
			if err := commitBlock(i); err != nil {
				blockErr = err
				return
			}
		}
	})
	require.NoError(t, blockErr, "consensus block loop failed")

	// This is what distinguishes the test from TestQueryRace_FastIndexParity.
	// With a per-call mutex on the query connection the peak is 1: no two
	// simulates can ever be inside Application.Query at once, and the -race run
	// above proves nothing about parallel queries.
	require.Greater(t, peak, int64(1),
		"queries never overlapped (peak in-flight = %d): the query connection is still serialising, "+
			"so this test is not exercising parallel queries", peak)

	// Phase 2: no committer, so every simulate snapshots the same committed
	// height. (Readings differ BETWEEN queriers by design: each signs from a
	// different account, so the address and account encodings it reads differ in
	// length and depth, and so does the per-byte read gas. Only the per-querier
	// series is a fixed quantity.)
	gas, peak := runQueriers(rounds, nil)
	require.Greater(t, peak, int64(1), "phase 2: queries never overlapped")

	// Compare against a SERIAL baseline, not against round 0 of the concurrent
	// run. Round 0 is itself a parallel round, so a systematic shift caused by
	// concurrency would move every round together and a self-comparison would
	// pass. Only a reading taken with nothing else in flight can distinguish
	// "stable under concurrency" from "equal to serial execution".
	//
	// The baseline is taken AFTER the concurrent phase, deliberately. Anything
	// filled on first touch — a lazily memoised type field, a cache — would be
	// warmed by a serial pass placed first, which would hide from -race exactly
	// the kind of first-touch race this test exists to catch.
	serial := make([]int64, queriers)
	for g := range queriers {
		res, err := query.QuerySync(abci.RequestQuery{
			Path: ".app/simulate",
			Data: simTxs[g],
		})
		require.NoError(t, err, "serial baseline for querier %d", g)
		require.Falsef(t, res.IsErr(), "serial baseline for querier %d: %v (log %q)", g, res.Error, res.Log)
		var result sdk.Result
		require.NoError(t, amino.Unmarshal(res.Value, &result))
		require.Falsef(t, result.IsErr(), "serial baseline result for querier %d: %v", g, result.Error)
		serial[g] = result.GasUsed
	}

	for g := range queriers {
		for r, got := range gas[g] {
			require.Equal(t, serial[g], got,
				"querier %d round %d: simulate under concurrency charged %d gas, serial execution charges %d",
				g, r, got, serial[g])
		}
	}
}

// TestParallelQueries_PreFirstBlockSimulate covers the window between the
// genesis commit and the first BeginBlock.
//
// In that window the last block header still reads height 0 — Commit
// republishes the header it was handed, and after InitChain that is initHeader,
// which carries no Height — while the store already holds a committed version.
// BaseApp.Simulate used to serve it from app.checkState, whose gas meter is a
// single pointer shared by every caller: harmless while the query connection
// held a mutex, a data race the moment it stopped.
//
// The window is reachable rather than theoretical. The node starts its RPC
// listeners before the P2P switch and the consensus reactor, `gnoland start
// -x-early-start` widens it on purpose, and a node with CreateEmptyBlocks=false
// (gnodev, and the integration harness) idles in it until a transaction shows
// up. So this test commits genesis and stops there, which is exactly the state
// such a node serves queries from.
//
// Run with -race; that is the assertion.
func TestParallelQueries_PreFirstBlockSimulate(t *testing.T) {
	if testing.Short() {
		t.Skip("concurrent pre-first-block simulate")
	}

	const (
		chainID  = "dev"
		queriers = 8
		rounds   = 5
	)

	simKeys := make([]ed25519.PrivKeyEd25519, queriers)
	balances := make([]Balance, 0, queriers+1)
	funded := std.Coins{std.NewCoin("ugnot", 100_000_000)}
	for i := range simKeys {
		simKeys[i] = ed25519.GenPrivKey()
		balances = append(balances, Balance{Address: simKeys[i].PubKey().Address(), Amount: funded})
	}
	sink := ed25519.GenPrivKey().PubKey().Address()
	balances = append(balances, Balance{Address: sink, Amount: std.Coins{std.NewCoin("ugnot", 1)}})

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)

	creator := proxy.NewLocalClientCreator(bapp)
	cons, err := creator.NewABCIClient()
	require.NoError(t, err)
	require.NoError(t, cons.Start())
	defer cons.Stop()
	query, err := creator.NewReadOnlyABCIClient()
	require.NoError(t, err)
	require.NoError(t, query.Start())
	defer query.Stop()

	genState := DefaultGenState()
	genState.Balances = append(genState.Balances, balances...)
	_, err = cons.InitChainSync(abci.RequestInitChain{
		ChainID:         chainID,
		Time:            time.Now(),
		ConsensusParams: &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 100_000_000}},
		AppState:        genState,
	})
	require.NoError(t, err)
	_, err = cons.CommitSync()
	require.NoError(t, err)

	// Genesis is committed and no block has been begun: this is the window.
	// The header reads height 0 while the store is at version 1, which is
	// precisely the mismatch that used to select the checkState fallback.

	fee := std.NewFee(2_000_000, std.NewCoin("ugnot", 10_000_000))
	simTxs := make([][]byte, queriers)
	for i, key := range simKeys {
		addr := key.PubKey().Address()
		bech32 := crypto.AddressToBech32(addr)
		res, err := query.QuerySync(abci.RequestQuery{Path: "auth/accounts/" + bech32})
		require.NoError(t, err)
		require.Falsef(t, res.IsErr(), "account %s: %v", bech32, res.Error)
		var acct GnoAccount
		require.NoError(t, amino.UnmarshalJSON(res.Data, &acct))

		tx := std.Tx{
			Msgs: []std.Msg{bank.NewMsgSend(addr, sink, std.Coins{std.NewCoin("ugnot", 1_000_000)})},
			Fee:  fee,
		}
		signBytes, err := tx.GetSignBytes(chainID, acct.GetAccountNumber(), 0)
		require.NoError(t, err)
		sig, err := key.Sign(signBytes)
		require.NoError(t, err)
		tx.Signatures = []std.Signature{{PubKey: key.PubKey(), Signature: sig}}
		simTxs[i] = amino.MustMarshal(tx)
	}

	var (
		start sync.WaitGroup
		wg    sync.WaitGroup
	)
	start.Add(1)
	errs := make([]error, queriers)
	gas := make([][]int64, queriers)

	for g := range queriers {
		gas[g] = make([]int64, 0, rounds)
		wg.Go(func() {
			start.Wait()
			for range rounds {
				res, err := query.QuerySync(abci.RequestQuery{
					Path: ".app/simulate",
					Data: simTxs[g],
				})
				if err != nil {
					errs[g] = err
					return
				}
				if res.IsErr() {
					errs[g] = fmt.Errorf("simulate query: %w (log %q)", res.Error, res.Log)
					return
				}
				var result sdk.Result
				if err := amino.Unmarshal(res.Value, &result); err != nil {
					errs[g] = err
					return
				}
				if result.IsErr() {
					errs[g] = fmt.Errorf("simulate result: %w (log %q)", result.Error, result.Log)
					return
				}
				gas[g] = append(gas[g], result.GasUsed)
			}
		})
	}
	start.Done()
	wg.Wait()

	for g := range queriers {
		require.NoError(t, errs[g], "querier %d", g)
		require.Len(t, gas[g], rounds, "querier %d: missing rounds", g)
	}

	// simulate reads querier g's tx once and returns the gas charged.
	simulate := func(g int, why string) int64 {
		t.Helper()
		res, err := query.QuerySync(abci.RequestQuery{
			Path: ".app/simulate",
			Data: simTxs[g],
		})
		require.NoErrorf(t, err, "%s for querier %d", why, g)
		require.Falsef(t, res.IsErr(), "%s for querier %d: %v (log %q)", why, g, res.Error, res.Log)
		var result sdk.Result
		require.NoError(t, amino.Unmarshal(res.Value, &result))
		require.Falsef(t, result.IsErr(), "%s result for querier %d: %v", why, g, result.Error)
		return result.GasUsed
	}

	// Compare against a serial baseline rather than against round 0 of the
	// concurrent run: round 0 is itself a parallel round, so a systematic shift
	// caused by concurrency would move every round together and a
	// self-comparison would pass. Taken after the concurrent phase, so that a
	// serial pass does not warm the first-touch caches this test exists to race.
	//
	// Per querier, not across queriers: each signs from its own account, and
	// account numbers and addresses do not all encode to the same number of
	// bytes, so two queriers legitimately differ by a few gas.
	serial := make([]int64, queriers)
	for g := range queriers {
		serial[g] = simulate(g, "serial baseline")
		for r, got := range gas[g] {
			require.Equal(t, serial[g], got,
				"querier %d round %d: window simulate charged %d gas under concurrency, "+
					"serial execution charges %d", g, r, got, serial[g])
		}
	}

	// Stability only says the window is self-consistent; a path reading the
	// wrong state consistently would pass it. What the hunk under test changed
	// is WHICH state the window reads — app.checkState, or the committed
	// snapshot the post-first-block path uses — so pin that by making the two
	// disagree and checking which one the answer follows.
	//
	// CheckTx flushes its ante writes into checkState (baseapp.runTx, the
	// RunTxModeCheck branch), so checking querier 0's own tx advances that
	// account's sequence and deducts its fee THERE and nowhere else. The
	// committed state is untouched, so a simulate served from the snapshot must
	// charge exactly what it charged before. Served from checkState it cannot:
	// the sequence it just bumped no longer matches the signature.
	//
	// The comparison is against the window's own baseline, deliberately. A
	// reading taken after a real block is not an oracle for this: the first
	// block's commit writes block metadata that later reads walk over, which
	// moves the gas by ~5% on its own and stays there for every block after.
	res, err := cons.CheckTxSync(abci.RequestCheckTx{Tx: simTxs[0]})
	require.NoError(t, err)
	require.Falsef(t, res.IsErr(), "CheckTx: %v (log %q)", res.Error, res.Log)

	require.Equal(t, serial[0], simulate(0, "simulate after CheckTx"),
		"querier 0's window simulate moved after a CheckTx that only touched "+
			"checkState, so the window is being served from checkState rather "+
			"than from the committed snapshot")
}
