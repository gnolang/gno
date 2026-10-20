package blockchain

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/mempool/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

// deferringApp is testApp deferring its commits when asked, unless refuse is
// set. It counts its commits and writes, runs onWrite inside each
// WriteCommits, and asks for a flush point after the heights in flushAt, and
// at every call once flushNow is set.
type deferringApp struct {
	testApp

	mu          sync.Mutex
	refuse      bool
	deferring   bool
	committed   int64 // the height committed, in memory
	held        int   // commits since the last write, while deferring
	written     int   // commits on disk
	writes      int   // WriteCommits calls
	commitDelay time.Duration
	flushAt     map[int64]bool
	flushNow    bool
	failWrite   bool
	onWrite     func()
}

var _ abci.DeferredCommitter = (*deferringApp)(nil)

func (app *deferringApp) Commit() abci.ResponseCommit {
	time.Sleep(app.commitDelay)
	app.mu.Lock()
	defer app.mu.Unlock()
	app.committed++
	if app.deferring {
		app.held++
	} else {
		app.written++
	}
	return abci.ResponseCommit{}
}

func (app *deferringApp) DeferCommits(on bool) bool {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.deferring = on && !app.refuse
	return app.deferring
}

func (app *deferringApp) WriteCommits() error {
	if app.onWrite != nil {
		app.onWrite()
	}
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.failWrite {
		return errors.New("disk full")
	}
	app.written += app.held
	app.held = 0
	app.writes++
	return nil
}

func (app *deferringApp) HeldBytes() int64 {
	app.mu.Lock()
	defer app.mu.Unlock()
	return int64(app.held)
}

func (app *deferringApp) NeedsFlushPoint() bool {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.flushAt[app.committed] || app.flushNow
}

// deferringSyncer is an empty node of a new one-validator chain that
// block-syncs with deferred writes.
type deferringSyncer struct {
	reactor              *BlockchainReactor
	blockDB, stateDB     *dbm.HoldingDB
	blockDisk, stateDisk dbm.DB
	genDoc               *types.GenesisDoc
	privVals             []types.PrivValidator
	switched             chan switched // what the switch to consensus found

	mu     sync.Mutex
	events []int64 // the heights of the EventNewBlock events fired, in order
}

// switched is what a switch to consensus found.
type switched struct {
	height  int64
	holding bool // a database or the app still held writes
}

// newDeferringSyncer returns a syncer whose flush points come only from app
// and from the switch to consensus, not from the interval.
func newDeferringSyncer(t *testing.T, app *deferringApp) *deferringSyncer {
	t.Helper()

	config, _ = cfg.ResetTestRoot("blockchain_reactor_test")
	root := config.RootDir
	t.Cleanup(func() { os.RemoveAll(root) })
	genDoc, privVals := randGenesisDoc(1, false, 30)
	s := &deferringSyncer{
		blockDisk: memdb.NewMemDB(), stateDisk: memdb.NewMemDB(),
		genDoc: genDoc, privVals: privVals, switched: make(chan switched, 1),
	}
	s.blockDB, s.stateDB = dbm.NewHoldingDB(s.blockDisk), dbm.NewHoldingDB(s.stateDisk)
	proxyApp := appconn.NewAppConns(proxy.NewLocalClientCreator(app))
	require.NoError(t, proxyApp.Start())
	t.Cleanup(func() { proxyApp.Stop() })

	state, err := sm.LoadStateFromDBOrGenesisDoc(s.stateDB, genDoc)
	require.NoError(t, err)
	sm.SaveState(s.stateDB, state)

	evsw := events.NewEventSwitch()
	require.NoError(t, evsw.Start())
	t.Cleanup(func() { evsw.Stop() })
	evsw.AddListener("deferred-test", func(ev events.Event) {
		if nb, ok := ev.(types.EventNewBlock); ok {
			s.mu.Lock()
			s.events = append(s.events, nb.Block.Height)
			s.mu.Unlock()
		}
	})

	blockExec := sm.NewBlockExecutor(s.stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})
	blockExec.SetEventSwitch(evsw)
	s.reactor = NewBlockchainReactor(state.Copy(), blockExec, store.NewBlockStore(s.blockDB), true, func(state sm.State, _ int) {
		app.mu.Lock()
		holding := app.deferring
		app.mu.Unlock()
		s.switched <- switched{state.LastBlockHeight, holding || s.blockDB.Holding() || s.stateDB.Holding()}
	})
	s.reactor.SetLogger(log.NewTestingLogger(t).With("module", "blockchain"))
	s.reactor.SetDeferredWrites(app)
	s.reactor.deferred.interval = time.Hour
	return s
}

// syncFrom connects the syncer to a node of its chain holding height blocks,
// and returns that node's switch.
func (s *deferringSyncer) syncFrom(t *testing.T, height int64) *p2p.MultiplexSwitch {
	t.Helper()

	source := newBlockchainReactor(log.NewTestingLogger(t), s.genDoc, s.privVals, height)
	t.Cleanup(func() {
		source.reactor.Stop()
		source.app.Stop()
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sws, _ := p2pTesting.MakeConnectedPeers(t, ctx, p2pTesting.TestingConfig{
		Count:  2,
		P2PCfg: config.P2P,
		SwitchOptions: map[int][]p2p.SwitchOption{
			0: {p2p.WithReactor("BLOCKCHAIN", source.reactor)},
			1: {p2p.WithReactor("BLOCKCHAIN", s.reactor)},
		},
		Channels: []byte{BlockchainChannel},
	})
	return sws[0]
}

// waitHeight waits until the syncer has saved height blocks.
func (s *deferringSyncer) waitHeight(t *testing.T, height int64) {
	t.Helper()

	require.Eventually(t, func() bool { return s.reactor.store.Height() >= height },
		30*time.Second, 5*time.Millisecond, "block sync did not reach height %d", height)
}

func (s *deferringSyncer) waitRoutine(t *testing.T) {
	t.Helper()

	select {
	case <-s.reactor.routineDone:
	case <-time.After(30 * time.Second):
		t.Fatal("the block-sync routine did not return")
	}
}

func (s *deferringSyncer) waitSwitch(t *testing.T) switched {
	t.Helper()

	select {
	case sw := <-s.switched:
		return sw
	case <-time.After(60 * time.Second):
		t.Fatal("block sync did not switch to consensus")
		return switched{}
	}
}

func (s *deferringSyncer) requireNoSwitch(t *testing.T, msg string) {
	t.Helper()

	select {
	case <-s.switched:
		t.Fatal(msg)
	default:
	}
}

// TestBlockSyncDefersWrites pins deferred block sync end to end: the app's
// own flush points and the one before the switch are taken; by the time it
// switches to consensus, every block, the state and the app's commits are on
// disk and nothing is held any more; each block's events fired once, in
// order.
func TestBlockSyncDefersWrites(t *testing.T) {
	app := &deferringApp{flushAt: map[int64]bool{10: true, 30: true}}
	s := newDeferringSyncer(t, app)
	s.syncFrom(t, 65)

	sw := s.waitSwitch(t)
	s.waitRoutine(t)
	require.False(t, sw.holding, "writes still held when consensus starts")
	require.Positive(t, sw.height)
	require.Equal(t, sw.height, store.NewBlockStore(s.blockDisk).Height(), "blocks held at the switch")
	require.Equal(t, sw.height, sm.LoadState(s.stateDisk).LastBlockHeight, "state held at the switch")
	app.mu.Lock()
	require.Equal(t, int(sw.height), app.written, "app commits held at the switch")
	require.Equal(t, 3, app.writes, "flush points: the app's two and the switch's")
	app.mu.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	require.Len(t, s.events, int(sw.height))
	for i, h := range s.events {
		require.Equal(t, int64(i+1), h, "events out of order or repeated")
	}
}

// TestBlockSyncWithoutDeferral pins that block sync writes every block as it
// applies it when the app does not defer its commits.
func TestBlockSyncWithoutDeferral(t *testing.T) {
	app := &deferringApp{refuse: true}
	s := newDeferringSyncer(t, app)
	app.onWrite = func() { t.Error("WriteCommits called on an app that does not defer") }
	s.syncFrom(t, 65)

	sw := s.waitSwitch(t)
	s.waitRoutine(t)
	require.False(t, sw.holding)
	require.Equal(t, sw.height, store.NewBlockStore(s.blockDisk).Height())
	app.mu.Lock()
	defer app.mu.Unlock()
	require.Equal(t, int(sw.height), app.written)
}

// TestBlockSyncStopDuringFlushPointDoesNotSwitch pins that a stop arriving
// while the flush point before the switch is writing ends block sync after
// it, without starting consensus, and that OnStop waits for it.
func TestBlockSyncStopDuringFlushPointDoesNotSwitch(t *testing.T) {
	app := &deferringApp{}
	inFlush, release := make(chan struct{}), make(chan struct{})
	app.onWrite = func() {
		close(inFlush) // the only flush point is the switch's
		<-release
	}
	s := newDeferringSyncer(t, app)
	s.syncFrom(t, 65)

	select {
	case <-inFlush:
	case <-time.After(60 * time.Second):
		t.Fatal("no flush point before the switch")
	}
	stopped := make(chan struct{})
	go func() {
		s.reactor.Stop()
		close(stopped)
	}()
	require.Eventually(t, func() bool { return !s.reactor.pool.IsRunning() },
		10*time.Second, time.Millisecond, "OnStop did not stop the pool")
	select {
	case <-stopped:
		t.Fatal("OnStop returned before the flush point finished")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)

	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		t.Fatal("OnStop did not return after the flush point")
	}
	s.requireNoSwitch(t, "block sync switched to consensus while stopping")
	app.mu.Lock()
	defer app.mu.Unlock()
	require.Zero(t, app.held, "the flush point did not finish")
}

// TestBlockSyncStopBeforeCaughtUp pins that a stop while block sync is
// behind returns, drops what is held, as a crash would, and writes nothing.
func TestBlockSyncStopBeforeCaughtUp(t *testing.T) {
	app := &deferringApp{commitDelay: 20 * time.Millisecond}
	s := newDeferringSyncer(t, app)
	s.syncFrom(t, 65)

	s.waitHeight(t, 5)
	stopped := make(chan struct{})
	go func() {
		s.reactor.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		t.Fatal("OnStop hung while block sync was behind")
	}

	s.requireNoSwitch(t, "block sync switched to consensus while stopping")
	require.Zero(t, store.NewBlockStore(s.blockDisk).Height(), "a held block was written")
	app.mu.Lock()
	defer app.mu.Unlock()
	require.Zero(t, app.written, "a held commit was written")
	require.Less(t, app.committed, int64(64), "block sync caught up before the stop")
}

// TestBlockSyncFlushesAStalledWindow pins that a window whose blocks stop
// arriving, as when the only peer leaves, is written by the routine's ticker,
// without waiting for another block or for the switch to consensus.
func TestBlockSyncFlushesAStalledWindow(t *testing.T) {
	app := &deferringApp{commitDelay: 20 * time.Millisecond}
	s := newDeferringSyncer(t, app)
	source := s.syncFrom(t, 65)

	s.waitHeight(t, 5)
	require.NoError(t, source.Stop())
	// The switch removes a peer from the pool before it drops it from its set.
	require.Eventually(t, func() bool { return len(s.reactor.Switch.Peers().List()) == 0 },
		10*time.Second, 5*time.Millisecond, "the peer did not leave")
	// The block being applied when the peer left is the last one.
	time.Sleep(100 * time.Millisecond)
	held := s.reactor.store.Height()
	require.Zero(t, store.NewBlockStore(s.blockDisk).Height(), "a flush point before the stall")

	app.mu.Lock()
	app.flushNow = true
	app.mu.Unlock()
	// The app's write is the flush point's last.
	require.Eventually(t, func() bool {
		app.mu.Lock()
		defer app.mu.Unlock()
		return app.written == int(held)
	}, 5*time.Second, 5*time.Millisecond, "the stalled window was not written")
	require.Equal(t, held, store.NewBlockStore(s.blockDisk).Height())
	require.Equal(t, held, sm.LoadState(s.stateDisk).LastBlockHeight)
	s.requireNoSwitch(t, "block sync switched to consensus without a peer")
}

// TestOnStopWithoutFastSync pins that OnStop returns when the block-sync
// routine never started, as for a node that is the only validator.
func TestOnStopWithoutFastSync(t *testing.T) {
	t.Parallel()

	genDoc, _ := randGenesisDoc(1, false, 30)
	stateDB := memdb.NewMemDB()
	state, err := sm.LoadStateFromDBOrGenesisDoc(stateDB, genDoc)
	require.NoError(t, err)
	proxyApp := appconn.NewAppConns(proxy.NewLocalClientCreator(&testApp{}))
	require.NoError(t, proxyApp.Start())
	defer proxyApp.Stop()
	blockExec := sm.NewBlockExecutor(stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})
	r := NewBlockchainReactor(state, blockExec, store.NewBlockStore(memdb.NewMemDB()), false, nil)
	r.SetLogger(log.NewTestingLogger(t))

	require.NoError(t, r.Start())
	done := make(chan struct{})
	go func() {
		r.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("OnStop hung with fast sync off")
	}
}

// newTestDeferredWrites returns DeferredWrites over two HoldingDBs on the
// given disks and app, started at height 0.
func newTestDeferredWrites(t *testing.T, blockDisk, stateDisk dbm.DB, app *deferringApp) (*DeferredWrites, *dbm.HoldingDB, *dbm.HoldingDB) {
	t.Helper()

	blockDB, stateDB := dbm.NewHoldingDB(blockDisk), dbm.NewHoldingDB(stateDisk)
	proxyApp := appconn.NewAppConns(proxy.NewLocalClientCreator(app))
	require.NoError(t, proxyApp.Start())
	t.Cleanup(func() { proxyApp.Stop() })
	blockExec := sm.NewBlockExecutor(stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})
	d := NewDeferredWrites(store.NewBlockStore(blockDB), blockExec, app)
	require.NotNil(t, d)
	require.True(t, d.Start(0))
	return d, blockDB, stateDB
}

// TestDeferredWritesNeedsFlushPoint pins each reason for a flush point: the
// interval, each of the three databases over the budget, and the app's own.
func TestDeferredWritesNeedsFlushPoint(t *testing.T) {
	t.Parallel()

	app := &deferringApp{flushAt: map[int64]bool{}}
	d, blockDB, stateDB := newTestDeferredWrites(t, memdb.NewMemDB(), memdb.NewMemDB(), app)
	d.interval, d.maxBytes = time.Hour, 4
	require.False(t, d.NeedsFlushPoint())

	d.interval = 0
	require.True(t, d.NeedsFlushPoint(), "interval")
	d.interval = time.Hour

	blockDB.Set([]byte("block"), []byte("1"))
	require.True(t, d.NeedsFlushPoint(), "block store over the budget")
	d.FlushPoint(1)
	require.False(t, d.NeedsFlushPoint())

	stateDB.Set([]byte("state"), []byte("1"))
	require.True(t, d.NeedsFlushPoint(), "state over the budget")
	d.FlushPoint(2)

	app.held = 5
	require.True(t, d.NeedsFlushPoint(), "app over the budget")
	d.FlushPoint(3)

	app.committed, app.flushAt[4] = 4, true
	require.True(t, d.NeedsFlushPoint(), "the app's own")
}

// TestDeferredWritesIntervalFromTheEnd pins that the interval runs from the
// end of the last flush point: one slower than the interval does not leave the
// next one due at once.
func TestDeferredWritesIntervalFromTheEnd(t *testing.T) {
	t.Parallel()

	app := &deferringApp{onWrite: func() { time.Sleep(300 * time.Millisecond) }}
	d, blockDB, _ := newTestDeferredWrites(t, memdb.NewMemDB(), memdb.NewMemDB(), app)
	d.interval = 200 * time.Millisecond
	blockDB.Set([]byte("block"), []byte("1"))
	d.FlushPoint(1)
	require.False(t, d.NeedsFlushPoint(), "a slow flush point left the next one due")
}

// TestDeferredWritesFlushPoint pins each flush point's writes: the block
// store, with a record of the window starting where the last one ended,
// then the state, both on disk by the time the app writes; nothing for an
// empty window.
func TestDeferredWritesFlushPoint(t *testing.T) {
	t.Parallel()

	blockDisk, stateDisk := memdb.NewMemDB(), memdb.NewMemDB()
	app := &deferringApp{}
	d, blockDB, stateDB := newTestDeferredWrites(t, blockDisk, stateDisk, app)
	var window string
	app.onWrite = func() {
		for _, disk := range []dbm.DB{blockDisk, stateDisk} {
			if v, _ := disk.Get([]byte(window)); v == nil {
				t.Errorf("%s not on disk when the app writes", window)
			}
		}
	}

	for _, want := range []store.FlushPoint{{From: 0, To: 3}, {From: 3, To: 7}} {
		window = string(rune('a' + want.To))
		blockDB.Set([]byte(window), []byte("1"))
		stateDB.Set([]byte(window), []byte("1"))
		d.FlushPoint(want.To)
		fp, ok := store.NewBlockStore(blockDisk).FlushPoint()
		require.True(t, ok)
		require.Equal(t, want, fp)
	}
	require.Equal(t, 2, app.writes)

	d.FlushPoint(7)
	require.Equal(t, 2, app.writes, "an empty window wrote")
}

// failingBatchDB is a database whose batches fail to write.
type failingBatchDB struct{ dbm.DB }

func (db failingBatchDB) NewBatch() dbm.Batch { return failingBatch{db.DB.NewBatch()} }

func (db failingBatchDB) NewBatchWithSize(size int) dbm.Batch {
	return failingBatch{db.DB.NewBatchWithSize(size)}
}

type failingBatch struct{ dbm.Batch }

func (failingBatch) WriteSync() error { return errors.New("disk full") }

// TestFlushPointWriteErrorEndsTheNode pins that a failed write at any step of
// a flush point panics, since going on would write the next window over a
// gap, and that nothing after the failed write is written.
func TestFlushPointWriteErrorEndsTheNode(t *testing.T) {
	t.Parallel()

	for _, failing := range []string{"block store", "state", "app"} {
		t.Run(failing, func(t *testing.T) {
			t.Parallel()

			blockDisk, stateDisk := dbm.DB(memdb.NewMemDB()), dbm.DB(memdb.NewMemDB())
			app := &deferringApp{}
			switch failing {
			case "block store":
				blockDisk = failingBatchDB{blockDisk}
			case "state":
				stateDisk = failingBatchDB{stateDisk}
			case "app":
				app.failWrite = true
			}
			d, blockDB, stateDB := newTestDeferredWrites(t, blockDisk, stateDisk, app)
			blockDB.Set([]byte("block"), []byte("1"))
			stateDB.Set([]byte("state"), []byte("1"))

			require.Panics(t, func() { d.FlushPoint(1) })
			if failing == "block store" {
				v, _ := stateDisk.Get([]byte("state"))
				require.Nil(t, v, "the state was written after the block store's write failed")
			}
			if failing != "app" {
				require.Zero(t, app.writes, "the app was written after an earlier write failed")
			}
		})
	}
}
