package main

import (
	"bytes"
	"context"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStrandedImporterWaitsWhileItsImportIsParked: an import still parked
// would leave the package pending again, so it is not offered. Its submitting
// block is read once, to learn which import holds it back, and from then on a
// pass asks the node about that import alone.
func TestStrandedImporterWaitsWhileItsImportIsParked(t *testing.T) {
	chain := newInertChain(t)
	dep, app := depAndImporter()
	chain.park(t, dep)
	chain.park(t, app)

	var blockReads, metaQueries atomic.Int64
	blocks := httptest.NewServer(countingBody(t, chain.remote, `"method":"block"`, &blockReads))
	t.Cleanup(blocks.Close)
	metas := httptest.NewServer(countingQuery(t, blocks.URL, "vm/qpkgmeta", &metaQueries))
	t.Cleanup(metas.Close)
	chain.remote = metas.URL

	o := chain.oracle(t)
	o.handleCandidate(t.Context(), candidate{mpkg: app, height: chain.submittedAt(t, app.Path)})
	o.reofferStranded(t.Context())
	metaQueries.Store(0)
	o.reofferStranded(t.Context())

	assert.Equal(t, statusPending, o.status.get(app.Path).Status)
	assert.Equal(t, dep.Path, o.stranded[app.Path].blockedOn)
	assert.Equal(t, int64(1), blockReads.Load(), "the block is read again only once the import goes live")
	assert.Equal(t, int64(1), metaQueries.Load(), "a pass asks about the parked import alone")
}

// TestCachedGenesisImportIsLiveWithoutAsking: the import cache holds only
// public genesis packages, live for the life of the chain, so one it holds is
// not asked about.
func TestCachedGenesisImportIsLiveWithoutAsking(t *testing.T) {
	const dep = "gno.land/p/nt/ufmt/v0"
	var metaQueries atomic.Int64
	counted := httptest.NewServer(countingQuery(t, nodeWithGenesisPackage(t, dep), "vm/qpkgmeta", &metaQueries))
	t.Cleanup(counted.Close)

	o := newTestOracleAt(t, counted.URL)
	o.cfg.verifyBudget = time.Minute
	require.NoError(t, o.verify(context.Background(), packageImporting(dep)))
	require.Contains(t, o.genesisImports(), dep, "premise: the verification cached the import")
	metaQueries.Store(0)

	assert.Empty(t, o.firstImportHoldingBack(t.Context(), packageImporting(dep), map[string]string{}))
	assert.Zero(t, metaQueries.Load())
}

// TestReoffersOfOneSubmissionAreCapped: a submission that keeps coming back
// pending is verified again at most maxReoffers times in a run.
func TestReoffersOfOneSubmissionAreCapped(t *testing.T) {
	chain := newInertChain(t)
	const path = "gno.land/r/test/slow"
	pkg := chainPackage(path, "package slow\n\nfunc N(cur realm) int { return 1 }\n")
	chain.park(t, pkg)

	o := chain.oracle(t)
	o.cfg.prepareBudget = time.Nanosecond // every verification ends unavailable
	o.handleCandidate(t.Context(), candidate{mpkg: pkg, height: chain.submittedAt(t, path)})
	require.Equal(t, statusPending, o.status.get(path).Status, "premise: the submission is stranded")

	for n := 1; n <= maxReoffers; n++ {
		o.reofferStranded(t.Context())
		require.Equal(t, n, o.stranded[path].offers)
	}
	o.reofferStranded(t.Context())
	assert.NotContains(t, o.stranded, path)
	assert.Equal(t, statusGaveUp, o.status.get(path).Status,
		"nothing offers it again this run, so the board must not say it will be")
}

// TestStrandedReasonNamesTheParkedImport: a submission whose first look could
// not finish, and which a later pass finds held back by a parked import, says
// so on the status board.
func TestStrandedReasonNamesTheParkedImport(t *testing.T) {
	chain := newInertChain(t)
	dep, app := depAndImporter()
	chain.park(t, dep)
	chain.park(t, app)

	o := chain.oracle(t)
	o.cfg.prepareBudget = time.Nanosecond // the first look cannot finish
	o.handleCandidate(t.Context(), candidate{mpkg: app, height: chain.submittedAt(t, app.Path)})
	o.cfg.prepareBudget = defaultPrepareBudget
	o.reofferStranded(t.Context())

	require.Equal(t, dep.Path, o.stranded[app.Path].blockedOn, "premise: the pass found the parked import")
	st := o.status.get(app.Path)
	assert.Equal(t, statusPending, st.Status)
	assert.Contains(t, st.Reason, dep.Path)
}

// TestStrandedImporterOfAnAbsentImportIsRejected: an import the chain holds
// nowhere is a rejection whatever else is parked, so a pass offers its
// importer, whose verification rejects it, instead of waiting on the import.
func TestStrandedImporterOfAnAbsentImportIsRejected(t *testing.T) {
	chain := newInertChain(t)
	const path = "gno.land/r/test/wantsmissing"
	pkg := chainPackage(path, "package wantsmissing\n\nimport \"gno.land/p/test/neversubmitted\"\n\n"+
		"func N(cur realm) int { return neversubmitted.Answer() }\n")
	chain.park(t, pkg)

	o := chain.oracle(t)
	o.cfg.prepareBudget = time.Nanosecond // the first look cannot finish
	o.handleCandidate(t.Context(), candidate{mpkg: pkg, height: chain.submittedAt(t, path)})
	require.Contains(t, o.stranded, path, "premise: the submission is stranded")
	o.cfg.prepareBudget = defaultPrepareBudget
	o.reofferStranded(t.Context())

	st := o.status.get(path)
	assert.Equal(t, statusRejected, st.Status, "recorded reason: %s", st.Reason)
	assert.NotContains(t, o.stranded, path)
}

// TestLaterStrandedPathIsReachedUnderTraffic: a pass stops for every queued
// candidate, so on a chain that keeps receiving them each pass is cut short.
// The next one resumes where it stopped, or a path late in order would wait
// behind every earlier one for the whole run.
func TestLaterStrandedPathIsReachedUnderTraffic(t *testing.T) {
	chain := newInertChain(t)
	early := chainPackage("gno.land/r/test/aaa", "package aaa\n\nfunc N(cur realm) int { return 1 }\n")
	late := chainPackage("gno.land/r/test/zzz", "package zzz\n\nfunc N(cur realm) int { return 1 }\n")
	chain.park(t, early)
	chain.park(t, late)
	earlyAt, lateAt := chain.submittedAt(t, early.Path), chain.submittedAt(t, late.Path)

	// A new candidate arrives while a pass reads a submission's block.
	var o *oracle
	var traffic atomic.Bool
	proxy := httptest.NewServer(rewriting(t, chain.remote, func(body []byte) []byte {
		if traffic.Load() && bytes.Contains(body, []byte(`"method":"block"`)) {
			select {
			case o.candidates <- candidate{mpkg: early, height: earlyAt}:
			default:
			}
		}
		return body
	}))
	t.Cleanup(proxy.Close)
	chain.remote = proxy.URL

	o = chain.oracle(t)
	o.cfg.prepareBudget = time.Nanosecond // every verification comes back pending
	o.handleCandidate(t.Context(), candidate{mpkg: early, height: earlyAt})
	o.handleCandidate(t.Context(), candidate{mpkg: late, height: lateAt})
	traffic.Store(true)

	for range 2 {
		o.reofferStranded(t.Context())
		for len(o.candidates) > 0 {
			<-o.candidates // the verifier takes the new block's candidate
		}
	}
	assert.Equal(t, 1, o.stranded[late.Path].offers, "the second pass resumes at the path the first stopped at")
}

// TestCutPassResumesOnceTheQueueDrains: a pass cut short for a new candidate
// goes on as soon as the verifier has handled it, not a whole interval later.
func TestCutPassResumesOnceTheQueueDrains(t *testing.T) {
	chain := newInertChain(t)
	early := chainPackage("gno.land/r/test/aaa", "package aaa\n\nfunc N(cur realm) int { return 1 }\n")
	late := chainPackage("gno.land/r/test/zzz", "package zzz\n\nfunc N(cur realm) int { return 1 }\n")
	chain.park(t, early)
	chain.park(t, late)
	earlyAt, lateAt := chain.submittedAt(t, early.Path), chain.submittedAt(t, late.Path)

	// One new candidate arrives while the first pass reads its first block.
	var o *oracle
	var blockReads atomic.Int64
	proxy := httptest.NewServer(rewriting(t, chain.remote, func(body []byte) []byte {
		if bytes.Contains(body, []byte(`"method":"block"`)) && blockReads.Add(1) == 1 {
			o.candidates <- candidate{mpkg: early, height: earlyAt}
		}
		return body
	}))
	t.Cleanup(proxy.Close)
	chain.remote = proxy.URL

	o = chain.oracle(t)
	o.cfg.prepareBudget = time.Nanosecond
	o.handleCandidate(t.Context(), candidate{mpkg: early, height: earlyAt})
	o.handleCandidate(t.Context(), candidate{mpkg: late, height: lateAt})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	tick := make(chan time.Time)
	stopped := make(chan struct{})
	go func() {
		o.runVerifier(ctx, tick)
		close(stopped)
	}()
	tick <- time.Now() // the only tick

	require.Eventually(t, func() bool { return blockReads.Load() >= 2 }, 10*time.Second, 50*time.Millisecond,
		"the cut pass must read the later submission's block without waiting for another tick")
	cancel()
	<-stopped
	assert.Positive(t, o.stranded[late.Path].offers)
}

// TestStrandedSubmissionLeavesWhenTheChainMovesOn: once the chain no longer
// holds the submission parked there is nothing to offer, whoever enabled it.
func TestStrandedSubmissionLeavesWhenTheChainMovesOn(t *testing.T) {
	chain := newInertChain(t)
	const path = "gno.land/r/test/elsewhere"
	pkg := chainPackage(path, "package elsewhere\n\nfunc N(cur realm) int { return 1 }\n")
	chain.park(t, pkg)
	height := chain.submittedAt(t, path)

	cut := chain.oracle(t)
	cut.cfg.prepareBudget = time.Nanosecond
	cut.handleCandidate(t.Context(), candidate{mpkg: pkg, height: height})
	require.Contains(t, cut.stranded, path, "premise: the submission is stranded")
	other := chain.oracle(t)
	other.handleCandidate(t.Context(), candidate{mpkg: pkg, height: height})
	require.Equal(t, statusApproved, other.status.get(path).Status, "premise: another approver enabled it")

	cut.reofferStranded(t.Context())

	assert.NotContains(t, cut.stranded, path)
	assert.Equal(t, statusPending, cut.status.get(path).Status, "nothing was verified again")
}

// TestVerifierReoffersOnItsTick: a package left pending on a parked import is
// verified again once that import goes live, without a restart or a
// resubmission. Heights are read once, so its block is never read again
// otherwise, and one stuck dependency would strand every package above it.
// The verifier does it between candidates, on its own tick, so the block
// reader never waits on it.
func TestVerifierReoffersOnItsTick(t *testing.T) {
	chain := newInertChain(t)
	dep, app := depAndImporter()
	chain.park(t, dep)
	chain.park(t, app)

	o := chain.oracle(t)
	o.handleCandidate(t.Context(), candidate{mpkg: app, height: chain.submittedAt(t, app.Path)})
	require.Equal(t, statusPending, o.status.get(app.Path).Status, "premise: the importer is stranded")
	o.handleCandidate(t.Context(), candidate{mpkg: dep, height: chain.submittedAt(t, dep.Path)})
	require.Equal(t, statusApproved, o.status.get(dep.Path).Status, "premise: the import is live")

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	tick := make(chan time.Time)
	stopped := make(chan struct{})
	go func() {
		o.runVerifier(ctx, tick)
		close(stopped)
	}()
	tick <- time.Now()

	require.Eventually(t, func() bool { return o.status.get(app.Path).Status == statusApproved },
		30*time.Second, 50*time.Millisecond)
	cancel()
	<-stopped
	assert.NotContains(t, o.stranded, app.Path, "a verdict ends the tracking")
}

// depAndImporter returns a package and a realm whose only chain import is it.
func depAndImporter() (dep, app *std.MemPackage) {
	const depPath = "gno.land/p/test/parkeddep"
	dep = chainPackage(depPath, "package parkeddep\n\nfunc Answer() int { return 42 }\n")
	app = chainPackage("gno.land/r/test/wantsdep",
		"package wantsdep\n\nimport \""+depPath+"\"\n\nfunc N(cur realm) int { return parkeddep.Answer() }\n")
	return dep, app
}

// submittedAt is the height the chain stamped on the submission at path.
func (c *inertChain) submittedAt(t *testing.T, path string) int64 {
	t.Helper()
	meta, err := queryPkgMeta(t.Context(), c.client.RPCClient, path)
	require.NoError(t, err)
	require.Positive(t, meta.Height)
	return int64(meta.Height)
}
