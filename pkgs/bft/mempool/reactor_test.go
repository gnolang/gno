package mempool

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"

	"github.com/gnolang/gno/pkgs/bft/abci/example/kvstore"
	memcfg "github.com/gnolang/gno/pkgs/bft/mempool/config"
	"github.com/gnolang/gno/pkgs/bft/proxy"
	"github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/colors"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/log"
	"github.com/gnolang/gno/pkgs/p2p"
	p2pcfg "github.com/gnolang/gno/pkgs/p2p/config"
	"github.com/gnolang/gno/pkgs/p2p/mock"
	"github.com/gnolang/gno/pkgs/testutils"
)

type peerState struct {
	height int64
}

func (ps peerState) GetHeight() int64 {
	return ps.height
}

// mempoolLogger is a TestingLogger which uses a different
// color for each validator ("validator" key must exist).
func mempoolLogger() log.Logger {
	return log.TestingLoggerWithColorFn(func(keyvals ...interface{}) colors.Color {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if keyvals[i] == "validator" {
				num := keyvals[i+1].(int)
				switch num % 8 {
				case 0:
					return colors.Red
				case 1:
					return colors.Green
				case 2:
					return colors.Yellow
				case 3:
					return colors.Blue
				case 4:
					return colors.Magenta
				case 5:
					return colors.Cyan
				case 6:
					return colors.White
				case 7:
					return colors.Gray
				default:
					panic("should not happen")
				}
			}
		}
		return colors.None
	})
}

// connect N mempool reactors through N switches
func makeAndConnectReactors(mconfig *memcfg.MempoolConfig, pconfig *p2pcfg.P2PConfig, n int) []*Reactor {
	reactors := make([]*Reactor, n)
	logger := mempoolLogger()
	for i := 0; i < n; i++ {
		app := kvstore.NewKVStoreApplication()
		cc := proxy.NewLocalClientCreator(app)
		mempool, cleanup := newMempoolWithApp(cc)
		defer cleanup()

		reactors[i] = NewReactor(mconfig, mempool) // so we dont start the consensus states
		reactors[i].SetLogger(logger.With("validator", i))
	}

	p2p.MakeConnectedSwitches(pconfig, n, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("MEMPOOL", reactors[i])
		return s
	}, p2p.Connect2Switches)
	return reactors
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
	mconfig := memcfg.TestMempoolConfig()
	pconfig := p2pcfg.TestP2PConfig()
	const N = 4
	reactors := makeAndConnectReactors(mconfig, pconfig, N)
	defer func() {
		for _, r := range reactors {
			r.Stop()
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().List() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	// send a bunch of txs to the first reactor's mempool
	// and wait for them all to be received in the others
	txs := checkTxs(t, reactors[0].mempool, numTxs, UnknownPeerID, true)
	waitForTxsOnReactors(t, txs, reactors)
}

func TestReactorNoBroadcastToSender(t *testing.T) {
	mconfig := memcfg.TestMempoolConfig()
	pconfig := p2pcfg.TestP2PConfig()
	const N = 2
	reactors := makeAndConnectReactors(mconfig, pconfig, N)
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
	testutils.FilterStability(t, testutils.Flappy)

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	mconfig := memcfg.TestMempoolConfig()
	pconfig := p2pcfg.TestP2PConfig()
	const N = 2
	reactors := makeAndConnectReactors(mconfig, pconfig, N)
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
	testutils.FilterStability(t, testutils.Flappy)

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	mconfig := memcfg.TestMempoolConfig()
	pconfig := p2pcfg.TestP2PConfig()
	const N = 2
	reactors := makeAndConnectReactors(mconfig, pconfig, N)

	// stop reactors
	for _, r := range reactors {
		r.Stop()
	}

	// check that we are not leaking any go-routines
	// i.e. broadcastTxRoutine finishes when reactor is stopped
	leaktest.CheckTimeout(t, 10*time.Second)()
}

func TestMempoolIDsBasic(t *testing.T) {
	ids := newMempoolIDs()

	peer := mock.NewPeer(net.IP{127, 0, 0, 1})

	ids.ReserveForPeer(peer)
	assert.EqualValues(t, 1, ids.GetForPeer(peer))
	ids.Reclaim(peer)

	ids.ReserveForPeer(peer)
	assert.EqualValues(t, 2, ids.GetForPeer(peer))
	ids.Reclaim(peer)
}

func TestMempoolIDsPanicsIfNodeRequestsOvermaxActiveIDs(t *testing.T) {
	if testing.Short() {
		return
	}

	// 0 is already reserved for UnknownPeerID
	ids := newMempoolIDs()

	for i := 0; i < maxActiveIDs-1; i++ {
		peer := mock.NewPeer(net.IP{127, 0, 0, 1})
		ids.ReserveForPeer(peer)
	}

	assert.Panics(t, func() {
		peer := mock.NewPeer(net.IP{127, 0, 0, 1})
		ids.ReserveForPeer(peer)
	})
}
