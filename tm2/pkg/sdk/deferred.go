package sdk

import (
	"errors"
	"fmt"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Deferred commits: during block sync the node asks the application to hold
// its commits in memory, and writes them at flush points together with its
// own records (tm2/adr, deferred writes during block sync). BaseApp defers
// only with AllowDeferredCommits, and only over a multistore that can hold
// commits.

var _ abci.DeferredCommitter = (*BaseApp)(nil)

// errNothingWritten answers queries while commits are held and nothing is on
// disk yet, as on a fresh node before its first flush point.
var errNothingWritten = std.ErrInternal("no state is written yet; the node is syncing")

// holder returns the multistore as a CommitHolder, or nil if it cannot hold.
func (app *BaseApp) holder() store.CommitHolder {
	h, _ := app.cms.(store.CommitHolder)
	return h
}

// deferringCommits reports whether commits are held.
func (app *BaseApp) deferringCommits() bool {
	h := app.holder()
	return h != nil && h.Holding()
}

// DeferCommits implements abci.DeferredCommitter.
func (app *BaseApp) DeferCommits(on bool) bool {
	h := app.holder()
	if !on {
		if h != nil {
			h.SetHolding(false) // panics with commits still held
		}
		return false
	}
	if !app.deferAllowed || h == nil {
		return false
	}
	h.SetHolding(true)
	return true
}

// WriteCommits implements abci.DeferredCommitter. Queries see the commits
// once they are written.
func (app *BaseApp) WriteCommits() error {
	if !app.deferringCommits() {
		return errors.New("baseapp: WriteCommits without deferred commits")
	}
	if err := app.holder().WriteHeld(); err != nil {
		return err
	}
	if app.heldHeader != nil {
		app.lastBlockHeader.Store(&headerSnapshot{header: app.heldHeader})
		app.heldHeader = nil
	}
	app.flushRequired.Store(false)
	return nil
}

// HeldBytes implements abci.DeferredCommitter.
func (app *BaseApp) HeldBytes() int64 {
	if h := app.holder(); h != nil {
		return h.HeldBytes()
	}
	return 0
}

// NeedsFlushPoint implements abci.DeferredCommitter. It compares the halt
// height with the version committed in memory: the version on disk stays
// below it until the flush point.
func (app *BaseApp) NeedsFlushPoint() bool {
	return app.haltHeight > 0 && uint64(app.LastBlockHeight()) >= app.haltHeight ||
		app.flushRequired.Load()
}

// RequireFlushPoint makes NeedsFlushPoint report true after the current
// block, so that what the block wrote is on disk before the next block runs.
// gno.land calls it when a block writes a halt param, which the binary
// started after the halt reads from the disk.
func (app *BaseApp) RequireFlushPoint() { app.flushRequired.Store(true) }

// writtenCommitID is the latest commit on disk, which queries answer from. It
// is LastCommitID except while commits are held.
func (app *BaseApp) writtenCommitID() store.CommitID {
	if h := app.holder(); h != nil {
		return h.WrittenCommitID()
	}
	return app.cms.LastCommitID()
}

// nothingWritten reports that commits are held while nothing is on disk yet:
// there is no version to answer queries from.
func (app *BaseApp) nothingWritten() bool {
	return app.deferringCommits() && app.writtenCommitID().Version == 0
}

// clearRefusedIterators and checkRefusedIterators run before and after each
// of a block's calls into the application: checkRefusedIterators panics if
// the call opened an iterator refused because commits are held. A recover
// inside the call, such as runTx's or the VM's, would otherwise turn the
// refusal into a failed transaction: a result no node in consensus has,
// which a flush point would then write. The check runs outside every
// recover, and CheckTx never overlaps a block's calls (they share the local
// client's mutex), so a refusal it records is cleared before the next one.
func (app *BaseApp) clearRefusedIterators() {
	if app.deferringCommits() {
		app.holder().TakeRefusedIterator()
	}
}

func (app *BaseApp) checkRefusedIterators() {
	if app.deferringCommits() && app.holder().TakeRefusedIterator() {
		panic(fmt.Errorf("%w: refused inside a block while commits are held", dbm.ErrIteratorOverHeldWrites))
	}
}
