package gnoland

// Full-app integration regression for https://github.com/gnolang/gno/issues/6011:
// concurrent query traffic must never corrupt the bptree fast index while the
// app commits blocks.
//
// This exercises the PRODUCTION concurrency topology at the outermost layer
// the bug is observable from: the real gnoland app (bptree main store with the
// fast index, pebble, production store wiring and options) wrapped in the real
// ABCI proxy client creator — block execution with signed bank sends driven
// through the CONSENSUS connection while goroutines hammer account and .store
// queries through the READ-ONLY QUERY connection, which has its own mutex and
// genuinely races the commit drain (tm2/pkg/bft/proxy). The txtar integration
// framework cannot express this (it is sequential); this is the flow that
// poisoned topaz-1.
//
// The racing interleaving itself is made deterministic the same way the
// tm2-level regression does it (the stochastic window — a version scan and a
// stamp read straddling a commit drain — is microseconds wide at this layer,
// far too narrow to hit reliably in CI time): the app's DB is wrapped so that
// the first query-path read of the fast-index stamp pauses until the block
// loop lands exactly one more commit. On the UNFIXED code the query connection
// reads the stamp during its per-query index maintenance, observes it ahead of
// the version it scanned, rebuilds the whole index from its outdated root, and
// permanently poisons the just-committed sender's entry (write-once senders:
// account i sends only at block i, so its reverted genesis-era entry is never
// healed) — caught by the end-of-run parity audit. On the fixed code the query
// path performs no index maintenance at all: the hook must never fire, and the
// audit must be clean.
//
// The test intentionally uses no post-fix API, so it can be run against the
// unfixed code to demonstrate the failure.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math/rand"
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
	bp "github.com/gnolang/gno/tm2/pkg/bptree"
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

// stampHookDB delegates to the wrapped DB; once armed, the first Get of the
// fast-index stamp key runs onStampGet before returning the (post-hook)
// value. It stands in for the consensus commit drain landing between a
// query-path load's version scan and its stamp read — the gno#6011 window.
type stampHookDB struct {
	dbm.DB
	armed      atomic.Bool
	fired      atomic.Bool
	onStampGet func()
}

func (h *stampHookDB) Get(key []byte) ([]byte, error) {
	if h.armed.Load() && bytes.HasSuffix(key, []byte("fastidx")) &&
		h.fired.CompareAndSwap(false, true) {
		h.onStampGet()
	}
	return h.DB.Get(key)
}

func TestQueryRace_FastIndexParity(t *testing.T) {
	if testing.Short() {
		t.Skip("full-app query/commit race")
	}

	const (
		chainID = "dev"
		blocks  = 40
	)
	appDir := t.TempDir()

	// One write-once sender per block, plus a sink receiver touched every block.
	senders := make([]ed25519.PrivKeyEd25519, blocks)
	balances := make([]Balance, 0, blocks+1)
	for i := range senders {
		senders[i] = ed25519.GenPrivKey()
		balances = append(balances, Balance{
			Address: senders[i].PubKey().Address(),
			Amount:  std.Coins{std.NewCoin("ugnot", 100_000_000)},
		})
	}
	sink := ed25519.GenPrivKey().PubKey().Address()
	balances = append(balances, Balance{
		Address: sink,
		Amount:  std.Coins{std.NewCoin("ugnot", 1)},
	})

	// The app over a stamp-hooked pebble DB; all other options exactly as
	// gnoland.NewApp assembles them.
	rawDB, err := dbm.NewDB("gnolang", dbm.PebbleDBBackend, filepath.Join(appDir, bftCfg.DefaultDBDir))
	require.NoError(t, err)
	hooked := &stampHookDB{DB: rawDB}
	appCfg := config.DefaultAppConfig()
	genesisCfg := NewTestGenesisAppConfig()
	app, err := NewAppWithOptions(&AppOptions{
		DB:          hooked,
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

	// The real production connection topology: consensus and query connections
	// with independent mutexes.
	creator := proxy.NewLocalClientCreator(app)
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

	// queryAccount goes through the QUERY connection — the surface gnokey and
	// gnoweb use.
	queryAccount := func(addr crypto.Address) (std.Account, error) {
		res, err := query.QuerySync(abci.RequestQuery{
			Path: "auth/accounts/" + crypto.AddressToBech32(addr),
		})
		if err != nil {
			return nil, err
		}
		if res.IsErr() || len(res.Data) == 0 || string(res.Data) == "null" {
			return nil, fmt.Errorf("account query failed: %w (log: %q)", res.Error, res.Log)
		}
		var acct GnoAccount
		if err := amino.UnmarshalJSON(res.Data, &acct); err != nil {
			return nil, err
		}
		return &acct, nil
	}

	// Prefetch account numbers BEFORE arming the hook: the block loop must
	// never touch the query mutex once the hook can pause a query (the hooked
	// query holds the query mutex while waiting for the loop to commit).
	accNums := make([]uint64, blocks)
	for i := range senders {
		acct, err := queryAccount(senders[i].PubKey().Address())
		require.NoError(t, err, "prefetch account %d", i)
		accNums[i] = acct.GetAccountNumber()
		require.EqualValues(t, 0, acct.GetSequence())
	}

	// Hook orchestration: the first query-path stamp read (a) freezes the
	// hammer so no later query can trigger a healing rebuild on the unfixed
	// code, and (b) waits for the block loop to land one more commit — placing
	// the stamp ahead of the version the query already scanned.
	var (
		stop        atomic.Bool
		wantCommit  atomic.Bool
		committed   = make(chan struct{}, 1)
		queryLoads  atomic.Int64
		hammerGroup sync.WaitGroup
	)
	hooked.onStampGet = func() {
		stop.Store(true)
		wantCommit.Store(true)
		select {
		case <-committed:
		case <-time.After(30 * time.Second):
			t.Error("hook: no commit arrived within 30s")
		}
	}
	hooked.armed.Store(true)

	for g := range 4 {
		hammerGroup.Go(func() {
			rng := rand.New(rand.NewSource(int64(1000 + g)))
			for !stop.Load() {
				addr := senders[rng.Intn(len(senders))].PubKey().Address()
				if g%2 == 0 {
					if _, err := queryAccount(addr); err == nil {
						queryLoads.Add(1)
					}
					continue
				}
				res, err := query.QuerySync(abci.RequestQuery{
					Path: ".store/main/key",
					Data: append([]byte("/a/"), addr.Bytes()...),
				})
				if err == nil && !res.IsErr() {
					queryLoads.Add(1)
				}
			}
		})
	}

	// Blocks: sender i sends once at block i (sequence 0, never touched again).
	sendAmount := std.Coins{std.NewCoin("ugnot", 1_000_000)}
	fee := std.NewFee(2_000_000, std.NewCoin("ugnot", 10_000_000))
	base := app.(*sdk.BaseApp)
	startHeight := base.LastBlockHeight() + 1
	for i := range blocks {
		h := startHeight + int64(i)
		sender := senders[i]
		addr := sender.PubKey().Address()

		tx := std.Tx{
			Msgs: []std.Msg{bank.NewMsgSend(addr, sink, sendAmount)},
			Fee:  fee,
		}
		signBytes, err := tx.GetSignBytes(chainID, accNums[i], 0)
		require.NoError(t, err)
		sig, err := sender.Sign(signBytes)
		require.NoError(t, err)
		tx.Signatures = []std.Signature{{PubKey: sender.PubKey(), Signature: sig}}

		_, err = cons.BeginBlockSync(abci.RequestBeginBlock{
			Header: &bft.Header{ChainID: chainID, Height: h, Time: time.Now()},
		})
		require.NoError(t, err)
		resTx, err := cons.DeliverTxSync(abci.RequestDeliverTx{Tx: amino.MustMarshal(tx)})
		require.NoError(t, err)
		require.True(t, resTx.IsOK(), "block %d: send failed: %v %q", h, resTx.Error, resTx.Log)
		_, err = cons.EndBlockSync(abci.RequestEndBlock{})
		require.NoError(t, err)
		_, err = cons.CommitSync()
		require.NoError(t, err)

		// Serve a pending hook request: this commit is the one that lands
		// between the hooked query's version scan and its stamp read.
		if wantCommit.CompareAndSwap(true, false) {
			committed <- struct{}{}
		}
	}
	stop.Store(true)
	hammerGroup.Wait()

	// The hook fires only on a stamp read through the LIVE DB handle. On the
	// fixed code the query path never touches the live handle — it loads
	// read-only from the frozen snapshot (the getImmutable stamp gate reads the
	// snapshot's stamp, which bypasses this hook), so the hook must not have
	// fired. (On the unfixed code the query's index maintenance reads the live
	// stamp, fires the hook, and the audit below reports the poisoning.)
	require.False(t, hooked.fired.Load(),
		"query path read the live fast-index stamp: query-side index maintenance is back")
	require.Positive(t, queryLoads.Load(), "hammer performed no successful queries")

	// Shut the app down and audit the persisted fast index from the raw DB,
	// exactly as issue #6011's diagnostic did: every 'F' entry must match the
	// authoritative tree walk.
	require.NoError(t, base.Close())
	db, err := dbm.NewDB("gnolang", dbm.PebbleDBBackend, filepath.Join(appDir, bftCfg.DefaultDBDir))
	require.NoError(t, err)
	defer db.Close()

	pdb := dbm.NewPrefixDB(db, []byte("s/_/"))
	plain := bp.NewMutableTreeWithDB(pdb, 10000, bp.NewNopLogger()) // fast index OFF: authoritative walks
	_, err = plain.Load()
	require.NoError(t, err)

	itr, err := pdb.Iterator([]byte{bp.PrefixFast}, []byte{bp.PrefixFast + 1})
	require.NoError(t, err)
	defer itr.Close()
	stale := 0
	for ; itr.Valid(); itr.Next() {
		key := append([]byte(nil), itr.Key()[1:]...)
		payload, cerr := queryRaceVerifyChecksum(itr.Value())
		if cerr != nil || len(payload) < 8 {
			t.Fatalf("corrupt F entry %q: %v", key, cerr)
		}
		want, err := plain.Get(key)
		require.NoError(t, err)
		if want == nil {
			stale++
			t.Errorf("STALE-PRESENT F entry for deleted key %q", key)
			continue
		}
		if !bytes.Equal(payload[8:], want) {
			stale++
			if stale <= 5 {
				t.Errorf("STALE F entry key=%q (issue #6011 shape)", key)
			}
		}
	}
	require.NoError(t, itr.Error())
	if stale > 0 {
		t.Fatalf("issue #6011 regressed at the full-app layer: %d stale fast-index entries persisted", stale)
	}
}

var queryRaceCRC = crc32.MakeTable(crc32.Castagnoli)

// queryRaceVerifyChecksum re-implements bptree's record framing check
// (payload ‖ big-endian crc32c) independently of the code under test.
// Sibling of reproVerifyChecksum in tm2/pkg/store/rootmulti's fastindex
// tests (test-package helpers are not importable); update both together
// if the framing changes.
func queryRaceVerifyChecksum(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("record too short (%d bytes)", len(data))
	}
	payload := data[:len(data)-4]
	want := binary.BigEndian.Uint32(data[len(data)-4:])
	if got := crc32.Checksum(payload, queryRaceCRC); got != want {
		return nil, fmt.Errorf("crc mismatch")
	}
	return payload, nil
}
