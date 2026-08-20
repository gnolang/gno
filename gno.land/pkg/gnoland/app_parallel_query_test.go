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

	queryAccountNumber := func(addr crypto.Address) (uint64, error) {
		res, err := query.QuerySync(abci.RequestQuery{
			Path: "auth/accounts/" + crypto.AddressToBech32(addr),
		})
		if err != nil {
			return 0, err
		}
		if res.IsErr() || len(res.Data) == 0 || string(res.Data) == "null" {
			return 0, fmt.Errorf("account query failed: %w (log: %q)", res.Error, res.Log)
		}
		var acct GnoAccount
		if err := amino.UnmarshalJSON(res.Data, &acct); err != nil {
			return 0, err
		}
		return acct.GetAccountNumber(), nil
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
	for i := range simKeys {
		accNum, err := queryAccountNumber(simKeys[i].PubKey().Address())
		require.NoError(t, err, "prefetch sim account %d", i)
		simTxs[i] = signedSend(simKeys[i], accNum, 0)
	}
	blockTxs := make([][]byte, blocks)
	for i := range senders {
		accNum, err := queryAccountNumber(senders[i].PubKey().Address())
		require.NoError(t, err, "prefetch sender account %d", i)
		blockTxs[i] = signedSend(senders[i], accNum, 0)
	}

	// Commit one real block BEFORE any querier starts. BaseApp.Simulate only
	// takes the immutable-snapshot path once getLastBlockHeader() reports a
	// height >= 1; below that it falls back to the shared checkState context
	// (tm2/pkg/sdk/helpers.go:60-64), which is single-threaded by construction
	// and races on its gas meter and live stores. That fallback is out of scope
	// here — this test covers the production snapshot path.
	base := app.(*sdk.BaseApp)
	commitBlock := func(h int64, i int) error {
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
	startHeight := base.LastBlockHeight() + 1
	require.NoError(t, commitBlock(startHeight, 0))
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
		probe.inFlight.Store(0)
		probe.peak.Store(0)
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
			if err := commitBlock(startHeight+int64(i), i); err != nil {
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
	// height. Each querier's own readings must be bit-identical across rounds —
	// concurrency must not perturb a pure read. (Readings differ BETWEEN
	// queriers by design: each signs from a different account, so the address
	// and account encodings it reads differ in length and depth, and so does
	// the per-byte read gas. Only the per-querier series is a fixed quantity.)
	gas, peak := runQueriers(rounds, nil)
	require.Greater(t, peak, int64(1), "phase 2: queries never overlapped")
	for g := range queriers {
		for r, got := range gas[g] {
			require.Equal(t, gas[g][0], got,
				"querier %d round %d: simulate gas is not deterministic under concurrency (%d != %d)",
				g, r, got, gas[g][0])
		}
	}
}
