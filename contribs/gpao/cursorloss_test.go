package main

import (
	"context"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestShutdownDoesNotRecordTheBlockItAbandoned: a clean shutdown must not move
// the cursor past work it dropped.
//
// Cancelling the context kills the verifier child, and verify() classifies that
// as a shutdown rather than a verdict -- it returns errVerifyBudget wrapped in
// "shutting down", and handleCandidate files the package `pending` with "will be
// retried". Before the ctx guards in runVerifier, recordVerified then ran
// anyway, so the status board said "will be retried" about a package the cursor
// guaranteed would never be reached again: the next bare restart resumed past
// its block and the package stayed inert for good.
//
// Up to one verify-budget wide per block, on every `systemctl restart gpao` --
// the workflow the cursor exists to serve.
//
// Driven through the real loop body rather than through runVerifier, because the
// select at the top of that function would take ctx.Done() and never reach the
// candidate: the bug lives after the dequeue, so the test has to start there.
func TestShutdownDoesNotRecordTheBlockItAbandoned(t *testing.T) {
	o := newCursorOracle(t, 500)
	require.NoError(t, o.state.setLastVerifiedHeight(4))

	const path = "gno.land/r/test/interrupted"
	mpkg := &std.MemPackage{
		Name: "interrupted",
		Path: path,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
			{Name: "interrupted.gno", Body: "package interrupted\n\nfunc F(cur realm) string { return \"x\" }\n"},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // SIGTERM, mid-block

	work := blockWork{height: 5, pkgs: []*std.MemPackage{mpkg}}
	for _, m := range work.pkgs {
		if ctx.Err() != nil {
			break
		}
		o.handleCandidate(ctx, m, work.height)
	}
	if ctx.Err() == nil {
		o.recordVerified(work.height)
	}

	assert.Equal(t, int64(4), o.state.lastVerifiedHeight(),
		"the block was abandoned, not verified: recording it strands every package it carried")
	assert.NotEqual(t, statusApproved, o.status.get(path).Status,
		"nothing was approved, so the cursor has no business claiming the block is done")
}

// catchingUpRPC is a node that is replaying or fast-syncing: RPC is up, and the
// height it reports is its own progress rather than the chain's.
type catchingUpRPC struct {
	rpcclient.Client
	tip int64
}

func (r catchingUpRPC) Status(context.Context, *int64) (*ctypes.ResultStatus, error) {
	return &ctypes.ResultStatus{
		SyncInfo: ctypes.SyncInfo{LatestBlockHeight: r.tip, CatchingUp: true},
	}, nil
}

// TestACatchingUpNodeIsNotAnAnswer: a node that has not finished syncing must
// not be read as a chain that was reset.
//
// startHeight refuses a cursor above the tip, and that refusal is fatal -- run()
// returns the error and the process exits. A replaying node serves RPC with a
// height climbing from wherever it restarted, so gpao brought up alongside its
// node (the systemd case this PR is for) would read a low tip, decide the chain
// had been reset under the same id, and exit. With Restart=always that is a
// crash loop until the node passes the cursor; StartLimitBurst can end it for
// good.
//
// Same class TestRunSurvivesTheBootRace pins for a failing Status call: one
// startup RPC must not be fatal. So the tip is treated as unsettled and the
// caller polls.
func TestACatchingUpNodeIsNotAnAnswer(t *testing.T) {
	o := newStubOracle(t, catchingUpRPC{tip: 12})
	require.NoError(t, o.state.setLastVerifiedHeight(26021))

	h, answered, err := o.startHeight(context.Background())
	require.NoError(t, err,
		"a node that is still syncing has said nothing about the chain's tip; "+
			"exiting on it re-opens the boot race")
	assert.False(t, answered, "unsettled, so the caller has to ask again")
	assert.Zero(t, h)

	// And the refusal still works once the node has actually caught up.
	synced := newCursorOracle(t, 12)
	require.NoError(t, synced.state.setLastVerifiedHeight(26021))
	_, _, err = synced.startHeight(context.Background())
	require.Error(t, err,
		"a synced node whose tip is below the cursor is the real reset case")
	assert.Contains(t, err.Error(), "recorded cursor")
}
