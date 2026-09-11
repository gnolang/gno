package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// answersPkgMeta is a nil rpcclient.Client that answers the startup probe for
// vm/qpkgmeta_json, which every stub that reaches the work loops has to pass.
// Any other query, and every other method, still panics, as in stubRPC.
type answersPkgMeta struct {
	rpcclient.Client
}

func (answersPkgMeta) ABCIQuery(_ context.Context, path string, _ []byte) (*ctypes.ResultABCIQuery, error) {
	if path != "vm/qpkgmeta_json" {
		panic("unexpected ABCIQuery path: " + path)
	}
	return &ctypes.ResultABCIQuery{Response: abci.ResponseQuery{ResponseBase: abci.ResponseBase{
		Data: []byte(`{"path":"gno.land/p/gpao/probe","status":"absent"}`),
	}}}, nil
}

// bootRaceRPC is a node that is not up yet: the first `failures` Status calls
// error, later ones answer with a tip that advances a block per poll, the way a
// live chain's does. Asking for a block proves the oracle survived the boot
// race, so the stub ends the test by cancelling then. The embedded interface is
// nil on purpose, as in stubRPC: an unexpected call panics rather than passing
// silently.
type bootRaceRPC struct {
	answersPkgMeta
	mu         sync.Mutex
	failures   int
	tip        int64
	blockAsked *int64 // first height asked of Block, nil until then
	cancel     context.CancelFunc
}

func (s *bootRaceRPC) Status(context.Context, *int64) (*ctypes.ResultStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failures > 0 {
		s.failures--
		return nil, errors.New("connection refused: the node is still booting")
	}
	res := &ctypes.ResultStatus{}
	res.SyncInfo.LatestBlockHeight = s.tip
	// The chain keeps producing while the oracle follows it. A tip that never
	// moved would leave the oracle waiting one block ahead of it forever, and
	// no Block would ever be asked for.
	s.tip++
	return res, nil
}

// The ceiling gates the work loops, so a node that never answers this is never
// followed at all. This one answers immediately: the boot race under test is
// the tip query's, not the ceiling's.
func (s *bootRaceRPC) ConsensusParams(context.Context, *int64) (*ctypes.ResultConsensusParams, error) {
	return &ctypes.ResultConsensusParams{
		ConsensusParams: abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 500_000}},
	}, nil
}

func (s *bootRaceRPC) Block(_ context.Context, height *int64) (*ctypes.ResultBlock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.blockAsked == nil {
		h := *height
		s.blockAsked = &h
		s.cancel() // the point is proven; end the run
	}
	return blockWith(), nil
}

// TestRunSurvivesTheBootRace pins that -start-height 0 does not make one
// startup RPC fatal.
//
// With no -start-height, run() asks the node for its tip to decide where to
// begin. That call used to be fatal on error while the IDENTICAL Status call
// inside the polling loop is logged and retried -- same endpoint, same query,
// opposite outcomes on one flag. A supervisor that brings gpao up alongside
// the chain rather than after it hit this on every boot, and the status
// listener comes up before the fatal query, so a readiness probe could pass
// on a process that was already dying.
func TestRunSurvivesTheBootRace(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rpc := &bootRaceRPC{failures: 2, tip: 7, cancel: cancel}
	o := newStubOracle(rpc)
	o.cfg.pollInterval = time.Millisecond
	o.cfg.startHeight = 0 // the flag under test: begin from the tip

	err := o.run(ctx)
	require.NoError(t, err,
		"a startup tip-query failure must be retried like the in-loop one, not end the process")

	require.NotNil(t, rpc.blockAsked, "the oracle never started following blocks")
	require.Equal(t, int64(8), *rpc.blockAsked,
		"the tip at the first answered poll decides where to begin: LatestBlockHeight+1, exactly as -start-height 0 documents")
}

// explicitStartRPC is a node that is already up and settled at `tip`. Asking
// for a block ends the run, so the height recorded is the one the oracle chose
// to begin at. The embedded interface is nil on purpose, as in stubRPC.
type explicitStartRPC struct {
	answersPkgMeta
	mu         sync.Mutex
	tip        int64
	blockAsked *int64 // first height asked of Block, nil until then
	cancel     context.CancelFunc
}

func (s *explicitStartRPC) Status(context.Context, *int64) (*ctypes.ResultStatus, error) {
	res := &ctypes.ResultStatus{}
	res.SyncInfo.LatestBlockHeight = s.tip
	return res, nil
}

func (s *explicitStartRPC) ConsensusParams(context.Context, *int64) (*ctypes.ResultConsensusParams, error) {
	return &ctypes.ResultConsensusParams{
		ConsensusParams: abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 500_000}},
	}, nil
}

func (s *explicitStartRPC) Block(_ context.Context, height *int64) (*ctypes.ResultBlock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.blockAsked == nil {
		h := *height
		s.blockAsked = &h
		s.cancel() // the first height asked is the whole question; end the run
	}
	return blockWith(), nil
}

// TestRunHonoursExplicitStartHeight pins that resolving the tip does not
// swallow -start-height.
//
// Only an unset height is resolved, so a configured one reaches the block
// reader untouched and is the first block read. Resolving whenever the height
// lags the tip would start at the tip instead, and the history the operator
// asked for would go unread with nothing said about it.
func TestRunHonoursExplicitStartHeight(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rpc := &explicitStartRPC{tip: 7, cancel: cancel}
	o := newStubOracle(rpc)
	o.cfg.pollInterval = time.Millisecond
	o.cfg.startHeight = 5 // behind the tip: there is history to catch up on

	require.NoError(t, o.run(ctx))

	require.NotNil(t, rpc.blockAsked, "the oracle never started following blocks")
	require.Equal(t, int64(5), *rpc.blockAsked,
		"-start-height names the first block to read; the tip only says how far to catch up")
}

// ceilingLateRPC is a real node that will not say what its ceiling is yet: the
// first `failures` ConsensusParams calls error, later ones are served by the
// node behind it. Answering ends the run, because what the oracle settles on is
// the whole question. Only the block reader asks for the ceiling, so `failures`
// needs no lock.
type ceilingLateRPC struct {
	rpcclient.Client
	failures int
	cancel   context.CancelFunc
}

func (c *ceilingLateRPC) ConsensusParams(ctx context.Context, height *int64) (*ctypes.ResultConsensusParams, error) {
	if c.failures > 0 {
		c.failures--
		return nil, errors.New("connection refused: the node is still booting")
	}
	res, err := c.Client.ConsensusParams(ctx, height)
	c.cancel() // the chain has answered; end the run
	return res, err
}

// TestRunAdoptsTheCeilingOnceTheChainAnswers pins that a startup window where
// the node will not answer does not leave the oracle on the fallback ceiling
// for the rest of the process.
//
// The ceiling gates the work loops, so a startup window where the node will
// not answer delays the first approval rather than settling the ceiling. What
// this pins is the value that ends up held: the chain's own, not the stand-in.
// On a chain configured BELOW the stand-in, as here, the ante REFUSES a probe
// signed above Block.MaxGas rather than clamping it, so holding the wrong one
// costs every approval.
//
// Against a real node because the claim is about the chain's own number -- that
// a chain configured below the fallback reports it through ConsensusParams, and
// that the oracle ends up holding that value rather than its own guess.
func TestRunAdoptsTheCeilingOnceTheChainAnswers(t *testing.T) {
	const chainMaxGas = int64(500_000) // deliberately below the stand-in ceiling

	cfg := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	cfg.Genesis.ConsensusParams.Block.MaxGas = chainMaxGas

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	defer node.Stop()

	nodeRPC, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	o := newStubOracle(&ceilingLateRPC{Client: nodeRPC, failures: 2, cancel: cancel})
	o.cfg.pollInterval = time.Millisecond

	require.NoError(t, o.run(ctx))

	require.Equal(t, chainMaxGas, o.blockMaxGas,
		"the ceiling must be asked for until the chain answers; left on the fallback, every probe is signed above Block.MaxGas and the ante refuses them all")
}

// ceilingFirstRPC is a node that will not answer ConsensusParams for the first
// `failures` calls, and records any block read that arrives before it does.
// Reading a block is the work the ceiling gates; pinning the start height is
// not, so Status answers throughout and its tip advances a block per call.
// The first block read ends the run. The embedded interface is nil on purpose,
// as in stubRPC.
type ceilingFirstRPC struct {
	answersPkgMeta
	mu           sync.Mutex
	failures     int
	tip          int64
	answered     bool
	readTooEarly []int64 // heights read while the ceiling was still unknown
	cancel       context.CancelFunc
}

func (c *ceilingFirstRPC) ConsensusParams(context.Context, *int64) (*ctypes.ResultConsensusParams, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failures > 0 {
		c.failures--
		return nil, errors.New("connection refused: the node is still booting")
	}
	c.answered = true
	return &ctypes.ResultConsensusParams{
		ConsensusParams: abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 500_000}},
	}, nil
}

func (c *ceilingFirstRPC) Status(context.Context, *int64) (*ctypes.ResultStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := &ctypes.ResultStatus{}
	res.SyncInfo.LatestBlockHeight = c.tip
	c.tip++
	return res, nil
}

func (c *ceilingFirstRPC) Block(_ context.Context, height *int64) (*ctypes.ResultBlock, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.answered {
		c.readTooEarly = append(c.readTooEarly, *height)
	}
	c.cancel() // a block is being read, so the work loop is running; end the run
	return blockWith(), nil
}

// TestRunWaitsForTheCeilingBeforeWorking pins that the chain's gas ceiling is
// settled before any block is read or verified.
//
// Adopting the ceiling once the chain answers is not enough on its own: a
// candidate reached in the meantime is signed against the fallback, and on a
// chain configured below it the ante refuses the probe rather than clamping it,
// so the enable is graded as a message the node ran and rejected. Approvals in
// that window fail for a reason that has nothing to do with the package.
//
// The ceiling is what every approval's fee is sized against, so it is a
// prerequisite of the work loops rather than one more thing the poll retries.
func TestRunWaitsForTheCeilingBeforeWorking(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rpc := &ceilingFirstRPC{failures: 3, tip: 1, cancel: cancel}
	o := newStubOracle(rpc)
	o.cfg.pollInterval = time.Millisecond

	require.NoError(t, o.run(ctx))

	require.Empty(t, rpc.readTooEarly,
		"no block may be read or verified while the ceiling is still the fallback: every approval in that window is signed against a number the chain did not give")
	require.True(t, rpc.answered, "the ceiling was never asked for until it answered")
	require.Equal(t, int64(500_000), o.blockMaxGas,
		"the chain's own ceiling must be the one held once it has answered")
}

// startupTipRPC is a node that is up and answering. It records that the tip was
// asked for, and answering ends the run. The embedded interface is nil on
// purpose, as in stubRPC.
type startupTipRPC struct {
	answersPkgMeta
	mu     sync.Mutex
	asked  bool
	cancel context.CancelFunc
}

// The ceiling answers at once: what this stub is about is the tip query, which
// is asked before it.
func (s *startupTipRPC) ConsensusParams(context.Context, *int64) (*ctypes.ResultConsensusParams, error) {
	return &ctypes.ResultConsensusParams{
		ConsensusParams: abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 500_000}},
	}, nil
}

func (s *startupTipRPC) Status(context.Context, *int64) (*ctypes.ResultStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.asked = true
	s.cancel() // the tip has been asked for, which is the whole question
	res := &ctypes.ResultStatus{}
	res.SyncInfo.LatestBlockHeight = 41
	return res, nil
}

func (s *startupTipRPC) tipAsked() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.asked
}

// TestRunResolvesTheTipBeforeTheFirstPollInterval pins that -start-height 0
// begins at the tip as of startup rather than the tip a poll interval later.
//
// Heights only move forward from wherever the first answered poll lands, so a
// tip resolved one interval late means every block committed in that interval
// is never read and every package submitted in it is never seen. Nothing is
// logged about them: the oracle reports itself healthy and approves nothing
// that arrived while it was starting. A supervisor that restarts gpao while
// submissions are in flight loses exactly that window.
//
// The poll interval here is far longer than the wait below, so a tick cannot
// account for the query: the only way to ask this soon is to ask before
// waiting for one.
func TestRunResolvesTheTipBeforeTheFirstPollInterval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rpc := &startupTipRPC{cancel: cancel}
	o := newStubOracle(rpc)
	o.cfg.pollInterval = time.Minute
	o.cfg.startHeight = 0 // the flag under test: begin from the tip

	errc := make(chan error, 1)
	go func() { errc <- o.run(ctx) }()

	require.Eventually(t, rpc.tipAsked, 5*time.Second, 10*time.Millisecond,
		"the tip must be asked for at startup; every block committed before the first poll is one the oracle never reads")
	require.NoError(t, <-errc)
}

// unusableTipRPC is a node that answers the tip query with a height no chain
// has. It counts the calls, because what a nonsense answer must not do is turn
// the retry into a spin. The embedded interface is nil on purpose, as in
// stubRPC.
type unusableTipRPC struct {
	rpcclient.Client
	mu           sync.Mutex
	calls        int
	ceilingAsked bool
}

// The ceiling is asked for only once the start height is settled, so a call
// here is the oracle having moved on with a height it cannot use.
func (s *unusableTipRPC) ConsensusParams(context.Context, *int64) (*ctypes.ResultConsensusParams, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ceilingAsked = true
	return &ctypes.ResultConsensusParams{
		ConsensusParams: abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 500_000}},
	}, nil
}

func (s *unusableTipRPC) Status(context.Context, *int64) (*ctypes.ResultStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	res := &ctypes.ResultStatus{}
	res.SyncInfo.LatestBlockHeight = -1
	return res, nil
}

func (s *unusableTipRPC) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func (s *unusableTipRPC) ceilingWasAsked() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ceilingAsked
}

// TestRunPacesAnUnusableTipAnswer pins that a height no chain has leaves the
// start height unresolved, asked again on the poll interval, interruptible
// throughout.
//
// Accepting such a height is worse than getting no answer. One less than the
// first real block is 0, and no chain has a block 0: the node refuses every
// request for it, and the work loop does not advance past a height it could
// not read, so the daemon reports the same refusal every interval and approves
// nothing for the life of the process.
//
// One tick cannot arrive inside the window below, so the count also prices the
// pacing: asked once, then waiting.
func TestRunPacesAnUnusableTipAnswer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rpc := &unusableTipRPC{}
	o := newStubOracle(rpc)
	o.cfg.pollInterval = time.Minute
	o.cfg.startHeight = 0

	errc := make(chan error, 1)
	go func() { errc <- o.run(ctx) }()

	require.Eventually(t, func() bool { return rpc.callCount() > 0 }, 5*time.Second, 10*time.Millisecond,
		"the tip must be asked for at startup")
	require.False(t, rpc.ceilingWasAsked(),
		"a height no chain has must not settle the start height: starting at 0 stalls the run on a block the node refuses, every interval, forever")
	require.Equal(t, 1, rpc.callCount(),
		"an unresolved start height waits for the next tick before asking again")

	cancel()
	require.NoError(t, <-errc, "the wait must stay interruptible while the start height is unresolved")
}

// ceilingSlowRPC is a node whose ceiling is unavailable for the first
// `failures` calls while its chain keeps committing: every RPC that answers or
// fails advances the tip, the way a real chain does not stop for a daemon
// waiting on it. The first height asked of Block says which tip the oracle
// anchored to, so asking ends the run. The embedded interface is nil on
// purpose, as in stubRPC.
type ceilingSlowRPC struct {
	answersPkgMeta
	mu         sync.Mutex
	failures   int
	tip        int64
	blockAsked *int64 // first height asked of Block, nil until then
	cancel     context.CancelFunc
}

func (s *ceilingSlowRPC) ConsensusParams(context.Context, *int64) (*ctypes.ResultConsensusParams, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failures > 0 {
		s.failures--
		s.tip++ // the chain committed another block while it could not answer
		return nil, errors.New("connection refused: the node is still booting")
	}
	return &ctypes.ResultConsensusParams{
		ConsensusParams: abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 500_000}},
	}, nil
}

func (s *ceilingSlowRPC) Status(context.Context, *int64) (*ctypes.ResultStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := &ctypes.ResultStatus{}
	res.SyncInfo.LatestBlockHeight = s.tip
	s.tip++
	return res, nil
}

func (s *ceilingSlowRPC) Block(_ context.Context, height *int64) (*ctypes.ResultBlock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.blockAsked == nil {
		h := *height
		s.blockAsked = &h
		s.cancel() // the first height asked is the whole question; end the run
	}
	return blockWith(), nil
}

// TestRunPinsTheTipBeforeWaitingOnTheCeiling pins that a ceiling the chain
// cannot report yet does not cost the blocks committed while it is asked.
//
// The ceiling gates verification, because a probe signed against a stand-in is
// refused by the ante rather than clamped. It does not gate reading the tip:
// pinning a height approves nothing. Asked in the other order, every interval
// spent retrying the ceiling is an interval whose blocks the oracle anchors
// past and never reads, and the log cannot even name the range, because the
// starting tip was never learned.
//
// Here the tip is 5 at startup and the chain reaches 9 while the ceiling is
// unavailable, so the run must begin at 6 and catch up the rest.
func TestRunPinsTheTipBeforeWaitingOnTheCeiling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rpc := &ceilingSlowRPC{failures: 3, tip: 5, cancel: cancel}
	o := newStubOracle(rpc)
	o.cfg.pollInterval = time.Millisecond
	o.cfg.startHeight = 0 // the flag under test: begin from the tip

	require.NoError(t, o.run(ctx))

	require.NotNil(t, rpc.blockAsked, "the oracle never started following blocks")
	require.Equal(t, int64(6), *rpc.blockAsked,
		"the tip must be pinned before the ceiling is waited on; anchoring after it skips every block committed during the wait")
}

// renamingRoute proxies to remote, rewriting one ABCI query path in every
// request so the node answers with its own unknown-request error: a node
// built without the route, as an older release is.
func renamingRoute(t *testing.T, remote, from, to string) http.Handler {
	t.Helper()
	proxy := nodeProxy(t, remote)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		body = bytes.ReplaceAll(body, []byte(`"`+from+`"`), []byte(`"`+to+`"`))
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		proxy.ServeHTTP(w, r)
	})
}

// blockEndsRun is a real node whose first block read ends the run. Reading a
// block is what the startup checks gate, so reaching one proves they passed.
type blockEndsRun struct {
	rpcclient.Client
	cancel context.CancelFunc
}

func (c *blockEndsRun) Block(ctx context.Context, height *int64) (*ctypes.ResultBlock, error) {
	c.cancel()
	return c.Client.Block(ctx, height)
}

// TestRunRefusesANodeWithoutThePackageMetaRoute pins that the daemon does not
// start against a node that cannot answer vm/qpkgmeta_json.
//
// The verifier asks that route about every import vm/qfile would not serve;
// it is what tells a parked import from an absent one. A node without it
// answers "unknown request", which the child reads as the node describing
// itself rather than the path, so every package importing an absent path is
// left pending and uncounted where it should be rejected, and nothing in the
// log names the route. Refusing at startup is the one place the mismatch is
// visible.
func TestRunRefusesANodeWithoutThePackageMetaRoute(t *testing.T) {
	cfg := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	defer node.Stop()

	t.Run("a node without the route is refused", func(t *testing.T) {
		gone := httptest.NewServer(renamingRoute(t, remote, "vm/qpkgmeta_json", "vm/qpkgmeta_gone"))
		t.Cleanup(gone.Close)
		rpc, err := rpcclient.NewHTTPClient(gone.URL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		o := newStubOracle(rpc)
		o.cfg.pollInterval = time.Millisecond
		o.cfg.startHeight = 1

		err = o.run(ctx)
		require.Error(t, err, "a node that cannot classify an unresolved import must not be followed")
		assert.ErrorContains(t, err, "vm/qpkgmeta_json", "the refusal must name the route")
	})

	t.Run("a node answering the route in a form gpao cannot read is refused", func(t *testing.T) {
		// Renamed to vm/qpaths, the probe draws a successful answer that is not
		// a package's metadata, so every classification would fail the same way.
		garbled := httptest.NewServer(renamingRoute(t, remote, "vm/qpkgmeta_json", "vm/qpaths"))
		t.Cleanup(garbled.Close)
		rpc, err := rpcclient.NewHTTPClient(garbled.URL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		o := newStubOracle(rpc)
		o.cfg.pollInterval = time.Millisecond
		o.cfg.startHeight = 1

		err = o.run(ctx)
		require.Error(t, err, "a node whose answers the verifier cannot read must not be followed")
		assert.ErrorContains(t, err, "vm/qpkgmeta_json", "the refusal must name the route")
	})

	t.Run("a node answering with another error is asked again", func(t *testing.T) {
		// Rewritten to vm/qfile, the first two probes draw the node's "package
		// not available": an answered error that says nothing about the route.
		var probes atomic.Int32
		proxy := nodeProxy(t, remote)
		flaky := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			if bytes.Contains(body, []byte(`"vm/qpkgmeta_json"`)) && probes.Add(1) <= 2 {
				body = bytes.ReplaceAll(body, []byte(`"vm/qpkgmeta_json"`), []byte(`"vm/qfile"`))
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
			proxy.ServeHTTP(w, r)
		}))
		t.Cleanup(flaky.Close)
		rpc, err := rpcclient.NewHTTPClient(flaky.URL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		o := newStubOracle(&blockEndsRun{Client: rpc, cancel: cancel})
		o.cfg.pollInterval = time.Millisecond
		o.cfg.startHeight = 1

		require.NoError(t, o.run(ctx), "only a missing route may refuse the node")
		require.ErrorIs(t, ctx.Err(), context.Canceled,
			"the run must have ended on the block read, which cancels, not on the deadline")
		assert.Greater(t, probes.Load(), int32(2), "premise: the probe was refused before it was answered")
	})

	t.Run("a node with the route is followed", func(t *testing.T) {
		rpc, err := rpcclient.NewHTTPClient(remote)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		o := newStubOracle(&blockEndsRun{Client: rpc, cancel: cancel})
		o.cfg.pollInterval = time.Millisecond
		o.cfg.startHeight = 1

		require.NoError(t, o.run(ctx))
		require.ErrorIs(t, ctx.Err(), context.Canceled,
			"the run must have ended on the block read, which cancels, not on the deadline")
	})
}
