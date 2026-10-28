package blockchain

import (
	"fmt"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// DeferredWrites holds, during block sync, the writes of the block store, of
// the state and of the application in memory, and writes all of them at flush
// points (see FlushPoint and tm2/adr, deferred writes during block sync).
// Its methods run on the block-sync goroutine, between blocks.
type DeferredWrites struct {
	blockDB, stateDB *dbm.HoldingDB
	app              abci.DeferredCommitter
	blockExec        *sm.BlockExecutor

	interval  time.Duration // a flush point at least this often
	maxBytes  int64         // and when a database holds more
	flushed   int64         // the height on disk
	lastFlush time.Time
}

// NewDeferredWrites returns the DeferredWrites of blockStore and blockExec,
// whose databases must be HoldingDBs, and of their application; nil if the
// databases cannot hold.
func NewDeferredWrites(blockStore *store.BlockStore, blockExec *sm.BlockExecutor, app abci.DeferredCommitter) *DeferredWrites {
	blockDB, ok := blockStore.DB().(*dbm.HoldingDB)
	if !ok {
		return nil
	}
	stateDB, ok := blockExec.DB().(*dbm.HoldingDB)
	if !ok {
		return nil
	}
	return &DeferredWrites{
		blockDB: blockDB, stateDB: stateDB, app: app, blockExec: blockExec,
		interval: flushPointInterval, maxBytes: flushPointBytes,
	}
}

// Start starts holding the writes of the blocks applied after height, the
// height on disk, if the application agrees to defer its commits, and
// reports whether it did: deferring only the node's side would let the
// application's commits reach the disk ahead of the node's records, which no
// restart recovers from.
func (d *DeferredWrites) Start(height int64) bool {
	if !d.app.DeferCommits(true) {
		return false
	}
	d.blockDB.Hold(true)
	d.stateDB.Hold(true)
	d.blockExec.HoldEvents(true)
	d.flushed, d.lastFlush = height, time.Now()
	return true
}

// Stop stops holding writes, right after a flush point.
func (d *DeferredWrites) Stop() {
	d.blockDB.Hold(false)
	d.stateDB.Hold(false)
	d.app.DeferCommits(false)
	d.blockExec.HoldEvents(false)
}

// NeedsFlushPoint reports whether to take a flush point now: the interval
// since the last one has passed, a database holds too much, or the
// application needs one (its halt height, or a halt param written).
func (d *DeferredWrites) NeedsFlushPoint() bool {
	return time.Since(d.lastFlush) >= d.interval ||
		int64(d.blockDB.HeldBytes()) > d.maxBytes ||
		int64(d.stateDB.HeldBytes()) > d.maxBytes ||
		d.app.HeldBytes() > d.maxBytes ||
		d.app.NeedsFlushPoint()
}

// FlushPoint writes everything held since the last flush point, up to the
// block at height, in the order that leaves the three databases consistent
// after a crash at any instant: the block store, with a record of the
// window, then the state, then the application, each as one synced,
// all-or-nothing write. The window's events fire only then. A failed write
// ends the node: going on would write the next window over a gap.
func (d *DeferredWrites) FlushPoint(height int64) {
	from := d.flushed
	if height == from {
		d.lastFlush = time.Now()
		return // nothing applied since the last flush point
	}
	store.FlushPoint{From: from, To: height}.Save(d.blockDB)
	if err := d.blockDB.WriteHeld(); err != nil {
		panic(fmt.Sprintf("block sync: writing blocks %d to %d: %v", from+1, height, err))
	}
	if err := d.stateDB.WriteHeld(); err != nil {
		panic(fmt.Sprintf("block sync: writing the state of blocks %d to %d: %v", from+1, height, err))
	}
	if err := d.app.WriteCommits(); err != nil {
		panic(fmt.Sprintf("block sync: writing the application's commits of blocks %d to %d: %v", from+1, height, err))
	}
	d.flushed = height
	d.blockExec.FireHeldEvents()
	d.lastFlush = time.Now() // the interval runs from the end, however long this took
}
