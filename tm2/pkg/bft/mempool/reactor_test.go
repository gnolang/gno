package mempool

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	memcfg "github.com/gnolang/gno/tm2/pkg/bft/mempool/config"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pcfg "github.com/gnolang/gno/tm2/pkg/p2p/config"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

// testP2PConfig returns a configuration for testing the peer-to-peer layer
func testP2PConfig() *p2pcfg.P2PConfig {
	cfg := p2pcfg.DefaultP2PConfig()
	cfg.ListenAddress = "tcp://0.0.0.0:26656"
	cfg.FlushThrottleTimeout = 10 * time.Millisecond

	return cfg
}

type peerState struct {
	height int64
}

func (ps peerState) GetHeight() int64 {
	return ps.height
}

// connect N mempool reactors through N switches
func makeAndConnectReactors(t *testing.T, mconfig *memcfg.MempoolConfig, pconfig *p2pcfg.P2PConfig, n int) []*Reactor {
	t.Helper()

	var (
		reactors = make([]*Reactor, n)
		logger   = log.NewNoopLogger()
		options  = make(map[int][]p2p.SwitchOption)
	)

	for i := range n {
		app := kvstore.NewKVStoreApplication()
		cc := proxy.NewLocalClientCreator(app)
		mempool, cleanup := newMempoolWithApp(cc)
		defer cleanup()

		reactor := NewReactor(mconfig, mempool) // so we dont start the consensus states
		reactor.SetLogger(logger.With("validator", i))

		options[i] = []p2p.SwitchOption{
			p2p.WithReactor("MEMPOOL", reactor),
		}

		reactors[i] = reactor
	}

	// "Simulate" the networking layer
	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	cfg := p2pTesting.TestingConfig{
		Count:         n,
		P2PCfg:        pconfig,
		SwitchOptions: options,
		Channels:      []byte{MempoolChannel},
	}

	p2pTesting.MakeConnectedPeers(t, ctx, cfg)

	return reactors
}

func waitForTxsOnReactors(
	t *testing.T,
	txs types.Txs,
	reactors []*Reactor,
) {
	t.Helper()

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	// Wait for the txs to propagate in all mempools
	var wg sync.WaitGroup

	for i, reactor := range reactors {
		wg.Add(1)

		go func(r *Reactor, reactorIndex int) {
			defer wg.Done()

			reapedTxs := waitForTxsOnReactor(t, ctx, len(txs), r)

			for i, tx := range txs {
				assert.Equalf(t, tx, reapedTxs[i],
					fmt.Sprintf(
						"txs at index %d on reactor %d don't match: %v vs %v",
						i, reactorIndex,
						tx,
						reapedTxs[i],
					),
				)
			}
		}(reactor, i)
	}

	wg.Wait()
}

func waitForTxsOnReactor(
	t *testing.T,
	ctx context.Context,
	expectedLength int,
	reactor *Reactor,
) types.Txs {
	t.Helper()

	var (
		mempool = reactor.mempool
		ticker  = time.NewTicker(100 * time.Millisecond)
	)

	for {
		select {
		case <-ctx.Done():
			t.Fatal("timed out waiting for txs")
		case <-ticker.C:
			if mempool.Size() < expectedLength {
				continue
			}

			return mempool.ReapMaxTxs(expectedLength)
		}
	}
}

// ensure no txs on reactor after some timeout
func ensureNoTxs(t *testing.T, reactor *Reactor, timeout time.Duration) {
	t.Helper()

	time.Sleep(timeout) // wait for the txs in all mempools
	assert.Zero(t, reactor.mempool.Size())
}

func TestReactorBroadcastTxMessage(t *testing.T) {
	t.Parallel()

	mconfig := memcfg.TestMempoolConfig()
	pconfig := testP2PConfig()
	const N = 4
	reactors := makeAndConnectReactors(t, mconfig, pconfig, N)
	t.Cleanup(func() {
		for _, r := range reactors {
			assert.NoError(t, r.Stop())
		}
	})

	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().List() {
			fmt.Printf("Setting peer %s\n", peer.ID())
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	// send a bunch of txs to the first reactor's mempool
	// and wait for them all to be received in the others
	txs := checkTxs(t, reactors[0].mempool, 1000, UnknownPeerID, true)
	waitForTxsOnReactors(t, txs, reactors)
}

func TestReactorNoBroadcastToSender(t *testing.T) {
	t.Parallel()

	mconfig := memcfg.TestMempoolConfig()
	pconfig := testP2PConfig()
	const N = 2
	reactors := makeAndConnectReactors(t, mconfig, pconfig, N)
	defer func() {
		for _, r := range reactors {
			r.Stop()
		}
	}()

	// send a bunch of txs to the first reactor's mempool, claiming it came from peer
	// ensure peer gets no txs
	checkTxs(t, reactors[0].mempool, 1000, 1, true)
	ensureNoTxs(t, reactors[1], 100*time.Millisecond)
}

func TestFlappyBroadcastTxForPeerStopsWhenPeerStops(t *testing.T) {
	t.Parallel()

	testutils.FilterStability(t, testutils.Flappy)

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	mconfig := memcfg.TestMempoolConfig()
	pconfig := testP2PConfig()
	const N = 2
	reactors := makeAndConnectReactors(t, mconfig, pconfig, N)
	defer func() {
		for _, r := range reactors {
			r.Stop()
		}
	}()

	// stop peer
	sw := reactors[1].Switch
	sw.StopPeerForError(sw.Peers().List()[0], errors.New("some reason"))

	// check that we are not leaking any go-routines
	// i.e. broadcastTxRoutine finishes when peer is stopped
	leaktest.CheckTimeout(t, 10*time.Second)()
}

func TestFlappyBroadcastTxForPeerStopsWhenReactorStops(t *testing.T) {
	t.Parallel()

	testutils.FilterStability(t, testutils.Flappy)

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	mconfig := memcfg.TestMempoolConfig()
	pconfig := testP2PConfig()
	const N = 2
	reactors := makeAndConnectReactors(t, mconfig, pconfig, N)

	// stop reactors
	for _, r := range reactors {
		r.Stop()
	}

	// check that we are not leaking any go-routines
	// i.e. broadcastTxRoutine finishes when reactor is stopped
	leaktest.CheckTimeout(t, 10*time.Second)()
}

func TestMempoolIDsBasic(t *testing.T) {
	t.Parallel()

	ids := newMempoolIDs()

	id := p2pTypes.GenerateNodeKey().ID()

	ids.ReserveForPeer(id)
	assert.EqualValues(t, 1, ids.GetForPeer(id))
	ids.Reclaim(id)

	ids.ReserveForPeer(id)
	assert.EqualValues(t, 2, ids.GetForPeer(id))
	ids.Reclaim(id)
}

func TestMempoolIDsPanicsIfNodeRequestsOvermaxActiveIDs(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	// 0 is already reserved for UnknownPeerID
	ids := newMempoolIDs()

	for range maxActiveIDs - 1 {
		id := p2pTypes.GenerateNodeKey().ID()
		ids.ReserveForPeer(id)
	}

	assert.Panics(t, func() {
		id := p2pTypes.GenerateNodeKey().ID()

		ids.ReserveForPeer(id)
	})
}
