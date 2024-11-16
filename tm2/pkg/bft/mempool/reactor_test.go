package mempool

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/events"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/versionset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	memcfg "github.com/gnolang/gno/tm2/pkg/bft/mempool/config"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pcfg "github.com/gnolang/gno/tm2/pkg/p2p/config"
	"github.com/gnolang/gno/tm2/pkg/testutils"
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

	for i := 0; i < n; i++ {
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
	MakeConnectedSwitches(t, pconfig, n, options)

	return reactors
}

func MakeConnectedSwitches(
	t *testing.T,
	cfg *p2pcfg.P2PConfig,
	n int,
	opts map[int][]p2p.SwitchOption,
) []*p2p.MultiplexSwitch {
	t.Helper()

	var (
		sws   = make([]*p2p.MultiplexSwitch, 0, n)
		ts    = make([]*p2p.MultiplexTransport, 0, n)
		addrs = make([]*p2pTypes.NetAddress, 0, n)

		// TODO remove
		lgs = make([]*slog.Logger, 0, n)
	)

	// Generate the switches
	for i := range n {
		var (
			key     = p2pTypes.GenerateNodeKey()
			tcpAddr = &net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 0, // random port
			}
		)

		addr, err := p2pTypes.NewNetAddress(key.ID(), tcpAddr)
		require.NoError(t, err)

		info := p2pTypes.NodeInfo{
			VersionSet: versionset.VersionSet{
				versionset.VersionInfo{
					Name:    "p2p",
					Version: "v0.0.0",
				},
			},
			NetAddress: addr,
			Network:    "testing",
			Software:   "p2ptest",
			Version:    "v1.2.3-rc.0-deadbeef",
			Channels:   []byte{0x01},
			Moniker:    fmt.Sprintf("node-%d", i),
			Other: p2pTypes.NodeInfoOther{
				TxIndex:    "off",
				RPCAddress: fmt.Sprintf("127.0.0.1:%d", 0),
			},
		}

		// TODO remove
		lg := slog.New(slog.NewTextHandler(os.Stdout, nil))

		// Create the multiplex transport
		multiplexTransport := p2p.NewMultiplexTransport(
			info,
			*key,
			conn.MConfigFromP2P(cfg),
			lg,
		)

		// Start the transport
		require.NoError(t, multiplexTransport.Listen(*info.NetAddress))

		t.Cleanup(func() {
			assert.NoError(t, multiplexTransport.Close())
		})

		dialAddr := multiplexTransport.NetAddress()
		addrs = append(addrs, &dialAddr)

		ts = append(ts, multiplexTransport)

		// TODO remove
		lgs = append(lgs, lg)
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	g, _ := errgroup.WithContext(ctx)

	for i := range n {
		// Make sure the switches connect to each other.
		// Set up event listeners to make sure
		// the setup method blocks until switches are connected
		// Create the multiplex switch
		newOpts := []p2p.SwitchOption{
			p2p.WithPersistentPeers(addrs),
		}

		newOpts = append(newOpts, opts[i]...)

		multiplexSwitch := p2p.NewMultiplexSwitch(ts[i], newOpts...)

		multiplexSwitch.SetLogger(lgs[i].With("sw", i)) // TODO remove

		ch, unsubFn := multiplexSwitch.Subscribe(func(event events.Event) bool {
			return event.Type() == events.PeerConnected
		})

		// Start the switch
		require.NoError(t, multiplexSwitch.Start())

		sws = append(sws, multiplexSwitch)

		g.Go(func() error {
			defer func() {
				unsubFn()
			}()

			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()

			connectedPeers := make(map[p2pTypes.ID]struct{})

			for {
				select {
				case evRaw := <-ch:
					ev := evRaw.(events.PeerConnectedEvent)

					connectedPeers[ev.PeerID] = struct{}{}

					if len(connectedPeers) == n-1 {
						return nil
					}
				case <-timer.C:
					return errors.New("timed out waiting for peers to connect")
				}
			}
		})

		sws[i].DialPeers(addrs...)
	}

	require.NoError(t, g.Wait())

	fmt.Printf("\n\nDONE\n\n")

	return sws
}

func waitForTxsOnReactors(t *testing.T, txs types.Txs, reactors []*Reactor) {
	t.Helper()

	// wait for the txs in all mempools
	wg := new(sync.WaitGroup)
	for i, reactor := range reactors {
		wg.Add(1)
		go func(r *Reactor, reactorIndex int) {
			defer wg.Done()
			waitForTxsOnReactor(t, txs, r, reactorIndex)
		}(reactor, i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	timer := time.After(timeout)
	select {
	case <-timer:
		t.Fatal("Timed out waiting for txs")
	case <-done:
	}
}

func waitForTxsOnReactor(t *testing.T, txs types.Txs, reactor *Reactor, reactorIndex int) {
	t.Helper()

	mempool := reactor.mempool
	for mempool.Size() < len(txs) {
		time.Sleep(time.Millisecond * 100)
	}

	reapedTxs := mempool.ReapMaxTxs(len(txs))
	for i, tx := range txs {
		assert.Equalf(t, tx, reapedTxs[i],
			"txs at index %d on reactor %d don't match: %v vs %v", i, reactorIndex, tx, reapedTxs[i])
	}
}

// ensure no txs on reactor after some timeout
func ensureNoTxs(t *testing.T, reactor *Reactor, timeout time.Duration) {
	t.Helper()

	time.Sleep(timeout) // wait for the txs in all mempools
	assert.Zero(t, reactor.mempool.Size())
}

const (
	numTxs  = 1000
	timeout = 120 * time.Second // ridiculously high because CircleCI is slow
)

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
	txs := checkTxs(t, reactors[0].mempool, numTxs, UnknownPeerID, true)
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
	checkTxs(t, reactors[0].mempool, numTxs, 1, true)
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

	for i := 0; i < maxActiveIDs-1; i++ {
		id := p2pTypes.GenerateNodeKey().ID()
		ids.ReserveForPeer(id)
	}

	assert.Panics(t, func() {
		id := p2pTypes.GenerateNodeKey().ID()

		ids.ReserveForPeer(id)
	})
}
