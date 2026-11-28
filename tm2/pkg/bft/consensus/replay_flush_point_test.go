package consensus

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	bc "github.com/gnolang/gno/tm2/pkg/bft/blockchain"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
)

const (
	flushCutBlocks = 8 // two windows
	flushCutWindow = 4 // a flush point every four blocks: a block reads the validators a block three before it saved, in the same window
)

// TestHandshakeAfterFlushPointPowerCut syncs a chain with deferred writes, a
// flush point every flushCutWindow blocks (blockchain.DeferredWrites), cuts
// the power at every step of every flush point, and checks that a node
// restarted on what survives, and syncing on with deferred writes too, ends
// with the run's state: the same state, every key of both databases and the
// app's height and hash, and no block's events fired twice. The chain
// changes its validator set in the block a restart replays after a crash
// between a flush point's first two writes, changes its consensus params,
// has an empty block, and starts at height 1 or, as a hardfork does, at a
// later InitialHeight.
//
// A database keeps the writes its last flush covered and, of a batch written
// since, all of it or none; the app keeps every height it last wrote, or any
// lower one, as an app that loses its writes might.
func TestHandshakeAfterFlushPointPowerCut(t *testing.T) {
	t.Parallel()

	for _, initial := range []int64{1, 5} {
		t.Run(fmt.Sprintf("InitialHeight=%d", initial), func(t *testing.T) {
			t.Parallel()

			run := newFlushCutRun(t, initial)
			restarts := 0
			for step := run.start; step <= *run.clock; step++ {
				blockDurable, blockUnflushed := run.blockDisk.at(step)
				stateDurable, stateUnflushed := run.stateDisk.at(step)
				for _, blockKept := range flushCutVariants(blockUnflushed) {
					for _, stateKept := range flushCutVariants(stateUnflushed) {
						for appHeight := run.app.heightAt(step); appHeight >= 0; appHeight-- {
							if appHeight > 0 && appHeight < run.initial {
								continue // no such heights: the app is at 0
							}
							blockDB := rebuildGroups(slices.Concat(blockDurable, blockKept))
							stateDB := rebuildGroups(slices.Concat(stateDurable, stateKept))
							restarts++
							if err := run.restart(blockDB, stateDB, appHeight, step); err != nil {
								t.Fatalf("power cut at step %d, block store keeping %d of %d unflushed batches, "+
									"state keeping %d of %d, app at %d: %v",
									step, len(blockKept), len(blockUnflushed), len(stateKept), len(stateUnflushed), appHeight, err)
							}
						}
					}
				}
			}
			// Two flush points of five steps each, every one cut, each cut at
			// least once restarted.
			cuts := *run.clock - run.start + 1
			require.Equal(t, 2*5, cuts, "steps of block sync")
			require.GreaterOrEqual(t, restarts, cuts)
			require.Empty(t, run.earlyEvents, "events fired before their block's commit was on disk")
			t.Logf("%d power cuts, %d restarts", cuts, restarts)
		})
	}
}

// flushCutRun is a block-sync run with deferred writes whose block store and
// state disks log every write and batch (cutDisk), and whose app notes when
// each height it commits becomes durable, so a test can rebuild what a
// restart finds after a power cut at any step.
type flushCutRun struct {
	genDoc    *types.GenesisDoc
	initial   int64 // the chain's first height
	clock     *int  // the steps taken so far
	start     int   // the first step of block sync
	blockDisk *cutDisk
	stateDisk *cutDisk
	app       *flushCutApp

	blocks  []*types.Block // blocks[i] is at height initial+i
	parts   []*types.PartSet
	commits []*types.Commit
	final   sm.State

	eventSteps  map[int64]int // the step after which the run fired each height's EventNewBlock
	earlyEvents []int64       // heights whose events fired before the app's commit of them was durable
}

func newFlushCutRun(t *testing.T, initial int64) *flushCutRun {
	t.Helper()

	genDoc, privVals := randGenesisDoc(1, false, 10)
	genDoc.InitialHeight = initial
	clock := new(int)
	run := &flushCutRun{
		genDoc:     genDoc,
		initial:    initial,
		clock:      clock,
		blockDisk:  &cutDisk{DB: memdb.NewMemDB(), clock: clock},
		stateDisk:  &cutDisk{DB: memdb.NewMemDB(), clock: clock},
		app:        newFlushCutApp(clock, initial),
		eventSteps: map[int64]int{},
	}
	blockDB, stateDB := dbm.NewHoldingDB(run.blockDisk), dbm.NewHoldingDB(run.stateDisk)

	// Start as a node does: load the genesis state, then shake hands with the
	// app, which sends it InitChain.
	blockStore := store.NewBlockStore(blockDB)
	state, err := sm.LoadStateFromDBOrGenesisDoc(stateDB, genDoc)
	require.NoError(t, err)
	conns := appconn.NewAppConns(proxy.NewLocalClientCreator(run.app))
	require.NoError(t, conns.Start())
	defer conns.Stop()
	handshaker := NewHandshaker(stateDB, state, blockStore, genDoc)
	handshaker.SetEventSwitch(events.NewEventSwitch())
	require.NoError(t, handshaker.Handshake(conns))
	state = sm.LoadState(stateDB)
	run.start = *clock + 1

	// Sync the chain as block sync does with deferred writes: save each
	// block, apply it, and take a flush point every flushCutWindow blocks.
	evsw := events.NewEventSwitch()
	require.NoError(t, evsw.Start())
	defer evsw.Stop()
	evsw.AddListener("flush-cut", func(ev events.Event) {
		if nb, ok := ev.(types.EventNewBlock); ok {
			run.eventSteps[nb.Block.Height] = *clock
			if run.app.heightAt(*clock) < nb.Block.Height {
				run.earlyEvents = append(run.earlyEvents, nb.Block.Height)
			}
		}
	})
	blockExec := sm.NewBlockExecutor(stateDB, log.NewNoopLogger(), conns.Consensus(), mempool)
	blockExec.SetEventSwitch(evsw)
	deferred := bc.NewDeferredWrites(blockStore, blockExec, run.app)
	require.True(t, deferred.Start(state.LastBlockHeight))

	lastCommit := new(types.Commit)
	for h := initial; h < initial+flushCutBlocks; h++ {
		txs := []types.Tx{fmt.Appendf(nil, "a%d=%d", h, h), fmt.Appendf(nil, "b%d=%d", h, h)}
		if h-initial == 6 {
			txs = nil // an empty block, whose results encode to no bytes
		}
		block, parts := state.MakeBlock(h, txs, lastCommit, state.Validators.GetProposer().Address)
		blockID := types.BlockID{Hash: block.Hash(), PartsHeader: parts.Header()}
		lastCommit = flushCutCommit(t, genDoc, privVals[0], state.Validators, blockID, h)

		blockStore.SaveBlock(block, parts, lastCommit)
		state, err = blockExec.ApplyBlock(state, blockID, block)
		require.NoError(t, err)
		if (h-initial+1)%flushCutWindow == 0 {
			deferred.FlushPoint(h)
		}

		run.blocks = append(run.blocks, block)
		run.parts = append(run.parts, parts)
		run.commits = append(run.commits, lastCommit)
	}
	run.final = state
	require.Len(t, run.eventSteps, flushCutBlocks, "a block's events never fired")

	return run
}

// flushCutCommit returns height's commit: the precommit of the validator
// signing, and an absent one for every other validator in vals.
func flushCutCommit(t *testing.T, genDoc *types.GenesisDoc, privVal types.PrivValidator, vals *types.ValidatorSet,
	blockID types.BlockID, height int64,
) *types.Commit {
	t.Helper()

	sigs := make([]*types.CommitSig, vals.Size())
	for i, val := range vals.Validators {
		if !val.PubKey.Equals(privVal.PubKey()) {
			continue
		}
		vote := &types.Vote{
			ValidatorAddress: val.Address,
			ValidatorIndex:   i,
			Height:           height,
			Timestamp:        genDoc.GenesisTime.Add(time.Duration(height) * time.Second),
			Type:             types.PrecommitType,
			BlockID:          blockID,
		}
		require.NoError(t, privVal.SignVote(genDoc.ChainID, vote))
		sigs[i] = vote.CommitSig()
	}

	return types.NewCommit(blockID, sigs)
}

// restart starts a node on blockDB, stateDB and an app at appHeight, as found
// after a power cut at step, syncs it to the end of the run's chain as block
// sync does, deferring again, and reports where it differs from the run.
func (run *flushCutRun) restart(blockDB, stateDB dbm.DB, appHeight int64, step int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	app, err := run.appAt(appHeight)
	if err != nil {
		return err
	}
	conns := appconn.NewAppConns(proxy.NewLocalClientCreator(app))
	if err := conns.Start(); err != nil {
		return err
	}
	defer conns.Stop()

	fired := map[int64]int{}
	evsw := events.NewEventSwitch()
	if err := evsw.Start(); err != nil {
		return err
	}
	defer evsw.Stop()
	evsw.AddListener("flush-cut", func(ev events.Event) {
		if nb, ok := ev.(types.EventNewBlock); ok {
			fired[nb.Block.Height]++
		}
	})

	blockHeld, stateHeld := dbm.NewHoldingDB(blockDB), dbm.NewHoldingDB(stateDB)
	blockStore := store.NewBlockStore(blockHeld)
	state, err := sm.LoadStateFromDBOrGenesisDoc(stateHeld, run.genDoc)
	if err != nil {
		return err
	}
	handshaker := NewHandshaker(stateHeld, state, blockStore, run.genDoc)
	handshaker.SetEventSwitch(evsw)
	if err := handshaker.Handshake(conns); err != nil {
		return fmt.Errorf("handshake: %w", err)
	}

	// Sync on as block sync does, deferring again.
	state = sm.LoadState(stateHeld)
	blockExec := sm.NewBlockExecutor(stateHeld, log.NewNoopLogger(), conns.Consensus(), mempool)
	blockExec.SetEventSwitch(evsw)
	deferred := bc.NewDeferredWrites(blockStore, blockExec, app.(abci.DeferredCommitter))
	if !deferred.Start(state.LastBlockHeight) {
		return fmt.Errorf("the restarted app does not defer")
	}
	next := blockStore.Height() + 1
	if blockStore.Height() == 0 {
		next = run.initial
	}
	for h := next; h < run.initial+flushCutBlocks; h++ {
		i := h - run.initial
		block, parts := run.blocks[i], run.parts[i]
		blockStore.SaveBlock(block, parts, run.commits[i])
		state, err = blockExec.ApplyBlock(state, types.BlockID{Hash: block.Hash(), PartsHeader: parts.Header()}, block)
		if err != nil {
			return fmt.Errorf("apply block %d: %w", h, err)
		}
		if (h-run.initial+1)%flushCutWindow == 0 {
			deferred.FlushPoint(h)
		}
	}
	deferred.FlushPoint(state.LastBlockHeight)
	deferred.Stop()

	for h, n := range fired {
		if before, ok := run.eventSteps[h]; ok && before < step {
			n++
		}
		if n > 1 {
			return fmt.Errorf("height %d's events fired %d times", h, n)
		}
	}

	return run.compare(state, blockDB, stateDB, conns.Query())
}

// appAt returns an app that has committed the run's first height-initial+1
// blocks, or none for a height below initial.
func (run *flushCutRun) appAt(height int64) (abci.Application, error) {
	app := newFlushCutApp(new(int), run.initial)
	conns := appconn.NewAppConns(proxy.NewLocalClientCreator(app))
	if err := conns.Start(); err != nil {
		return nil, err
	}
	defer conns.Stop()
	for h := run.initial; h <= height; h++ {
		if _, err := sm.ExecCommitBlock(conns.Consensus(), run.blocks[h-run.initial], run.final, log.NewNoopLogger(), run.stateDisk); err != nil {
			return nil, err
		}
	}

	return app, nil
}

// compare reports the first thing in state, the databases or the app that
// differs from what the run ended with.
func (run *flushCutRun) compare(state sm.State, blockDB, stateDB dbm.DB, query appconn.Query) error {
	want := run.final
	if !bytes.Equal(state.Bytes(), want.Bytes()) {
		return fmt.Errorf("synced state differs: height %d, app hash %X; want height %d, app hash %X",
			state.LastBlockHeight, state.AppHash, want.LastBlockHeight, want.AppHash)
	}
	info, err := query.InfoSync(abci.RequestInfo{})
	if err != nil {
		return err
	}
	if info.LastBlockHeight != want.LastBlockHeight || !bytes.Equal(info.LastBlockAppHash, want.AppHash) {
		return fmt.Errorf("app at height %d with hash %X; want %d, %X",
			info.LastBlockHeight, info.LastBlockAppHash, want.LastBlockHeight, want.AppHash)
	}
	if err := sameKeys("block store", blockDB, run.blockDisk); err != nil {
		return err
	}

	return sameKeys("state", stateDB, run.stateDisk)
}

// sameKeys reports the first key whose value differs between got and want,
// but for the flush-point record, which only a deferring node writes.
func sameKeys(name string, got, want dbm.DB) error {
	all := func(db dbm.DB) map[string][]byte {
		m := map[string][]byte{}
		it, err := db.Iterator(nil, nil)
		if err != nil {
			panic(err)
		}
		defer it.Close()
		for ; it.Valid(); it.Next() {
			if string(it.Key()) != "blockStoreFlushPoint" {
				m[string(it.Key())] = slices.Clone(it.Value())
			}
		}
		return m
	}
	g, w := all(got), all(want)
	for k, v := range w {
		if !bytes.Equal(g[k], v) {
			return fmt.Errorf("%s key %q is %X, want %X", name, k, g[k], v)
		}
	}
	for k := range g {
		if _, ok := w[k]; !ok {
			return fmt.Errorf("%s has key %q, which the run does not", name, k)
		}
	}

	return nil
}

// flushCutApp is a kvstore that reports the height and hash it last
// committed, as a persistent app does, returns each tx as its result's data,
// adds a validator and changes the consensus params on the way, defers its
// commits when asked (abci.DeferredCommitter), and notes the step at which
// each height it commits becomes durable.
type flushCutApp struct {
	*kvstore.KVStoreApplication

	clock     *int
	initial   int64
	height    int64
	hash      []byte
	deferring bool
	held      int   // heights committed since the last WriteCommits, while deferring
	durable   []int // durable[i] is the step height initial+i became durable at
}

var _ abci.DeferredCommitter = (*flushCutApp)(nil)

func newFlushCutApp(clock *int, initial int64) *flushCutApp {
	return &flushCutApp{KVStoreApplication: kvstore.NewKVStoreApplication(), clock: clock, initial: initial, height: initial - 1}
}

func (app *flushCutApp) Info(req abci.RequestInfo) abci.ResponseInfo {
	res := app.KVStoreApplication.Info(req)
	if len(app.durable) > 0 || app.held > 0 {
		res.LastBlockHeight = app.height
		res.LastBlockAppHash = app.hash
	}

	return res
}

func (app *flushCutApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	res := app.KVStoreApplication.DeliverTx(req)
	res.Data = req.Tx

	return res
}

func (app *flushCutApp) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock {
	var res abci.ResponseEndBlock
	switch req.Height - app.initial {
	case flushCutWindow:
		// The first block of the second window, the one a restart replays
		// after a crash between that window's first two writes: a second
		// validator, with too little power to matter for the commits.
		pubKey := ed25519.GenPrivKeyFromSecret([]byte("flush-cut")).PubKey()
		res.ValidatorUpdates = []abci.ValidatorUpdate{{Address: pubKey.Address(), PubKey: pubKey, Power: 1}}
	case flushCutWindow + 1:
		params := *types.DefaultBlockParams()
		params.MaxGas = 10_000_000
		res.ConsensusParams = &abci.ConsensusParams{Block: &params}
	}

	return res
}

func (app *flushCutApp) Commit() abci.ResponseCommit {
	res := app.KVStoreApplication.Commit()
	app.hash = res.Data
	app.height++
	if app.deferring {
		app.held++
	} else {
		*app.clock++
		app.durable = append(app.durable, *app.clock)
	}

	return res
}

func (app *flushCutApp) DeferCommits(on bool) bool {
	app.deferring = on

	return on
}

func (app *flushCutApp) HeldBytes() int64 { return int64(app.held) }

func (app *flushCutApp) WriteCommits() error {
	*app.clock++
	for ; app.held > 0; app.held-- {
		app.durable = append(app.durable, *app.clock)
	}

	return nil
}

func (app *flushCutApp) NeedsFlushPoint() bool { return false }

// heightAt returns the last height the app had made durable at step.
func (app *flushCutApp) heightAt(step int) int64 {
	height := app.initial - 1
	for _, s := range app.durable {
		if s <= step {
			height++
		}
	}
	if height < app.initial {
		return 0
	}

	return height
}

// cutDisk is a memdb that logs every write, batch and flush, so a test can
// rebuild the database a restart finds after a power cut at any step.
type cutDisk struct {
	dbm.DB

	clock  *int
	groups []cutGroup
}

// cutGroup is one write or one batch: what a power cut keeps or loses as a
// whole.
type cutGroup struct {
	step    int // when it was written
	flushed int // when its flush completed, or 0
	writes  []cutWrite
}

type cutWrite struct{ key, value []byte }

func (d *cutDisk) Set(key, value []byte) error {
	d.log([]cutWrite{{slices.Clone(key), slices.Clone(value)}}, false)
	return d.DB.Set(key, value)
}

// SetSync takes two steps, as a power cut can come after the write and before
// the flush.
func (d *cutDisk) SetSync(key, value []byte) error {
	d.log([]cutWrite{{slices.Clone(key), slices.Clone(value)}}, true)
	return d.DB.SetSync(key, value)
}

func (d *cutDisk) Delete([]byte) error { panic("cutDisk does not model deletes") }

func (d *cutDisk) DeleteSync([]byte) error { panic("cutDisk does not model deletes") }

func (d *cutDisk) NewBatch() dbm.Batch { return &cutBatch{disk: d} }

func (d *cutDisk) NewBatchWithSize(int) dbm.Batch { return &cutBatch{disk: d} }

func (d *cutDisk) log(writes []cutWrite, flush bool) {
	*d.clock++
	g := cutGroup{step: *d.clock, writes: writes}
	if flush {
		*d.clock++
		g.flushed = *d.clock
	}
	d.groups = append(d.groups, g)
}

// at splits the groups written up to step into those the last flush up to
// step made durable, and those written after it.
func (d *cutDisk) at(step int) (durable, unflushed []cutGroup) {
	flushed := 0
	for _, g := range d.groups {
		if g.flushed != 0 && g.flushed <= step {
			flushed = g.flushed
		}
	}
	for _, g := range d.groups {
		switch {
		case g.step < flushed:
			durable = append(durable, g)
		case g.step <= step:
			unflushed = append(unflushed, g)
		}
	}

	return durable, unflushed
}

type cutBatch struct {
	disk   *cutDisk
	writes []cutWrite
}

func (b *cutBatch) Set(key, value []byte) error {
	b.writes = append(b.writes, cutWrite{slices.Clone(key), slices.Clone(value)})
	return nil
}

func (b *cutBatch) Delete([]byte) error { panic("cutDisk does not model deletes") }

func (b *cutBatch) Write() error { return b.write(false) }

func (b *cutBatch) WriteSync() error { return b.write(true) }

func (b *cutBatch) write(flush bool) error {
	b.disk.log(b.writes, flush)
	for _, w := range b.writes {
		b.disk.DB.Set(w.key, w.value)
	}
	b.writes = nil
	return nil
}

func (b *cutBatch) Close() error { return nil }

func (b *cutBatch) GetByteSize() (int, error) { return 0, nil }

// flushCutVariants lists the sets of unflushed groups a power cut can keep:
// every prefix, each group whole.
func flushCutVariants(unflushed []cutGroup) [][]cutGroup {
	variants := make([][]cutGroup, 0, len(unflushed)+1)
	for kept := 0; kept <= len(unflushed); kept++ {
		variants = append(variants, unflushed[:kept])
	}

	return variants
}

// rebuildGroups returns a database holding the groups' writes, in order.
func rebuildGroups(groups []cutGroup) dbm.DB {
	db := memdb.NewMemDB()
	for _, g := range groups {
		for _, w := range g.writes {
			db.Set(w.key, w.value)
		}
	}

	return db
}

// TestLowerAfterFlushPointCrash pins when the handshake lowers a block store
// that is more than one block ahead of the state: only when the flush point's
// record matches the store and the state and the app is not ahead of the
// state, the shape a crash between a flush point's first two writes leaves.
// Any other shape is refused, with the store's height left as it is on disk.
func TestLowerAfterFlushPointCrash(t *testing.T) {
	t.Parallel()

	const storeHeight, stateHeight = 6, 2
	for _, tc := range []struct {
		name      string
		record    *store.FlushPoint
		appHeight int64
		lowered   bool
	}{
		{"the crash shape", &store.FlushPoint{From: stateHeight, To: storeHeight}, stateHeight, true},
		{"app below the state", &store.FlushPoint{From: stateHeight, To: storeHeight}, 0, true},
		{"no record", nil, stateHeight, false},
		{"record from another state", &store.FlushPoint{From: stateHeight - 1, To: storeHeight}, stateHeight, false},
		{"record to another store", &store.FlushPoint{From: stateHeight, To: storeHeight - 1}, stateHeight, false},
		{"app ahead of the state", &store.FlushPoint{From: stateHeight, To: storeHeight}, stateHeight + 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := memdb.NewMemDB()
			store.BlockStoreStateJSON{Height: storeHeight}.Save(db)
			if tc.record != nil {
				tc.record.Save(db)
			}
			genDoc, _ := randGenesisDoc(1, false, 10)
			state, err := sm.MakeGenesisState(genDoc)
			require.NoError(t, err)
			h := NewHandshaker(memdb.NewMemDB(), state, store.NewBlockStore(db), genDoc)

			require.Equal(t, tc.lowered, h.lowerAfterFlushPointCrash(storeHeight, stateHeight, tc.appHeight))
			want := int64(storeHeight)
			if tc.lowered {
				want = stateHeight + 1
			}
			require.Equal(t, want, store.NewBlockStore(db).Height(), "the block store's height on disk")
		})
	}
}
