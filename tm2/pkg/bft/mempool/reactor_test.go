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
	p2pMock "github.com/gnolang/gno/tm2/pkg/p2p/mock"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	peer := p2pMock.GeneratePeers(t, 1)[0]

	ids.ReserveForPeer(peer)
	assert.EqualValues(t, 1, ids.GetForPeer(peer))
	ids.Reclaim(peer)

	ids.ReserveForPeer(peer)
	assert.EqualValues(t, 2, ids.GetForPeer(peer))
	ids.Reclaim(peer)
}

func TestMempoolIDsPanicsIfNodeRequestsOvermaxActiveIDs(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	// 0 is already reserved for UnknownPeerID
	ids := newMempoolIDs()

	// Reservations are keyed on the connection, so a distinct connection is
	// all it takes to consume an ID
	for range maxActiveIDs - 1 {
		ids.ReserveForPeer(&p2pMock.Peer{})
	}

	assert.Panics(t, func() {
		ids.ReserveForPeer(&p2pMock.Peer{})
	})
}

func TestMempoolIDsReserveIsIdempotent(t *testing.T) {
	t.Parallel()

	ids := newMempoolIDs()

	peer := p2pMock.GeneratePeers(t, 1)[0]

	ids.ReserveForPeer(peer)
	first := ids.GetForPeer(peer)
	require.NotEqual(t, UnknownPeerID, first, "the first reservation took an ID")
	require.Len(t, ids.activeIDs, 2) // UnknownPeerID and first

	// InitPeer runs once per connection, so a second reservation for the same
	// connection must not take a second ID.
	ids.ReserveForPeer(peer)
	assert.Equal(t, first, ids.GetForPeer(peer))
	assert.Len(t, ids.activeIDs, 2)

	ids.Reclaim(peer)
	assert.Len(t, ids.activeIDs, 1) // UnknownPeerID
	assert.Empty(t, ids.peerMap)
}

func TestMempoolIDsSupersededConnKeepsItsOwnID(t *testing.T) {
	t.Parallel()

	ids := newMempoolIDs()

	conns := p2pMock.GeneratePeers(t, 2)
	conn1, conn2 := conns[0], conns[1]

	// conn2 is a reconnect from the same node, so it carries the same peer ID,
	// and reaches InitPeer before the switch has torn conn1 down.
	id := conn1.ID()
	conn2.IDFn = func() p2pTypes.ID { return id }

	ids.ReserveForPeer(conn1)
	ids.ReserveForPeer(conn2)

	first, second := ids.GetForPeer(conn1), ids.GetForPeer(conn2)
	require.NotEqual(t, UnknownPeerID, first)
	require.NotEqual(t, UnknownPeerID, second)
	assert.NotEqual(t, first, second, "connections sharing a peer ID share an ID")

	// conn1's teardown gives back conn1's ID and leaves the connection that
	// superseded it able to gossip under an ID of its own.
	ids.Reclaim(conn1)
	assert.Equal(t, second, ids.GetForPeer(conn2))
	assert.NotEqual(t, UnknownPeerID, ids.GetForPeer(conn2))

	ids.Reclaim(conn2)
	assert.Len(t, ids.activeIDs, 1) // UnknownPeerID
	assert.Empty(t, ids.peerMap)
}

func TestMempoolIDsRemovedBeforeAdded(t *testing.T) {
	t.Parallel()

	mempool, cleanup := newMempoolWithApp(proxy.NewLocalClientCreator(kvstore.NewKVStoreApplication()))
	defer cleanup()

	memR := NewReactor(memcfg.TestMempoolConfig(), mempool)

	// A connection that fails on its first read is removed from every reactor
	// before the switch reaches AddPeer, so RemovePeer runs in between.
	for _, peer := range p2pMock.GeneratePeers(t, 100) {
		memR.InitPeer(peer)
		memR.RemovePeer(peer, nil)
		memR.AddPeer(peer)
	}

	assert.Len(t, memR.ids.activeIDs, 1) // UnknownPeerID
	assert.Empty(t, memR.ids.peerMap)
}
