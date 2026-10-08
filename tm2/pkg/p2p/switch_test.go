package p2p

import (
	"context"
	"net"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/p2p/dial"
	"github.com/gnolang/gno/tm2/pkg/p2p/events"
	"github.com/gnolang/gno/tm2/pkg/p2p/mock"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiplexSwitch_Options(t *testing.T) {
	t.Parallel()

	t.Run("custom reactors", func(t *testing.T) {
		t.Parallel()

		var (
			name        = "custom reactor"
			mockReactor = &mockReactor{
				setSwitchFn: func(s Switch) {
					require.NotNil(t, s)
				},
			}
		)

		sw := NewMultiplexSwitch(nil, WithReactor(name, mockReactor))

		assert.Equal(t, mockReactor, sw.reactors[name])
	})

	t.Run("persistent peers", func(t *testing.T) {
		t.Parallel()

		peers := generateNetAddr(t, 10)

		sw := NewMultiplexSwitch(nil, WithPersistentPeers(peers))

		for _, p := range peers {
			assert.True(t, sw.isPersistentPeer(p.ID))
		}
	})

	t.Run("seeds", func(t *testing.T) {
		t.Parallel()

		seeds := generateNetAddr(t, 10)

		sw := NewMultiplexSwitch(nil, WithSeeds(seeds))

		for _, s := range seeds {
			addr, ok := sw.seeds.Load(s.ID)

			require.True(t, ok)
			assert.Equal(t, s, addr)
		}
	})

	t.Run("private peers", func(t *testing.T) {
		t.Parallel()

		var (
			peers = generateNetAddr(t, 10)
			ids   = make([]types.ID, 0, len(peers))
		)

		for _, p := range peers {
			ids = append(ids, p.ID)
		}

		sw := NewMultiplexSwitch(nil, WithPrivatePeers(ids))

		for _, p := range peers {
			assert.True(t, sw.isPrivatePeer(p.ID))
		}
	})

	t.Run("max inbound peers", func(t *testing.T) {
		t.Parallel()

		maxInbound := uint64(500)

		sw := NewMultiplexSwitch(nil, WithMaxInboundPeers(maxInbound))

		assert.Equal(t, maxInbound, sw.maxInboundPeers)
	})

	t.Run("max outbound peers", func(t *testing.T) {
		t.Parallel()

		maxOutbound := uint64(500)

		sw := NewMultiplexSwitch(nil, WithMaxOutboundPeers(maxOutbound))

		assert.Equal(t, maxOutbound, sw.maxOutboundPeers)
	})
}

func TestMultiplexSwitch_Broadcast(t *testing.T) {
	t.Parallel()

	var (
		wg sync.WaitGroup

		expectedChID = byte(10)
		expectedData = []byte("broadcast data")

		mockTransport = &mockTransport{
			acceptFn: func(_ context.Context, _ PeerBehavior) (PeerConn, error) {
				return nil, errors.New("constant error")
			},
		}

		peers = mock.GeneratePeers(t, 10)
		sw    = NewMultiplexSwitch(mockTransport)
	)

	// Create a new peer set.
	// The switch services read the peer set as soon as they are started,
	// so it has to be in place before OnStart
	sw.peers = newSet()

	for _, p := range peers {
		wg.Add(1)

		p.SendFn = func(chID byte, data []byte) bool {
			wg.Done()

			require.Equal(t, expectedChID, chID)
			assert.Equal(t, expectedData, data)

			return false
		}

		// Load it up with peers
		require.NoError(t, sw.peers.Add(p))
	}

	require.NoError(t, sw.OnStart())
	t.Cleanup(sw.OnStop)

	// Broadcast the data
	sw.Broadcast(expectedChID, expectedData)

	wg.Wait()
}

func TestMultiplexSwitch_Peers(t *testing.T) {
	t.Parallel()

	var (
		peers = mock.GeneratePeers(t, 10)
		sw    = NewMultiplexSwitch(nil)
	)

	// Create a new peer set
	sw.peers = newSet()

	for _, p := range peers {
		// Load it up with peers
		require.NoError(t, sw.peers.Add(p))
	}

	// Broadcast the data
	ps := sw.Peers()

	require.EqualValues(
		t,
		len(peers),
		ps.NumInbound()+ps.NumOutbound(),
	)

	for _, p := range peers {
		assert.True(t, ps.Has(p.ID()))
	}
}

func TestMultiplexSwitch_StopPeer(t *testing.T) {
	t.Parallel()

	t.Run("peer not persistent", func(t *testing.T) {
		t.Parallel()

		var (
			p             = mock.GeneratePeers(t, 1)[0]
			mockTransport = &mockTransport{
				removeFn: func(removedPeer PeerConn) {
					assert.Equal(t, p.ID(), removedPeer.ID())
				},
			}

			sw = NewMultiplexSwitch(mockTransport)
		)

		// Create a new peer set
		sw.peers = newSet()

		// Save the single peer
		require.NoError(t, sw.peers.Add(p))

		// Stop and remove the peer
		sw.StopPeerForError(p, nil)

		// Make sure the peer is removed
		assert.False(t, sw.peers.Has(p.ID()))
	})

	t.Run("persistent peer", func(t *testing.T) {
		t.Parallel()

		var (
			p             = mock.GeneratePeers(t, 1)[0]
			mockTransport = &mockTransport{
				removeFn: func(removedPeer PeerConn) {
					assert.Equal(t, p.ID(), removedPeer.ID())
				},
			}

			sw = NewMultiplexSwitch(
				mockTransport,
				WithPersistentPeers([]*types.NetAddress{p.SocketAddr()}),
			)
		)

		require.True(t, sw.isPersistentPeer(p.ID()))

		// Create a new peer set
		sw.peers = newSet()

		// Save the single peer
		require.NoError(t, sw.peers.Add(p))

		// Stop and remove the peer
		sw.StopPeerForError(p, nil)

		// Make sure the peer is removed
		assert.False(t, sw.peers.Has(p.ID()))

		// The redial loop owns persistent peers: nothing is queued here
		assert.Nil(t, sw.dialQueue.Peek())
		assert.Nil(t, sw.persistentDialQueue.Peek())
	})
}

func TestMultiplexSwitch_StopPeer_AnnouncesDisconnectOnce(t *testing.T) {
	t.Parallel()

	var (
		p  = mock.GeneratePeers(t, 1)[0]
		sw = NewMultiplexSwitch(&mockTransport{
			removeFn: func(PeerConn) {},
		})
	)

	sw.peers = newSet()
	require.NoError(t, sw.peers.Add(p))

	subCh, unsubFn := sw.Subscribe(func(event events.Event) bool {
		return event.Type() == events.PeerDisconnected
	})
	defer unsubFn()

	// A second teardown of the same connection finds the peer set entry gone
	sw.stopAndRemovePeer(p, nil)
	sw.stopAndRemovePeer(p, nil)

	// Notify delivers into the buffered subscription channel before it
	// returns, so everything announced is already queued
	assert.Len(t, subCh, 1, "one disconnect for the one connection the peer set held")
}

// peerErrorOnStart reports a peer error to the switch from inside Start, the
// way an MConnection recv routine does on a bad first packet.
type peerErrorOnStart struct {
	*mock.Peer

	startFn func() error
}

func (p *peerErrorOnStart) Start() error {
	return p.startFn()
}

func TestMultiplexSwitch_AddPeerRemovedBeforeAdded(t *testing.T) {
	t.Parallel()

	var (
		calls []string

		record = func(call string) {
			calls = append(calls, call)
		}

		mockTransport = &mockTransport{
			removeFn: func(PeerConn) {},
		}

		mockReactor = &mockReactor{
			initPeerFn: func(peer PeerConn) PeerConn {
				record("InitPeer")

				return peer
			},
			addPeerFn: func(PeerConn) {
				record("AddPeer")
			},
			removePeerFn: func(PeerConn, any) {
				record("RemovePeer")
			},
		}

		sw = NewMultiplexSwitch(
			mockTransport,
			WithReactor("mock", mockReactor),
		)

		p = &peerErrorOnStart{Peer: mock.GeneratePeers(t, 1)[0]}
	)

	disconnects, unsubFn := sw.Subscribe(func(event events.Event) bool {
		return event.Type() == events.PeerDisconnected
	})
	defer unsubFn()

	// An error on the peer's first read reaches the switch from inside Start.
	p.startFn = func() error {
		sw.StopPeerForError(p, errors.New("peer error"))

		return nil
	}

	// addPeer must refuse a peer that was stopped while it was being added.
	// stopAndRemovePeer removes from the peer set last, so its Remove runs
	// before the Add, and nothing would ever remove the peer again
	require.ErrorIs(t, sw.addPeer(p), errPeerStopped)

	// InitPeer is the hook that runs first, so it is the one that can pair
	// with the RemovePeer that frees what it took. AddPeer never runs
	assert.Equal(t, []string{"InitPeer", "RemovePeer"}, calls)

	// The peer holds neither a slot nor its ID in the peer set
	assert.False(t, sw.peers.Has(p.ID()))
	assert.Zero(t, sw.peers.NumInbound())
	assert.Empty(t, sw.peers.List())

	// A connection stopped before it joined the peer set is never announced
	assert.Len(t, disconnects, 0, "no disconnect for a peer that never joined the peer set")
}

func TestMultiplexSwitch_AddPeerRejectsDuplicateBeforeInit(t *testing.T) {
	t.Parallel()

	var (
		calls []string

		mockTransport = &mockTransport{
			removeFn: func(PeerConn) {},
		}

		mockReactor = &mockReactor{
			initPeerFn: func(peer PeerConn) PeerConn {
				calls = append(calls, "InitPeer")

				return peer
			},
			addPeerFn:    func(PeerConn) { calls = append(calls, "AddPeer") },
			removePeerFn: func(PeerConn, any) { calls = append(calls, "RemovePeer") },
		}

		sw = NewMultiplexSwitch(
			mockTransport,
			WithReactor("mock", mockReactor),
		)

		peers = mock.GeneratePeers(t, 2)
		live  = peers[0]
		dup   = peers[1]
	)

	// dup is a reconnect from the same node, so it carries the same peer ID.
	id := live.ID()
	dup.IDFn = func() types.ID { return id }

	require.NoError(t, sw.addPeer(live))
	require.Equal(t, []string{"InitPeer", "AddPeer"}, calls)

	// The duplicate is refused before any reactor sees it, so it neither
	// starts nor leaves reactor state behind
	require.ErrorIs(t, sw.addPeer(dup), errDuplicatePeer)
	assert.Equal(t, []string{"InitPeer", "AddPeer"}, calls)

	// The live peer keeps its peer set entry and its slot
	assert.Same(t, live, sw.peers.Get(id))
	assert.EqualValues(t, 1, sw.peers.NumInbound())
}

func TestMultiplexSwitch_StopPeerLeavesSupersedingConnAlone(t *testing.T) {
	t.Parallel()

	var (
		removed []PeerConn

		mockTransport = &mockTransport{
			removeFn: func(PeerConn) {},
		}

		mockReactor = &mockReactor{
			removePeerFn: func(peer PeerConn, _ any) {
				removed = append(removed, peer)
			},
		}

		sw = NewMultiplexSwitch(
			mockTransport,
			WithReactor("mock", mockReactor),
		)

		peers = mock.GeneratePeers(t, 2)
		live  = peers[0]
		dup   = peers[1]
	)

	id := live.ID()
	dup.IDFn = func() types.ID { return id }

	require.NoError(t, sw.addPeer(live))

	// dup lost the race for the peer set and reports an error of its own. Its
	// own reactor state is given back, since a reactor keying that state on
	// the connection can tell the two apart, but the peer set entry under the
	// ID it shares with live is live's
	sw.StopPeerForError(dup, errors.New("duplicate connection error"))

	assert.Equal(t, []PeerConn{dup}, removed, "only the superseded connection")
	assert.Same(t, live, sw.peers.Get(id), "the live peer is still registered")
	assert.EqualValues(t, 1, sw.peers.NumInbound())

	// The live peer's own teardown still works.
	sw.StopPeerForError(live, errors.New("peer error"))

	assert.Equal(t, []PeerConn{dup, live}, removed)
	assert.False(t, sw.peers.Has(id))
	assert.Zero(t, sw.peers.NumInbound())
}

func TestMultiplexSwitch_AddPeerUnwindsReactorStateOnError(t *testing.T) {
	t.Parallel()

	var (
		calls []string

		mockReactor = &mockReactor{
			initPeerFn: func(peer PeerConn) PeerConn {
				calls = append(calls, "InitPeer")

				return peer
			},
			addPeerFn:    func(PeerConn) { calls = append(calls, "AddPeer") },
			removePeerFn: func(PeerConn, any) { calls = append(calls, "RemovePeer") },
		}

		// The residual race the duplicate pre-check cannot cover: another
		// connection wins sw.peers.Add between the check and the Add
		mockSet = &mockSet{
			addFn: func(PeerConn) error { return errDuplicatePeer },
		}

		sw = NewMultiplexSwitch(
			&mockTransport{removeFn: func(PeerConn) {}},
			WithReactor("mock", mockReactor),
		)

		p = mock.GeneratePeers(t, 1)[0]
	)

	sw.peers = mockSet

	require.ErrorIs(t, sw.addPeer(p), errDuplicatePeer)

	// Whatever InitPeer took is given back, so a refused connection does not
	// hold reactor state for the lifetime of the process
	assert.Equal(t, []string{"InitPeer", "RemovePeer"}, calls)
}

func TestMultiplexSwitch_AddPeerOutboundLimit(t *testing.T) {
	t.Parallel()

	const maxOutbound = 3

	sw := NewMultiplexSwitch(
		&mockTransport{removeFn: func(PeerConn) {}},
		WithMaxOutboundPeers(maxOutbound),
	)

	// DialPeers only checks the limit when an address is queued, where
	// NumOutbound cannot have changed yet, so a single batch of queued dials
	// would otherwise overshoot it without bound
	for i, p := range mock.GeneratePeers(t, maxOutbound+2) {
		p.IsOutboundFn = func() bool { return true }

		if i < maxOutbound {
			require.NoError(t, sw.addPeer(p))

			continue
		}

		assert.ErrorIs(t, sw.addPeer(p), errMaxOutboundPeers)
	}

	assert.EqualValues(t, maxOutbound, sw.peers.NumOutbound())
}

func TestMultiplexSwitch_DialLoop(t *testing.T) {
	t.Parallel()

	t.Run("peer already connected", func(t *testing.T) {
		t.Parallel()

		ctx, cancelFn := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelFn()

		var (
			ch = make(chan struct{}, 1)

			peerDialed bool

			p        = mock.GeneratePeers(t, 1)[0]
			dialTime = time.Now().Add(-5 * time.Second) // in the past

			mockSet = &mockSet{
				hasFn: func(id types.ID) bool {
					require.Equal(t, p.ID(), id)

					cancelFn()

					ch <- struct{}{}

					return true
				},
			}

			mockTransport = &mockTransport{
				dialFn: func(
					_ context.Context,
					_ types.NetAddress,
					_ PeerBehavior,
				) (PeerConn, error) {
					peerDialed = true

					return nil, nil
				},
			}

			sw = NewMultiplexSwitch(mockTransport)
		)

		sw.peers = mockSet

		// Prepare the dial queue
		sw.dialQueue.Push(dial.Item{
			Time:    dialTime,
			Address: p.SocketAddr(),
		})

		// Run the dial loop
		go sw.runDialLoop(ctx)

		select {
		case <-ch:
		case <-time.After(5 * time.Second):
		}

		assert.False(t, peerDialed)
	})

	t.Run("peer undialable", func(t *testing.T) {
		t.Parallel()

		ctx, cancelFn := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelFn()

		var (
			ch = make(chan struct{}, 1)

			peerDialed bool

			p        = mock.GeneratePeers(t, 1)[0]
			dialTime = time.Now().Add(-5 * time.Second) // in the past

			mockSet = &mockSet{
				hasFn: func(id types.ID) bool {
					require.Equal(t, p.ID(), id)

					return false
				},
			}

			mockTransport = &mockTransport{
				dialFn: func(
					_ context.Context,
					_ types.NetAddress,
					_ PeerBehavior,
				) (PeerConn, error) {
					peerDialed = true

					cancelFn()

					ch <- struct{}{}

					return nil, errors.New("invalid dial")
				},
			}

			sw = NewMultiplexSwitch(mockTransport)
		)

		sw.peers = mockSet

		// Prepare the dial queue
		sw.dialQueue.Push(dial.Item{
			Time:    dialTime,
			Address: p.SocketAddr(),
		})

		// Run the dial loop
		go sw.runDialLoop(ctx)

		select {
		case <-ch:
		case <-time.After(5 * time.Second):
		}

		assert.True(t, peerDialed)
	})

	t.Run("peer dialed and added", func(t *testing.T) {
		t.Parallel()

		ctx, cancelFn := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelFn()

		var (
			ch = make(chan struct{}, 1)

			p        = mock.GeneratePeers(t, 1)[0]
			dialTime = time.Now().Add(-5 * time.Second) // in the past

			mockTransport = &mockTransport{
				dialFn: func(
					_ context.Context,
					_ types.NetAddress,
					_ PeerBehavior,
				) (PeerConn, error) {
					cancelFn()

					ch <- struct{}{}

					return p, nil
				},
			}

			sw = NewMultiplexSwitch(mockTransport)
		)

		// Prepare the dial queue
		sw.dialQueue.Push(dial.Item{
			Time:    dialTime,
			Address: p.SocketAddr(),
		})

		// Run the dial loop
		go sw.runDialLoop(ctx)

		select {
		case <-ch:
		case <-time.After(5 * time.Second):
		}

		// The dial signals before dialPeer registers the peer
		require.Eventually(
			t,
			func() bool { return sw.Peers().Has(p.ID()) },
			5*time.Second,
			10*time.Millisecond,
		)
	})
}

func TestMultiplexSwitch_DialPeer_RejectedBeforeStart(t *testing.T) {
	t.Parallel()

	// dialRejected dials a peer the switch refuses before starting it,
	// and returns how many times its connection was closed
	dialRejected := func(t *testing.T, sw *MultiplexSwitch, p *mock.Peer) int {
		t.Helper()

		var closed int

		p.IsOutboundFn = func() bool { return true }
		p.CloseConnFn = func() error {
			closed++

			return nil
		}

		sw.transport = &mockTransport{
			dialFn: func(context.Context, types.NetAddress, PeerBehavior) (PeerConn, error) {
				return p, nil
			},
		}

		sw.dialPeer(t.Context(), p.SocketAddr())

		return closed
	}

	t.Run("outbound limit reached", func(t *testing.T) {
		t.Parallel()

		sw := NewMultiplexSwitch(nil, WithMaxOutboundPeers(0))

		p := mock.GeneratePeers(t, 1)[0]

		assert.Equal(t, 1, dialRejected(t, sw, p))
		assert.False(t, sw.Peers().Has(p.ID()))
	})

	t.Run("duplicate peer", func(t *testing.T) {
		t.Parallel()

		p := mock.GeneratePeers(t, 1)[0]

		sw := NewMultiplexSwitch(nil)
		sw.peers = &mockSet{
			hasFn: func(id types.ID) bool { return id == p.ID() },
		}

		assert.Equal(t, 1, dialRejected(t, sw, p))
	})
}

func TestMultiplexSwitch_AcceptLoop(t *testing.T) {
	t.Parallel()

	t.Run("inbound limit reached", func(t *testing.T) {
		t.Parallel()

		ctx, cancelFn := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelFn()

		var (
			ch         = make(chan struct{}, 1)
			maxInbound = uint64(10)

			p = mock.GeneratePeers(t, 1)[0]

			mockTransport = &mockTransport{
				// Honour the context so cancelling it stops the accept loop.
				// Without this the loop spins on this mock at full speed for the
				// lifetime of the test binary, since runAcceptLoop only leaves
				// the loop on an error from Accept.
				acceptFn: func(ctx context.Context, _ PeerBehavior) (PeerConn, error) {
					if err := ctx.Err(); err != nil {
						return nil, err
					}

					return p, nil
				},
				removeFn: func(removedPeer PeerConn) {
					assert.Equal(t, p.ID(), removedPeer.ID())

					// The accept loop keeps accepting the same peer until the
					// context is cancelled, so the signal is best-effort; never
					// block it.
					select {
					case ch <- struct{}{}:
					default:
					}
				},
			}

			ps = &mockSet{
				numInboundFn: func() uint64 {
					return maxInbound
				},
			}

			sw = NewMultiplexSwitch(
				mockTransport,
				WithMaxInboundPeers(maxInbound),
			)
		)

		// Set the peer set
		sw.peers = ps

		// Run the accept loop
		go sw.runAcceptLoop(ctx)

		select {
		case <-ch: // the peer was removed
		case <-time.After(5 * time.Second):
			t.Fatal("the peer was not removed")
		}

		cancelFn() // stop the accept loop
	})

	t.Run("peer accepted", func(t *testing.T) {
		t.Parallel()

		ctx, cancelFn := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelFn()

		var (
			ch         = make(chan struct{}, 1)
			maxInbound = uint64(10)

			p = mock.GeneratePeers(t, 1)[0]

			mockTransport = &mockTransport{
				// Honour the context so cancelling it stops the accept loop.
				// Without this the loop spins on this mock at full speed for the
				// lifetime of the test binary, since runAcceptLoop only leaves
				// the loop on an error from Accept.
				acceptFn: func(ctx context.Context, _ PeerBehavior) (PeerConn, error) {
					if err := ctx.Err(); err != nil {
						return nil, err
					}

					return p, nil
				},
			}

			ps = &mockSet{
				numInboundFn: func() uint64 {
					return maxInbound - 1 // available slot
				},
				addFn: func(peer PeerConn) error {
					assert.Equal(t, p.ID(), peer.ID())

					// The accept loop keeps accepting the same peer until the
					// context is cancelled, so the signal is best-effort; never
					// block it.
					select {
					case ch <- struct{}{}:
					default:
					}

					return nil
				},
			}

			sw = NewMultiplexSwitch(
				mockTransport,
				WithMaxInboundPeers(maxInbound),
			)
		)

		// Set the peer set
		sw.peers = ps

		// Run the accept loop
		go sw.runAcceptLoop(ctx)

		select {
		case <-ch: // the peer was added
		case <-time.After(5 * time.Second):
			t.Fatal("the peer was not added")
		}

		cancelFn() // stop the accept loop
	})
}

func TestMultiplexSwitch_RedialLoop(t *testing.T) {
	t.Parallel()

	t.Run("no peers to dial", func(t *testing.T) {
		t.Parallel()

		var (
			ch = make(chan struct{}, 1)

			peersChecked = 0
			peers        = mock.GeneratePeers(t, 10)

			ps = &mockSet{
				hasFn: func(id types.ID) bool {
					exists := false
					for _, p := range peers {
						if p.ID() == id {
							exists = true

							break
						}
					}

					require.True(t, exists)

					peersChecked++

					if peersChecked == len(peers) {
						ch <- struct{}{}
					}

					return true
				},
			}
		)

		// Make sure the peers are the
		// switch persistent peers
		addrs := make([]*types.NetAddress, 0, len(peers))

		for _, p := range peers {
			addrs = append(addrs, p.SocketAddr())
		}

		// Create the switch
		sw := NewMultiplexSwitch(
			nil,
			WithPersistentPeers(addrs),
		)

		// Set the peer set
		sw.peers = ps

		// Run the redial loop
		ctx, cancelFn := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelFn()

		go sw.runRedialLoop(ctx)

		select {
		case <-ch:
		case <-time.After(5 * time.Second):
		}

		assert.Equal(t, len(peers), peersChecked)
	})

	t.Run("missing peers dialed", func(t *testing.T) {
		t.Parallel()

		var (
			peers       = mock.GeneratePeers(t, 10)
			missingPeer = peers[0]
			missingAddr = missingPeer.SocketAddr()

			peersDialed []types.NetAddress

			mockTransport = &mockTransport{
				dialFn: func(
					_ context.Context,
					address types.NetAddress,
					_ PeerBehavior,
				) (PeerConn, error) {
					peersDialed = append(peersDialed, address)

					if address.Equals(*missingPeer.SocketAddr()) {
						return missingPeer, nil
					}

					return nil, errors.New("invalid dial")
				},
			}
			ps = &mockSet{
				hasFn: func(id types.ID) bool {
					return id != missingPeer.ID()
				},
			}
		)

		// Make sure the peers are the
		// switch persistent peers
		addrs := make([]*types.NetAddress, 0, len(peers))

		for _, p := range peers {
			addrs = append(addrs, p.SocketAddr())
		}

		// Create the switch
		sw := NewMultiplexSwitch(
			mockTransport,
			WithPersistentPeers(addrs),
		)

		// Set the peer set
		sw.peers = ps

		// Run the redial loop
		ctx, cancelFn := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelFn()

		var wg sync.WaitGroup

		wg.Add(2)

		go func() {
			defer wg.Done()

			sw.runRedialLoop(ctx)
		}()

		go func() {
			defer wg.Done()

			deadline := time.After(5 * time.Second)

			for {
				select {
				case <-deadline:
					return
				default:
					if !sw.persistentDialQueue.Has(missingAddr) {
						continue
					}

					cancelFn()

					return
				}
			}
		}()

		wg.Wait()

		require.True(t, sw.persistentDialQueue.Has(missingAddr))
		assert.Equal(t, missingAddr, sw.persistentDialQueue.Peek().Address)
	})

	t.Run("connect and disconnect events drive the redial", func(t *testing.T) {
		t.Parallel()

		addr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithPersistentPeers([]*types.NetAddress{addr}),
		)

		// The peer is in the peer set while connected, as the redial loop reads
		// it from its own goroutine
		var connected atomic.Bool

		sw.peers = &mockSet{
			hasFn: func(id types.ID) bool { return connected.Load() && id == addr.ID },
		}

		// Only the start pass and the events can queue a dial: the next tick
		// is an hour away
		sw.redialInterval = time.Hour

		go sw.runRedialLoop(t.Context())

		// The start pass queues the first dial
		require.Eventually(t, func() bool {
			return sw.persistentDialQueue.Pop() != nil
		}, 5*time.Second, 5*time.Millisecond)

		// A dial queued while that one was in flight
		sw.persistentDialQueue.Push(dial.Item{Time: time.Now().Add(time.Minute), Address: addr})

		// The peer connects: its queued dial is removed
		connected.Store(true)

		sw.events.Notify(events.PeerConnectedEvent{PeerID: addr.ID})

		require.Eventually(t, func() bool {
			return sw.persistentDialQueue.Peek() == nil
		}, 5*time.Second, 5*time.Millisecond)

		// The peer drops: it is queued again without waiting for the tick
		sent := time.Now()

		connected.Store(false)

		sw.events.Notify(events.PeerDisconnectedEvent{PeerID: addr.ID})

		require.Eventually(t, func() bool {
			return sw.persistentDialQueue.Peek() != nil
		}, 5*time.Second, 5*time.Millisecond)

		// The connect left the attempt count in place, so the dial waits for
		// the first backoff step, about a second
		item := sw.persistentDialQueue.Peek()

		require.NotNil(t, item)
		assert.False(t, item.Time.Before(sent.Add(900*time.Millisecond)))
	})
}

func TestMultiplexSwitch_DialSeed(t *testing.T) {
	t.Parallel()

	t.Run("no seeds configured", func(t *testing.T) {
		t.Parallel()

		sw := NewMultiplexSwitch(&mockTransport{})

		sw.dialSeed()

		assert.Nil(t, sw.dialQueue.Peek())
	})

	t.Run("dialable item in the queue", func(t *testing.T) {
		t.Parallel()

		var (
			addrs    = generateNetAddr(t, 2)
			seedAddr = addrs[0]
			peerAddr = addrs[1]
		)

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds([]*types.NetAddress{seedAddr}),
		)

		// The switch still has something to dial right now
		sw.dialQueue.Push(dial.Item{
			Time:    time.Now(),
			Address: peerAddr,
		})

		sw.dialSeed()

		// The seed should not have been queued
		assert.False(t, sw.dialQueue.Has(seedAddr))
	})

	t.Run("dialable persistent item in the queue", func(t *testing.T) {
		t.Parallel()

		var (
			addrs    = generateNetAddr(t, 2)
			seedAddr = addrs[0]
			peerAddr = addrs[1]
		)

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds([]*types.NetAddress{seedAddr}),
		)

		// The switch still has a persistent peer to dial right now
		sw.persistentDialQueue.Push(dial.Item{
			Time:    time.Now(),
			Address: peerAddr,
		})

		sw.dialSeed()

		// The seed should not have been queued
		assert.False(t, sw.dialQueue.Has(seedAddr))
	})

	t.Run("persistent item fully backed off", func(t *testing.T) {
		t.Parallel()

		var (
			addrs    = generateNetAddr(t, 2)
			seedAddr = addrs[0]
			peerAddr = addrs[1]
		)

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds([]*types.NetAddress{seedAddr}),
		)

		// A node whose persistent peers are all down holds only backed-off
		// persistent dials, and must still fall back to its seeds
		sw.persistentDialQueue.Push(dial.Item{
			Time:    time.Now().Add(10 * time.Minute),
			Address: peerAddr,
		})

		sw.dialSeed()

		assert.True(t, sw.dialQueue.Has(seedAddr))
	})

	t.Run("queued items fully backed off", func(t *testing.T) {
		t.Parallel()

		var (
			addrs    = generateNetAddr(t, 2)
			seedAddr = addrs[0]
			peerAddr = addrs[1]
		)

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds([]*types.NetAddress{seedAddr}),
		)

		// Nothing in the queue can be dialed before this point in time
		sw.dialQueue.Push(dial.Item{
			Time:    time.Now().Add(10 * time.Minute),
			Address: peerAddr,
		})

		sw.dialSeed()

		assert.True(t, sw.dialQueue.Has(seedAddr))
	})

	t.Run("empty queue", func(t *testing.T) {
		t.Parallel()

		seedAddr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds([]*types.NetAddress{seedAddr}),
		)

		sw.dialSeed()

		require.NotNil(t, sw.dialQueue.Peek())
		assert.Equal(t, seedAddr, sw.dialQueue.Peek().Address)
	})

	t.Run("outbound peer limit reached", func(t *testing.T) {
		t.Parallel()

		seedAddr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds([]*types.NetAddress{seedAddr}),
			WithMaxOutboundPeers(1),
		)

		// Every outbound slot is taken
		sw.peers = &mockSet{
			numOutboundFn: func() uint64 {
				return 1
			},
		}

		sw.dialSeed()

		assert.Nil(t, sw.dialQueue.Peek())
	})

	t.Run("connected seed skipped", func(t *testing.T) {
		t.Parallel()

		seedAddr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds([]*types.NetAddress{seedAddr}),
		)

		// The seed is already an active peer
		sw.peers = &mockSet{
			hasFn: func(id types.ID) bool {
				return id == seedAddr.ID
			},
		}

		sw.dialSeed()

		assert.Nil(t, sw.dialQueue.Peek())
	})

	t.Run("queued seed not duplicated", func(t *testing.T) {
		t.Parallel()

		seedAddr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds([]*types.NetAddress{seedAddr}),
		)

		// The seed is already waiting in the dial queue
		sw.dialQueue.Push(dial.Item{
			Time:    time.Now().Add(10 * time.Minute),
			Address: seedAddr,
		})

		sw.dialSeed()

		// A single entry should remain
		require.NotNil(t, sw.dialQueue.Pop())
		assert.Nil(t, sw.dialQueue.Pop())
	})

	t.Run("single seed dialed per round", func(t *testing.T) {
		t.Parallel()

		seeds := generateNetAddr(t, 10)

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds(seeds),
		)

		sw.dialSeed()

		// Exactly one of the seeds should have been queued
		item := sw.dialQueue.Pop()

		require.NotNil(t, item)
		assert.Contains(t, seeds, item.Address)
		assert.Nil(t, sw.dialQueue.Pop())
	})

	t.Run("persistent seeds are not candidates", func(t *testing.T) {
		t.Parallel()

		// Nine of the ten seeds are also persistent peers, which DialPeers
		// leaves to the redial loop. Picking one would waste the round
		seeds := generateNetAddr(t, 10)

		for range 20 {
			sw := NewMultiplexSwitch(
				&mockTransport{},
				WithSeeds(seeds),
				WithPersistentPeers(seeds[1:]),
			)

			sw.dialSeed()

			item := sw.dialQueue.Pop()

			require.NotNil(t, item)
			assert.Equal(t, seeds[0], item.Address)
		}
	})
}

func TestMultiplexSwitch_SeedDialLoop(t *testing.T) {
	t.Parallel()

	t.Run("seed dialed on start", func(t *testing.T) {
		t.Parallel()

		seedAddr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithSeeds([]*types.NetAddress{seedAddr}),
		)

		ctx := t.Context()

		go sw.runSeedDialLoop(ctx)

		// The bootstrap round runs before the first tick
		require.Eventually(
			t,
			func() bool {
				return sw.dialQueue.Has(seedAddr)
			},
			5*time.Second,
			10*time.Millisecond,
		)
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()

		var (
			sw   = NewMultiplexSwitch(&mockTransport{})
			done = make(chan struct{})
		)

		ctx, cancelFn := context.WithCancel(context.Background())

		go func() {
			defer close(done)

			sw.runSeedDialLoop(ctx)
		}()

		cancelFn()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("seed dial loop did not exit")
		}
	})
}

func TestMultiplexSwitch_DialLoop_BackedOff(t *testing.T) {
	t.Parallel()

	t.Run("a due item is dialed while an earlier one backs off", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		var (
			mockTransport, dialed = newDialRecorder(1)
			addrs                 = generateNetAddr(t, 2)

			sw = NewMultiplexSwitch(mockTransport)
		)

		sw.peers = &mockSet{
			hasFn: func(types.ID) bool { return false },
		}

		// Park the loop on an item that is not due for a long time
		sw.dialQueue.Push(dial.Item{
			Time:    time.Now().Add(time.Hour),
			Address: addrs[0],
		})

		go sw.runDialLoop(ctx)

		time.Sleep(50 * time.Millisecond)

		// The loop is parked, so it only picks this up if the wait is
		// interruptible by a newly queued item
		sw.DialPeers(addrs[1])

		select {
		case addr := <-dialed:
			assert.Equal(t, *addrs[1], addr)
		case <-time.After(5 * time.Second):
			t.Fatal("a due item was not dialed while an earlier one was backing off")
		}
	})

	t.Run("the wait ends on context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancelFn := context.WithCancel(context.Background())

		var (
			sw   = NewMultiplexSwitch(&mockTransport{})
			done = make(chan struct{})
		)

		go func() {
			defer close(done)

			sw.waitForDialTime(ctx, time.Hour)
		}()

		cancelFn()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("the wait outlived the context")
		}
	})

	t.Run("an item is dialed once its wait elapses", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		var (
			mockTransport, dialed = newDialRecorder(1)
			addr                  = generateNetAddr(t, 1)[0]

			sw = NewMultiplexSwitch(mockTransport)
		)

		sw.peers = &mockSet{
			hasFn: func(types.ID) bool { return false },
		}

		// Nothing else is queued and the context stays live, so the loop can
		// only come back on the timer
		sw.dialQueue.Push(dial.Item{
			Time:    time.Now().Add(100 * time.Millisecond),
			Address: addr,
		})

		go sw.runDialLoop(ctx)

		select {
		case got := <-dialed:
			assert.Equal(t, *addr, got)
		case <-time.After(5 * time.Second):
			t.Fatal("the item was not dialed once it became due")
		}
	})
}

// TestMultiplexSwitch_DialLoop_DoesNotSpin guards against the dial loop busy
// waiting while every queued item is backing off. It is deliberately not
// parallel: it reads the goroutine dump, and parallel tests running their own
// dial loop would be indistinguishable
func TestMultiplexSwitch_DialLoop_DoesNotSpin(t *testing.T) {
	ctx := t.Context()

	sw := NewMultiplexSwitch(&mockTransport{})
	sw.peers = &mockSet{
		hasFn: func(types.ID) bool { return false },
	}

	// The redial loop queues exactly this for a persistent peer in backoff
	sw.persistentDialQueue.Push(dial.Item{
		Time:    time.Now().Add(time.Hour),
		Address: generateNetAddr(t, 1)[0],
	})

	go sw.runDialLoop(ctx)

	// Give the loop time to settle on the backed off item
	time.Sleep(100 * time.Millisecond)

	buf := make([]byte, 64<<10)
	buf = buf[:runtime.Stack(buf, true)]

	for g := range strings.SplitSeq(string(buf), "\n\n") {
		if !strings.Contains(g, "runDialLoop") {
			continue
		}

		// A parked loop sits in a select, a spinning one is running
		assert.Contains(
			t,
			g,
			"[select]",
			"the dial loop must park while every queued item is backing off",
		)

		return
	}

	t.Fatal("no goroutine running runDialLoop found")
}

// TestMultiplexSwitch_DialLoop_PopsOnlyDueItems guards the dial loop's pop
// against taking a head that is not due: the redial loop removes a queued dial
// between the peek and the pop, and the head is then a dial in backoff
func TestMultiplexSwitch_DialLoop_PopsOnlyDueItems(t *testing.T) {
	t.Parallel()

	var (
		addrs = generateNetAddr(t, 2)
		due   = addrs[0]
		later = addrs[1]

		dialedLater atomic.Bool
	)

	sw := NewMultiplexSwitch(&mockTransport{
		dialFn: func(_ context.Context, address types.NetAddress, _ PeerBehavior) (PeerConn, error) {
			if address.ID == later.ID {
				dialedLater.Store(true)
			}

			return nil, errors.New("dial failed")
		},
	})
	sw.peers = &mockSet{
		hasFn: func(types.ID) bool { return false },
	}

	// A dial in backoff, that must never be popped
	sw.persistentDialQueue.Push(dial.Item{
		Time:    time.Now().Add(time.Hour),
		Address: later,
	})

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})

	go func() {
		defer close(done)

		sw.runDialLoop(ctx)
	}()

	// Queue a due dial and remove it again, so that the removal regularly
	// lands between the loop's peek and its pop
	for range 2000 {
		sw.persistentDialQueue.Push(dial.Item{Time: time.Now(), Address: due})
		sw.notifyAddPeerToDial()
		sw.persistentDialQueue.Remove(due)
	}

	// A pop landing at the very end of the churn counts, so read the flag
	// once the loop has exited
	cancel()
	<-done

	assert.False(t, dialedLater.Load(), "a dial that is not due was popped")
}

func TestMultiplexSwitch_PeekDialItem(t *testing.T) {
	t.Parallel()

	var (
		now    = time.Now()
		due    = now.Add(-time.Second)
		later  = now.Add(time.Hour)
		latest = now.Add(2 * time.Hour)
	)

	var (
		persistentQueue = func(sw *MultiplexSwitch) *dial.Queue { return sw.persistentDialQueue }
		generalQueue    = func(sw *MultiplexSwitch) *dial.Queue { return sw.dialQueue }
	)

	testTable := []struct {
		name       string
		persistent []time.Time
		general    []time.Time
		want       func(*MultiplexSwitch) *dial.Queue // the queue the item comes from, nil for none
		wantTime   time.Time
	}{
		{
			name: "both queues empty",
		},
		{
			name:       "a due persistent peer",
			persistent: []time.Time{due},
			want:       persistentQueue,
			wantTime:   due,
		},
		{
			name:     "a due discovered peer",
			general:  []time.Time{due},
			want:     generalQueue,
			wantTime: due,
		},
		{
			name:       "a due persistent peer goes before an earlier discovered peer",
			persistent: []time.Time{due},
			general:    []time.Time{due.Add(-time.Minute)},
			want:       persistentQueue,
			wantTime:   due,
		},
		{
			name:       "a due discovered peer goes while the persistent peer backs off",
			persistent: []time.Time{later},
			general:    []time.Time{due},
			want:       generalQueue,
			wantTime:   due,
		},
		{
			name:       "nothing due, the persistent peer is due first",
			persistent: []time.Time{later},
			general:    []time.Time{latest},
			want:       persistentQueue,
			wantTime:   later,
		},
		{
			name:       "nothing due, the discovered peer is due first",
			persistent: []time.Time{latest},
			general:    []time.Time{later},
			want:       generalQueue,
			wantTime:   later,
		},
		{
			name:       "a backed off persistent peer alone",
			persistent: []time.Time{later},
			want:       persistentQueue,
			wantTime:   later,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				sw    = NewMultiplexSwitch(&mockTransport{})
				addrs = generateNetAddr(t, len(testCase.persistent)+len(testCase.general))
			)

			for i, dialTime := range testCase.persistent {
				sw.persistentDialQueue.Push(dial.Item{Time: dialTime, Address: addrs[i]})
			}

			for i, dialTime := range testCase.general {
				sw.dialQueue.Push(dial.Item{
					Time:    dialTime,
					Address: addrs[len(testCase.persistent)+i],
				})
			}

			item, queue := sw.peekDialItem()

			if testCase.want == nil {
				assert.Nil(t, item)

				return
			}

			require.NotNil(t, item)
			assert.Same(t, testCase.want(sw), queue)
			assert.True(t, item.Time.Equal(testCase.wantTime))
		})
	}
}

func TestMultiplexSwitch_DialLoop_Persistent(t *testing.T) {
	t.Parallel()

	t.Run("a due persistent peer is dialed before due discovered peers", func(t *testing.T) {
		t.Parallel()

		var (
			addrs                 = generateNetAddr(t, 3)
			mockTransport, dialed = newDialRecorder(len(addrs))

			sw  = NewMultiplexSwitch(mockTransport)
			now = time.Now()
		)

		// Discovered peers that have been due for a while
		sw.dialQueue.Push(dial.Item{Time: now.Add(-2 * time.Second), Address: addrs[0]})
		sw.dialQueue.Push(dial.Item{Time: now.Add(-time.Second), Address: addrs[1]})

		// A persistent peer that just became due
		sw.persistentDialQueue.Push(dial.Item{Time: now, Address: addrs[2]})

		go sw.runDialLoop(t.Context())

		for _, want := range []*types.NetAddress{addrs[2], addrs[0], addrs[1]} {
			select {
			case got := <-dialed:
				assert.Equal(t, *want, got)
			case <-time.After(5 * time.Second):
				t.Fatal("the dial loop stalled")
			}
		}
	})

	t.Run("a persistent peer connected meanwhile is not dialed", func(t *testing.T) {
		t.Parallel()

		var (
			addrs                 = generateNetAddr(t, 2)
			mockTransport, dialed = newDialRecorder(len(addrs))

			sw  = NewMultiplexSwitch(mockTransport)
			now = time.Now()
		)

		// The persistent peer connected, inbound, after its dial was queued
		sw.peers = &mockSet{
			hasFn: func(id types.ID) bool { return id == addrs[0].ID },
		}

		sw.persistentDialQueue.Push(dial.Item{Time: now, Address: addrs[0]})
		sw.dialQueue.Push(dial.Item{Time: now, Address: addrs[1]})

		go sw.runDialLoop(t.Context())

		// The persistent item is taken first and dropped, so the only dial is
		// the discovered peer
		select {
		case got := <-dialed:
			assert.Equal(t, *addrs[1], got)
		case <-time.After(5 * time.Second):
			t.Fatal("the dial loop stalled")
		}

		assert.Nil(t, sw.persistentDialQueue.Peek())
	})

	t.Run("a due discovered peer is dialed while the persistent peer backs off", func(t *testing.T) {
		t.Parallel()

		var (
			addrs                 = generateNetAddr(t, 2)
			mockTransport, dialed = newDialRecorder(len(addrs))

			sw  = NewMultiplexSwitch(mockTransport)
			now = time.Now()
		)

		sw.persistentDialQueue.Push(dial.Item{Time: now.Add(time.Hour), Address: addrs[0]})
		sw.dialQueue.Push(dial.Item{Time: now, Address: addrs[1]})

		go sw.runDialLoop(t.Context())

		select {
		case got := <-dialed:
			assert.Equal(t, *addrs[1], got)
		case <-time.After(5 * time.Second):
			t.Fatal("the dial loop stalled")
		}
	})
}

func TestMultiplexSwitch_QueueMissingPersistentPeers(t *testing.T) {
	t.Parallel()

	t.Run("missing peer queued on its configured address, due now", func(t *testing.T) {
		t.Parallel()

		var (
			addr = generateNetAddr(t, 1)[0]
			sw   = NewMultiplexSwitch(
				&mockTransport{},
				WithPersistentPeers([]*types.NetAddress{addr}),
			)
			now = time.Now()
		)

		sw.queueMissingPersistentPeers(make(map[types.ID]uint), now)

		// The dial loop is woken: it only notices new items through dialNotify
		assert.Len(t, sw.dialNotify, 1)

		item := sw.persistentDialQueue.Pop()

		require.NotNil(t, item)
		assert.Equal(t, addr, item.Address)
		assert.True(t, item.Time.Equal(now))
		assert.Nil(t, sw.persistentDialQueue.Pop())
		assert.Nil(t, sw.dialQueue.Peek())
	})

	t.Run("connected peer skipped", func(t *testing.T) {
		t.Parallel()

		addr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithPersistentPeers([]*types.NetAddress{addr}),
		)

		sw.peers = &mockSet{
			hasFn: func(id types.ID) bool { return id == addr.ID },
		}

		sw.queueMissingPersistentPeers(make(map[types.ID]uint), time.Now())

		assert.Nil(t, sw.persistentDialQueue.Peek())
	})

	t.Run("queued peer not duplicated", func(t *testing.T) {
		t.Parallel()

		var (
			addr = generateNetAddr(t, 1)[0]
			sw   = NewMultiplexSwitch(
				&mockTransport{},
				WithPersistentPeers([]*types.NetAddress{addr}),
			)
			attempts = make(map[types.ID]uint)
			now      = time.Now()
		)

		sw.queueMissingPersistentPeers(attempts, now)
		sw.queueMissingPersistentPeers(attempts, now)

		require.NotNil(t, sw.persistentDialQueue.Pop())
		assert.Nil(t, sw.persistentDialQueue.Pop())
	})

	t.Run("queued with no outbound slot", func(t *testing.T) {
		t.Parallel()

		addr := generateNetAddr(t, 1)[0]

		// An operator profile allowing no discovered peers at all
		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithPersistentPeers([]*types.NetAddress{addr}),
			WithMaxOutboundPeers(0),
		)

		sw.queueMissingPersistentPeers(make(map[types.ID]uint), time.Now())

		item := sw.persistentDialQueue.Pop()

		require.NotNil(t, item)
		assert.Equal(t, addr, item.Address)
	})

	t.Run("own address skipped", func(t *testing.T) {
		t.Parallel()

		var (
			addrs = generateNetAddr(t, 2)
			self  = addrs[0]

			// A shared persistent peer list, which includes the node itself
			sw = NewMultiplexSwitch(
				&mockTransport{
					netAddressFn: func() types.NetAddress {
						return *self
					},
				},
				WithPersistentPeers(addrs),
			)
		)

		sw.queueMissingPersistentPeers(make(map[types.ID]uint), time.Now())

		item := sw.persistentDialQueue.Pop()

		require.NotNil(t, item)
		assert.Equal(t, addrs[1], item.Address)
		assert.Nil(t, sw.persistentDialQueue.Pop())
	})

	t.Run("backoff doubles up to the ceiling", func(t *testing.T) {
		t.Parallel()

		var (
			addr = generateNetAddr(t, 1)[0]
			sw   = NewMultiplexSwitch(
				&mockTransport{},
				WithPersistentPeers([]*types.NetAddress{addr}),
			)
			attempts = make(map[types.ID]uint)
			now      = time.Now()
		)

		// With no attempt recorded, the dial is due right away
		sw.queueMissingPersistentPeers(attempts, now)

		item := sw.persistentDialQueue.Pop()

		require.NotNil(t, item)
		assert.True(t, item.Time.Equal(now))

		// Every later dial fails and waits for a doubling backoff, capped at
		// persistentRedialMaxBackoff. Forty attempts run well past the point
		// where an unbounded shift of the base interval overflows
		for attempt := range uint(40) {
			sw.queueMissingPersistentPeers(attempts, now)

			item := sw.persistentDialQueue.Pop()
			require.NotNil(t, item)

			want := persistentRedialMaxBackoff
			if attempt < 5 {
				want = time.Second << attempt
			}

			// The backoff carries a jitter of up to 10%
			got := item.Time.Sub(now)

			assert.GreaterOrEqual(t, got, want-want/10, "attempt %d", attempt)
			assert.LessOrEqual(t, got, want+want/10, "attempt %d", attempt)
		}
	})

	t.Run("a cleared peer is queued once per call, due right away", func(t *testing.T) {
		t.Parallel()

		var (
			addr = generateNetAddr(t, 1)[0]
			sw   = NewMultiplexSwitch(
				&mockTransport{},
				WithPersistentPeers([]*types.NetAddress{addr}),
			)
			attempts = make(map[types.ID]uint)
			now      = time.Now()
		)

		for pass := range 5 {
			passTime := now.Add(time.Duration(pass) * defaultRedialInterval)

			sw.queueMissingPersistentPeers(attempts, passTime)

			// Exactly one dial, due right away
			item := sw.persistentDialQueue.Pop()

			require.NotNil(t, item)
			assert.True(t, item.Time.Equal(passTime))
			assert.Nil(t, sw.persistentDialQueue.Pop())

			// Stands in for the reset a stable connection's drop performs. See
			// TestMultiplexSwitch_PersistentPeerEvents
			delete(attempts, addr.ID)
		}
	})
}

func TestMultiplexSwitch_PersistentPeerEvents(t *testing.T) {
	t.Parallel()

	// popDelay pops the queued dial, and returns how long after from it is due
	popDelay := func(t *testing.T, sw *MultiplexSwitch, from time.Time) time.Duration {
		t.Helper()

		item := sw.persistentDialQueue.Pop()
		require.NotNil(t, item)

		return item.Time.Sub(from)
	}

	// newPersistentSwitch returns a switch with a single persistent peer, and
	// that peer's configured address
	newPersistentSwitch := func(t *testing.T) (*MultiplexSwitch, *types.NetAddress) {
		t.Helper()

		addr := generateNetAddr(t, 1)[0]

		return NewMultiplexSwitch(
			&mockTransport{},
			WithPersistentPeers([]*types.NetAddress{addr}),
		), addr
	}

	// setConnected makes the switch's peer set report the peer as connected,
	// or not, following the returned flag
	setConnected := func(sw *MultiplexSwitch, id types.ID) *bool {
		connected := new(bool)

		sw.peers = &mockSet{
			hasFn: func(peerID types.ID) bool { return *connected && peerID == id },
		}

		return connected
	}

	// assertBackoff asserts a delay is the backoff after the given attempts,
	// within its 10% jitter
	assertBackoff := func(t *testing.T, attempts uint, delay time.Duration) {
		t.Helper()

		want := persistentRedialMaxBackoff
		if attempts < 5 {
			want = time.Second << attempts
		}

		assert.GreaterOrEqual(t, delay, want-want/10, "attempts %d", attempts)
		assert.LessOrEqual(t, delay, want+want/10, "attempts %d", attempts)
	}

	t.Run("a connect records the time and drops the peer's queued dial", func(t *testing.T) {
		t.Parallel()

		var (
			addrs = generateNetAddr(t, 2)
			sw    = NewMultiplexSwitch(
				&mockTransport{},
				WithPersistentPeers(addrs),
			)
			now         = time.Now()
			connectedAt = make(map[types.ID]time.Time)
		)

		// A dial queued while an earlier one was in flight, and another
		// persistent peer's dial
		sw.persistentDialQueue.Push(dial.Item{Time: now.Add(20 * time.Second), Address: addrs[0]})
		sw.persistentDialQueue.Push(dial.Item{Time: now.Add(10 * time.Second), Address: addrs[1]})

		// The peer is connected when the event is handled
		sw.peers = &mockSet{
			hasFn: func(id types.ID) bool { return id == addrs[0].ID },
		}

		sw.persistentPeerConnected(addrs[0].ID, connectedAt, now)

		assert.True(t, connectedAt[addrs[0].ID].Equal(now))

		item := sw.persistentDialQueue.Pop()

		require.NotNil(t, item)
		assert.Equal(t, addrs[1], item.Address)
		assert.Nil(t, sw.persistentDialQueue.Pop())
	})

	t.Run("a connect handled after the peer dropped is ignored", func(t *testing.T) {
		t.Parallel()

		var (
			sw, addr    = newPersistentSwitch(t)
			now         = time.Now()
			connectedAt = make(map[types.ID]time.Time)
		)

		// The peer is not in the peer set, and the disconnect handler already
		// queued its dial
		sw.persistentDialQueue.Push(dial.Item{Time: now.Add(time.Second), Address: addr})

		sw.persistentPeerConnected(addr.ID, connectedAt, now)

		assert.NotContains(t, connectedAt, addr.ID)

		item := sw.persistentDialQueue.Peek()

		require.NotNil(t, item)
		assert.Equal(t, addr, item.Address)
	})

	t.Run("a connect ignored after the peer dropped clears a stale connect time", func(t *testing.T) {
		t.Parallel()

		var (
			sw, addr    = newPersistentSwitch(t)
			now         = time.Now()
			connectedAt = map[types.ID]time.Time{addr.ID: now.Add(-time.Minute)}
		)

		// The stamp is left by a connection whose disconnect was lost, and the
		// peer is not in the peer set, so the connect is ignored
		sw.persistentDialQueue.Push(dial.Item{Time: now.Add(time.Second), Address: addr})

		sw.persistentPeerConnected(addr.ID, connectedAt, now)

		assert.NotContains(t, connectedAt, addr.ID)

		item := sw.persistentDialQueue.Peek()

		require.NotNil(t, item)
		assert.Equal(t, addr, item.Address)
	})

	t.Run("a short connection keeps the backoff", func(t *testing.T) {
		t.Parallel()

		var (
			sw, addr    = newPersistentSwitch(t)
			now         = time.Now()
			attempts    = map[types.ID]uint{addr.ID: 2}
			connectedAt = map[types.ID]time.Time{addr.ID: now}
			dropped     = now.Add(10 * time.Second)
		)

		sw.persistentPeerDisconnected(addr.ID, attempts, connectedAt, dropped)

		assertBackoff(t, 2, popDelay(t, sw, dropped))
		assert.Equal(t, uint(3), attempts[addr.ID])
		assert.NotContains(t, connectedAt, addr.ID)
	})

	t.Run("a stable connection resets the backoff", func(t *testing.T) {
		t.Parallel()

		var (
			sw, addr    = newPersistentSwitch(t)
			now         = time.Now()
			attempts    = map[types.ID]uint{addr.ID: 5}
			connectedAt = map[types.ID]time.Time{addr.ID: now}

			// Exactly the threshold counts as stable
			dropped = now.Add(persistentStableUptime)
		)

		sw.persistentPeerDisconnected(addr.ID, attempts, connectedAt, dropped)

		// The dial loop is woken, and the dial is due at once
		assert.Len(t, sw.dialNotify, 1)
		assert.Equal(t, time.Duration(0), popDelay(t, sw, dropped))
		assert.Equal(t, map[types.ID]uint{addr.ID: 0}, attempts)
	})

	t.Run("a disconnect without a recorded connect keeps the backoff", func(t *testing.T) {
		t.Parallel()

		var (
			sw, addr = newPersistentSwitch(t)
			dropped  = time.Now()
			attempts = map[types.ID]uint{addr.ID: 1}
		)

		sw.persistentPeerDisconnected(addr.ID, attempts, make(map[types.ID]time.Time), dropped)

		assertBackoff(t, 1, popDelay(t, sw, dropped))
	})

	t.Run("a disconnect handled after a reconnect queues nothing", func(t *testing.T) {
		t.Parallel()

		var (
			sw, addr = newPersistentSwitch(t)
			now      = time.Now()
		)

		// The peer is connected again by the time the event is handled
		sw.peers = &mockSet{
			hasFn: func(id types.ID) bool { return id == addr.ID },
		}

		sw.persistentPeerDisconnected(
			addr.ID,
			map[types.ID]uint{addr.ID: 0},
			map[types.ID]time.Time{addr.ID: now},
			now.Add(time.Second),
		)

		assert.Nil(t, sw.persistentDialQueue.Peek())
	})

	t.Run("repeated short connections back off up to the ceiling", func(t *testing.T) {
		t.Parallel()

		var (
			sw, addr = newPersistentSwitch(t)
			now      = time.Now()

			// As after the first dial on start
			attempts    = map[types.ID]uint{addr.ID: 0}
			connectedAt = make(map[types.ID]time.Time)
			connected   = setConnected(sw, addr.ID)
		)

		for cycle := range uint(8) {
			// Each connection is accepted, then dropped a second later
			*connected = true

			sw.persistentPeerConnected(addr.ID, connectedAt, now)

			dropped := now.Add(time.Second)

			*connected = false

			sw.persistentPeerDisconnected(addr.ID, attempts, connectedAt, dropped)

			delay := popDelay(t, sw, dropped)
			assertBackoff(t, cycle, delay)

			// The next connection happens when that dial is due
			now = dropped.Add(delay)
		}
	})

	t.Run("a leftover dial does not delay the redial of a stable connection", func(t *testing.T) {
		t.Parallel()

		var (
			sw, addr    = newPersistentSwitch(t)
			now         = time.Now()
			attempts    = map[types.ID]uint{addr.ID: 4}
			connectedAt = make(map[types.ID]time.Time)
			connected   = setConnected(sw, addr.ID)
		)

		// Queued by a tick while the dial that is about to connect was in flight
		sw.persistentDialQueue.Push(dial.Item{Time: now.Add(16 * time.Second), Address: addr})

		*connected = true

		sw.persistentPeerConnected(addr.ID, connectedAt, now)

		dropped := now.Add(time.Minute)

		*connected = false

		sw.persistentPeerDisconnected(addr.ID, attempts, connectedAt, dropped)

		assert.Equal(t, time.Duration(0), popDelay(t, sw, dropped))
		assert.Nil(t, sw.persistentDialQueue.Peek())
	})
}

func TestMultiplexSwitch_DialPeers(t *testing.T) {
	t.Parallel()

	t.Run("self dial request", func(t *testing.T) {
		t.Parallel()

		var (
			p    = mock.GeneratePeers(t, 1)[0]
			addr = types.NetAddress{
				ID:   "id",
				IP:   p.SocketAddr().IP,
				Port: p.SocketAddr().Port,
			}

			mockTransport = &mockTransport{
				netAddressFn: func() types.NetAddress {
					return addr
				},
			}
		)

		// Make sure the "peer" has the same address
		// as the transport (node)
		p.NodeInfoFn = func() types.NodeInfo {
			return types.NodeInfo{
				NetAddress: &addr,
			}
		}

		sw := NewMultiplexSwitch(mockTransport)

		// Dial the peers
		sw.DialPeers(p.SocketAddr())

		// Make sure the peer wasn't actually dialed
		assert.False(t, sw.dialQueue.Has(p.SocketAddr()))
	})

	t.Run("persistent peer left to the redial loop", func(t *testing.T) {
		t.Parallel()

		var (
			configured = generateNetAddr(t, 1)[0]

			// The same peer, as another node advertises it
			learned = advertisedElsewhere(configured)

			sw = NewMultiplexSwitch(
				&mockTransport{},
				WithPersistentPeers([]*types.NetAddress{configured}),
			)
		)

		sw.DialPeers(configured, learned)

		assert.Nil(t, sw.dialQueue.Peek())
		assert.Nil(t, sw.persistentDialQueue.Peek())
	})

	t.Run("connected peer skipped", func(t *testing.T) {
		t.Parallel()

		addr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(&mockTransport{})
		sw.peers = &mockSet{
			hasFn: func(id types.ID) bool { return id == addr.ID },
		}

		sw.DialPeers(addr)

		assert.Nil(t, sw.dialQueue.Peek())
	})

	t.Run("queued address not duplicated", func(t *testing.T) {
		t.Parallel()

		addr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(&mockTransport{})

		// Peer exchange responses repeat the same addresses
		sw.DialPeers(addr, addr)
		sw.DialPeers(addr)

		require.NotNil(t, sw.dialQueue.Pop())
		assert.Nil(t, sw.dialQueue.Pop())
	})

	t.Run("another address of a queued peer is queued", func(t *testing.T) {
		t.Parallel()

		var (
			addr  = generateNetAddr(t, 1)[0]
			other = advertisedElsewhere(addr)

			sw = NewMultiplexSwitch(&mockTransport{})
		)

		sw.DialPeers(addr, other)

		require.NotNil(t, sw.dialQueue.Pop())
		require.NotNil(t, sw.dialQueue.Pop())
		assert.Nil(t, sw.dialQueue.Pop())
	})

	t.Run("outbound peer limit reached", func(t *testing.T) {
		t.Parallel()

		var (
			maxOutbound = uint64(10)
			peers       = mock.GeneratePeers(t, 10)

			mockTransport = &mockTransport{
				netAddressFn: func() types.NetAddress {
					return types.NetAddress{
						ID: "id",
						IP: net.IP{},
					}
				},
			}

			ps = &mockSet{
				numOutboundFn: func() uint64 {
					return maxOutbound
				},
			}
		)

		sw := NewMultiplexSwitch(
			mockTransport,
			WithMaxOutboundPeers(maxOutbound),
		)

		// Set the peer set
		sw.peers = ps

		// Dial the peers
		addrs := make([]*types.NetAddress, 0, len(peers))

		for _, p := range peers {
			addrs = append(addrs, p.SocketAddr())
		}

		sw.DialPeers(addrs...)

		// Make sure no peers were dialed
		for _, p := range peers {
			assert.False(t, sw.dialQueue.Has(p.SocketAddr()))
		}
	})

	t.Run("peers dialed", func(t *testing.T) {
		t.Parallel()

		var (
			maxOutbound = uint64(1000)
			peers       = mock.GeneratePeers(t, int(maxOutbound/2))

			mockTransport = &mockTransport{
				netAddressFn: func() types.NetAddress {
					return types.NetAddress{
						ID: "id",
						IP: net.IP{},
					}
				},
			}
		)

		sw := NewMultiplexSwitch(
			mockTransport,
			WithMaxOutboundPeers(10),
		)

		// Dial the peers
		addrs := make([]*types.NetAddress, 0, len(peers))

		for _, p := range peers {
			addrs = append(addrs, p.SocketAddr())
		}

		sw.DialPeers(addrs...)

		// Make sure peers were dialed
		for _, p := range peers {
			assert.True(t, sw.dialQueue.Has(p.SocketAddr()))
		}
	})
}

// TestMultiplexSwitch_PersistentPeerDialedOnConfiguredAddress is a switch-level
// regression test for gnolang/gno#6287: a persistent peer reached through an
// address learned via peer exchange drops, and the same address keeps coming
// back through peer exchange. It must only ever be dialed on its configured
// address. It does not reproduce the pop-time drop itself
func TestMultiplexSwitch_PersistentPeerDialedOnConfiguredAddress(t *testing.T) {
	t.Parallel()

	var (
		configured = generateNetAddr(t, 1)[0]

		// The same peer, as its public external address
		learned = advertisedElsewhere(configured)

		mockTransport, dialed = newDialRecorder(16)

		sw = NewMultiplexSwitch(
			mockTransport,
			WithPersistentPeers([]*types.NetAddress{configured}),
		)

		p = mock.GeneratePeers(t, 1)[0]
	)

	// The connection over the learned address drops
	p.IDFn = func() types.ID { return configured.ID }
	p.SocketAddrFn = func() *types.NetAddress { return learned }
	p.IsOutboundFn = func() bool { return true }

	sw.StopPeerForError(p, errors.New("EOF"))

	// Peer exchange shares the learned address again
	sw.DialPeers(learned)

	// Neither the dropped connection nor peer exchange queued the learned address
	assert.Nil(t, sw.dialQueue.Peek())

	ctx := t.Context()

	go sw.runDialLoop(ctx)
	go sw.runRedialLoop(ctx)

	// The redial loop dials the configured address
	select {
	case addr := <-dialed:
		assert.Equal(t, *configured, addr)
	case <-time.After(5 * time.Second):
		t.Fatal("the persistent peer was not dialed")
	}
}

func TestCalculateBackoff(t *testing.T) {
	t.Parallel()

	checkJitterRange := func(t *testing.T, expectedAbs, actual time.Duration) {
		t.Helper()
		require.LessOrEqual(t, actual, expectedAbs)
		require.GreaterOrEqual(t, actual, expectedAbs*-1)
	}

	// Test that the default jitter factor is 10% of the backoff duration.
	t.Run("percentage jitter", func(t *testing.T) {
		t.Parallel()

		for range 1000 {
			checkJitterRange(t, 100*time.Millisecond, calculateBackoff(0, time.Second, 10*time.Minute)-time.Second)
			checkJitterRange(t, 200*time.Millisecond, calculateBackoff(1, time.Second, 10*time.Minute)-2*time.Second)
			checkJitterRange(t, 400*time.Millisecond, calculateBackoff(2, time.Second, 10*time.Minute)-4*time.Second)
			checkJitterRange(t, 800*time.Millisecond, calculateBackoff(3, time.Second, 10*time.Minute)-8*time.Second)
			checkJitterRange(t, 1600*time.Millisecond, calculateBackoff(4, time.Second, 10*time.Minute)-16*time.Second)
		}
	})

	// Test that the jitter factor is capped at 10 sec.
	t.Run("capped jitter", func(t *testing.T) {
		t.Parallel()

		for range 1000 {
			checkJitterRange(t, 10*time.Second, calculateBackoff(7, time.Second, 10*time.Minute)-128*time.Second)
			checkJitterRange(t, 10*time.Second, calculateBackoff(10, time.Second, 20*time.Minute)-1024*time.Second)
			checkJitterRange(t, 10*time.Second, calculateBackoff(20, time.Second, 300*time.Hour)-1048576*time.Second)
		}
	})

	// Test that the backoff interval is based on the baseInterval.
	t.Run("base interval", func(t *testing.T) {
		t.Parallel()

		for range 1000 {
			checkJitterRange(t, 4800*time.Millisecond, calculateBackoff(4, 3*time.Second, 10*time.Minute)-48*time.Second)
			checkJitterRange(t, 8*time.Second, calculateBackoff(3, 10*time.Second, 10*time.Minute)-80*time.Second)
			checkJitterRange(t, 10*time.Second, calculateBackoff(5, 3*time.Hour, 100*time.Hour)-96*time.Hour)
		}
	})

	// Test that the backoff interval is capped at maxInterval +/- jitter factor.
	t.Run("max interval", func(t *testing.T) {
		t.Parallel()

		for range 1000 {
			checkJitterRange(t, 100*time.Millisecond, calculateBackoff(10, 10*time.Hour, time.Second)-time.Second)
			checkJitterRange(t, 1600*time.Millisecond, calculateBackoff(10, 10*time.Hour, 16*time.Second)-16*time.Second)
			checkJitterRange(t, 10*time.Second, calculateBackoff(10, 10*time.Hour, 128*time.Second)-128*time.Second)
		}
	})

	// Test parameters sanitization for base and max intervals.
	t.Run("parameters sanitization", func(t *testing.T) {
		t.Parallel()

		for range 1000 {
			checkJitterRange(t, 100*time.Millisecond, calculateBackoff(0, -10, -10)-time.Second)
			checkJitterRange(t, 1600*time.Millisecond, calculateBackoff(4, -10, -10)-16*time.Second)
			checkJitterRange(t, 10*time.Second, calculateBackoff(7, -10, 10*time.Minute)-128*time.Second)
		}
	})

	// Test that the backoff interval stays capped however many attempts were made.
	t.Run("attempts overflow", func(t *testing.T) {
		t.Parallel()

		for _, attempts := range []uint{33, 34, 40, 50, 63, 64, 100} {
			for range 100 {
				checkJitterRange(t, 3*time.Second, calculateBackoff(attempts, time.Second, 30*time.Second)-30*time.Second)
			}
		}
	})
}

func TestSwitchAcceptLoopTransportClosed(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	var transportClosed bool
	mockTransport := &mockTransport{
		acceptFn: func(context.Context, PeerBehavior) (PeerConn, error) {
			transportClosed = true
			return nil, errTransportClosed
		},
	}

	sw := NewMultiplexSwitch(mockTransport)

	// Run the accept loop
	done := make(chan struct{})
	go func() {
		sw.runAcceptLoop(ctx)
		close(done) // signal that accept loop as ended
	}()

	select {
	case <-time.After(time.Second * 2):
		require.FailNow(t, "timeout while waiting for running loop to stop")
	case <-done:
		assert.True(t, transportClosed)
	}
}
