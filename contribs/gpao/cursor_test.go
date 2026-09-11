package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The cursor is written by the verifier goroutine, so the assertions on it
// wait rather than read straight after enqueueing.
const (
	cursorWaitFor = 2 * time.Second
	cursorTick    = 5 * time.Millisecond
)

// cursorOnDisk re-reads the recorded height through the same loader production
// uses, so there is only ever one decoder for the on-disk format.
//
// The file, not stateStore's own field: the store is single-writer by design
// (see state.go), so a test watching the verifier goroutine has to observe the
// durable artifact rather than race on the in-memory copy. Which is the better
// assertion anyway -- the file is what a restart reads.
//
// Returns an error rather than calling require: it is polled from inside
// require.Eventually, which runs its condition on another goroutine, where a
// testify assertion would report a confusing failure rather than this one.
func cursorOnDisk(s *stateStore) (int64, error) {
	reopened, err := openStateStore(filepath.Dir(s.path), s.chainID)
	if err != nil {
		return 0, err
	}
	return reopened.lastVerifiedHeight(), nil
}

// tipRPC answers Status with a fixed tip. The embedded interface is nil on
// purpose: anything startHeight calls beyond Status will panic rather than
// quietly return a zero value.
type tipRPC struct {
	rpcclient.Client
	tip int64
}

func (r tipRPC) Status(context.Context, *int64) (*ctypes.ResultStatus, error) {
	return &ctypes.ResultStatus{
		SyncInfo: ctypes.SyncInfo{LatestBlockHeight: r.tip},
	}, nil
}

func newCursorOracle(t *testing.T, tip int64) *oracle {
	t.Helper()
	return newStubOracle(t, tipRPC{tip: tip})
}

// TestStartHeightFallsBackToTheTip: with nothing recorded and no flag, the
// behaviour is what it was before there was a cursor.
func TestStartHeightFallsBackToTheTip(t *testing.T) {
	o := newCursorOracle(t, 500)

	h, answered, err := o.startHeight(context.Background())
	require.NoError(t, err)
	require.True(t, answered)
	assert.Equal(t, int64(501), h)
}

// TestStartHeightResumesFromTheCursor is the point of the change: after a
// crash, the operator restarts with no flag and the oracle picks up where the
// last run finished.
func TestStartHeightResumesFromTheCursor(t *testing.T) {
	o := newCursorOracle(t, 500)
	require.NoError(t, o.state.setLastVerifiedHeight(120))

	h, answered, err := o.startHeight(context.Background())
	require.NoError(t, err)
	require.True(t, answered)
	assert.Equal(t, int64(121), h, "resume past the last height that was finished, not at it")
}

// TestStartHeightFlagOverridesAndRewritesTheCursor: an operator replaying a
// range has to be able to contradict a recorded height, and the record has to
// follow -- otherwise the next bare restart jumps back to where the cursor was
// and undoes the replay.
func TestStartHeightFlagOverridesAndRewritesTheCursor(t *testing.T) {
	o := newCursorOracle(t, 500)
	require.NoError(t, o.state.setLastVerifiedHeight(400))
	o.cfg.startHeight = 10

	h, answered, err := o.startHeight(context.Background())
	require.NoError(t, err)
	require.True(t, answered)
	assert.Equal(t, int64(10), h)
	assert.Equal(t, int64(9), o.state.lastVerifiedHeight(),
		"the cursor names the last height finished, so the flag stores one below it")
}

// TestStartHeightResumesFromARecordedZero is the case a single "is there a
// cursor" integer has to get right.
//
// `-start-height 1` records "verified through 0", so a bare restart after such
// a run must resume at 1. Reading that 0 back as "nothing recorded" would send
// it to the tip instead and silently abandon the replay the operator asked for
// -- the exact conflation noCursor exists to prevent.
func TestStartHeightResumesFromARecordedZero(t *testing.T) {
	o := newCursorOracle(t, 500)
	require.NoError(t, o.state.reset(0))

	h, answered, err := o.startHeight(context.Background())
	require.NoError(t, err)
	require.True(t, answered)
	assert.Equal(t, int64(1), h, "a recorded 0 resumes at 1, it does not mean the tip")
}

// TestStartHeightRefusesACursorAheadOfTheChain: a cursor above the tip means
// this state was written for a chain that has since been reset, or against a
// node with history this one does not have. Waiting it out is indistinguishable
// from an oracle that works, so it stops instead.
func TestStartHeightRefusesACursorAheadOfTheChain(t *testing.T) {
	o := newCursorOracle(t, 12)
	require.NoError(t, o.state.setLastVerifiedHeight(4218))

	_, _, err := o.startHeight(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "4218")
	assert.Contains(t, err.Error(), "12")
	assert.Contains(t, err.Error(), "--start-height", "the error has to name a way out")
}

// TestStartHeightAcceptsACursorAtTheTip: caught up is not ahead. Off by one
// here would refuse to start every time the oracle had kept pace.
func TestStartHeightAcceptsACursorAtTheTip(t *testing.T) {
	o := newCursorOracle(t, 500)
	require.NoError(t, o.state.setLastVerifiedHeight(500))

	h, answered, err := o.startHeight(context.Background())
	require.NoError(t, err)
	require.True(t, answered)
	assert.Equal(t, int64(501), h)
}

// TestVerifierAdvancesTheCursorOverPackageFreeBlocks: most blocks submit
// nothing, and if those did not move the cursor it would sit at the last block
// that happened to carry a package -- so a restart on a quiet chain would
// re-read every height since.
func TestVerifierAdvancesTheCursorOverPackageFreeBlocks(t *testing.T) {
	o := newCursorOracle(t, 500)

	ctx := t.Context()
	go o.runVerifier(ctx)

	for h := int64(1); h <= 3; h++ {
		require.NoError(t, o.enqueue(ctx, blockWork{height: h}))
	}
	require.Eventually(t, func() bool {
		h, err := cursorOnDisk(o.state)
		return err == nil && h == 3
	}, cursorWaitFor, cursorTick, "the cursor must advance over blocks with nothing to verify")
}

// TestVerifierRecordsTheHeightAfterItsPackages: the cursor claims a block is
// finished, so it may only be written once every package in that block has been
// through handleCandidate. Driven with a package already in `seen`, which is
// the one path through handleCandidate that spawns nothing.
func TestVerifierRecordsTheHeightAfterItsPackages(t *testing.T) {
	o := newCursorOracle(t, 500)
	mpkg := &std.MemPackage{Name: "p", Path: "gno.land/r/test/p"}
	o.seen[candidateKey(mpkg)] = struct{}{}

	ctx := t.Context()
	go o.runVerifier(ctx)

	require.NoError(t, o.enqueue(ctx, blockWork{height: 8, pkgs: []*std.MemPackage{mpkg}}))
	require.Eventually(t, func() bool {
		h, err := cursorOnDisk(o.state)
		return err == nil && h == 8
	}, cursorWaitFor, cursorTick)
}

// TestSpendBoundFreezesTheCursor pins the interaction that would otherwise make
// -max-spend lossy. See the spendBlocked field for why it has to.
func TestSpendBoundFreezesTheCursor(t *testing.T) {
	o := newCursorOracle(t, 500)
	require.NoError(t, o.state.setLastVerifiedHeight(30))

	o.spendBlocked = true
	o.recordVerified(31)

	assert.Equal(t, int64(30), o.state.lastVerifiedHeight(),
		"a block whose package was declined for want of budget is not verified")
}
