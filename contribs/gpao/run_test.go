package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// bootRaceRPC is a node that is not up yet: the first `failures` Status calls
// error, later ones answer with a tip that advances a block per poll, the way a
// live chain's does. Asking for a block proves the oracle survived the boot
// race, so the stub ends the test by cancelling then. The embedded interface is
// nil on purpose, as in stubRPC: an unexpected call panics rather than passing
// silently.
type bootRaceRPC struct {
	rpcclient.Client
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

func (s *bootRaceRPC) ConsensusParams(context.Context, *int64) (*ctypes.ResultConsensusParams, error) {
	return nil, errors.New("still booting") // blockMaxGasFrom falls back
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
	rpcclient.Client
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
	return nil, errors.New("no consensus params here") // blockMaxGasFrom falls back
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

// TestRunHonoursExplicitStartHeight pins that deferring the tip resolution did
// not swallow -start-height.
//
// The guard that resolves the tip reads the loop-carried height, so a
// configured start passes through it untouched and is the first block read.
// A guard that also fired whenever the height lags the tip
// (height <= 0 || height < latest) would start at the tip instead, and the
// history the operator asked for would go unread with nothing said about it.
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
// Block.MaxGas was read once, before the polling loop, and never asked for
// again. A node that was not up yet left the oracle at defaultBlockMaxGas --
// and on a chain configured BELOW that, as here, the ante REFUSES a probe
// signed above Block.MaxGas rather than clamping it. Every simulate then comes
// back as a message the node ran and rejected (verdictWillFail, not
// verdictUnknown), so enable returns without broadcasting: the oracle approves
// nothing, silently, until someone restarts it.
//
// Against a real node because the claim is about the chain's own number -- that
// a chain configured below the fallback reports it through ConsensusParams, and
// that the oracle ends up holding that value rather than its own guess.
func TestRunAdoptsTheCeilingOnceTheChainAnswers(t *testing.T) {
	const chainMaxGas = int64(500_000) // deliberately below defaultBlockMaxGas

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

	require.Equal(t, chainMaxGas, o.blockMaxGas.Load(),
		"the ceiling must be asked for until the chain answers; left on the fallback, every probe is signed above Block.MaxGas and the ante refuses them all")
}
