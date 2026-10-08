package p2p

import (
	"bytes"
	"context"
	"log/slog"
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

	t.Run("a peer stopped twice is torn down once", func(t *testing.T) {
		t.Parallel()

		var (
			removed = make(map[string]int)

			countingReactor = func(name string) *mockReactor {
				return &mockReactor{
					removePeerFn: func(PeerConn, any) { removed[name]++ },
				}
			}

			sw = NewMultiplexSwitch(
				&mockTransport{removeFn: func(PeerConn) {}},
				WithReactor("first", countingReactor("first")),
				WithReactor("second", countingReactor("second")),
			)

			p = mock.GeneratePeers(t, 1)[0]
		)

		withRealStop(p)

		evCh, unsubFn := sw.Subscribe(func(events.Event) bool { return true })
		defer unsubFn()

		require.NoError(t, sw.addPeer(p))

		// Two errors on one connection, such as a reactor's and then its recv
		// routine's, tear it down once: the first stop owns the teardown
		sw.StopPeerForError(p, errors.New("peer error"))
		sw.StopPeerForError(p, errors.New("EOF"))

		assert.Equal(t, map[string]int{"first": 1, "second": 1}, removed)
		assert.False(t, sw.peers.Has(p.ID()))

		_, disconnected := countPeerEvents(evCh)
		assert.Equal(t, 1, disconnected)
	})

	t.Run("a peer the set never held announces no disconnect", func(t *testing.T) {
		t.Parallel()

		var (
			sw = NewMultiplexSwitch(&mockTransport{removeFn: func(PeerConn) {}})
			p  = mock.GeneratePeers(t, 1)[0]
		)

		withRealStop(p)
		require.NoError(t, p.Start())

		evCh, unsubFn := sw.Subscribe(func(events.Event) bool { return true })
		defer unsubFn()

		// A started connection the peer set does not hold, such as one refused
		// at registration, reports an error. Its peer was never announced
		sw.StopPeerForError(p, errors.New("EOF"))

		assert.False(t, p.IsRunning())

		_, disconnected := countPeerEvents(evCh)
		assert.Zero(t, disconnected)
	})

	t.Run("a socket closed by the stop is not logged as an error", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name      string
			closeErr  error
			wantError bool
		}{
			{
				"already closed by the stop",
				&net.OpError{Op: "close", Err: net.ErrClosed},
				false,
			},
			{
				"any other close error",
				errors.New("close failed"),
				true,
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				var (
					sw = NewMultiplexSwitch(&mockTransport{removeFn: func(PeerConn) {}})
					p  = mock.GeneratePeers(t, 1)[0]
				)

				p.CloseConnFn = func() error { return testCase.closeErr }
				withRealStop(p)

				logs := captureLogs(sw)

				require.NoError(t, sw.addPeer(p))

				sw.StopPeerForError(p, errors.New("peer error"))

				const closeFailure = "unable to gracefully close peer connection"

				if testCase.wantError {
					assert.Contains(t, logs.String(), closeFailure)

					return
				}

				assert.NotContains(t, logs.String(), closeFailure)
			})
		}
	})
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

// peerReplacedBeforeRunCheck runs replaceFn the first time its running state
// is checked, then reports itself stopped. addPeer's rollback check is the
// first IsRunning call addPeer makes on a peer, so this replaces the peer
// between its registration and that check.
type peerReplacedBeforeRunCheck struct {
	*mock.Peer

	checked   bool
	replaceFn func()
}

func (p *peerReplacedBeforeRunCheck) IsRunning() bool {
	if p.checked {
		return p.Peer.IsRunning()
	}

	p.checked = true
	p.replaceFn()

	return false
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

	// An error on the peer's first read reaches the switch from inside Start.
	p.startFn = func() error {
		sw.StopPeerForError(p, errors.New("peer error"))

		return nil
	}

	evCh, unsubFn := sw.Subscribe(func(events.Event) bool { return true })
	defer unsubFn()

	// addPeer must refuse a peer that was stopped while it was being added.
	// stopAndRemovePeer removes from the peer set last, so its Remove runs
	// before the Add, and nothing would ever remove the peer again
	require.ErrorIs(t, sw.addPeer(p), errPeerStopped)

	// The peer was never announced, so neither is its disconnect
	connected, disconnected := countPeerEvents(evCh)

	assert.Zero(t, connected)
	assert.Zero(t, disconnected)

	// InitPeer is the hook that runs first, so it is the one that can pair
	// with the RemovePeer that frees what it took. AddPeer never runs
	assert.Equal(t, []string{"InitPeer", "RemovePeer"}, calls)

	// The peer holds neither a slot nor its ID in the peer set
	assert.False(t, sw.peers.Has(p.ID()))
	assert.Zero(t, sw.peers.NumInbound())
	assert.Empty(t, sw.peers.List())
	assert.False(t, p.IsRunning())
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
	assert.False(t, dup.IsRunning())

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

	withRealStop(p)

	require.ErrorIs(t, sw.addPeer(p), errDuplicatePeer)

	// Whatever InitPeer took is given back, so a refused connection does not
	// hold reactor state for the lifetime of the process
	assert.Equal(t, []string{"InitPeer", "RemovePeer"}, calls)

	// The refused connection is stopped, so its caller only releases it
	assert.False(t, p.IsRunning())
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
		assert.False(t, p.IsRunning())
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

// dialRejected dials a peer the switch refuses, and returns how many times
// its connection was closed
func dialRejected(t *testing.T, sw *MultiplexSwitch, p *mock.Peer) int {
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

func TestMultiplexSwitch_DialPeer_Rejected(t *testing.T) {
	t.Parallel()

	t.Run("outbound limit reached", func(t *testing.T) {
		t.Parallel()

		sw := NewMultiplexSwitch(nil, WithMaxOutboundPeers(0))
		logs := captureLogs(sw)

		p := mock.GeneratePeers(t, 1)[0]

		assert.Equal(t, 1, dialRejected(t, sw, p))
		assert.False(t, sw.Peers().Has(p.ID()))
		assert.Contains(t, logs.String(), `level=ERROR msg="unable to add peer"`)
	})

	t.Run("duplicate peer", func(t *testing.T) {
		t.Parallel()

		p := mock.GeneratePeers(t, 1)[0]

		// A connection in the same direction is already registered, so the
		// tie-break keeps it and refuses the dialed one
		registered := peerWithID(t, p.ID(), true)

		sw := NewMultiplexSwitch(nil)
		sw.peers = &mockSet{
			hasFn: func(id types.ID) bool { return id == p.ID() },
			getFn: func(id types.ID) PeerConn {
				if id == p.ID() {
					return registered
				}

				return nil
			},
		}

		logs := captureLogs(sw)

		assert.Equal(t, 1, dialRejected(t, sw, p))

		// A refused duplicate logs at Info, since on the dial side it is the
		// tie-break's designed outcome on one node of every simultaneous open
		assert.Contains(t, logs.String(), `level=INFO msg="unable to add peer"`)
	})

	t.Run("peer stopped while being added", func(t *testing.T) {
		t.Parallel()

		p := mock.GeneratePeers(t, 1)[0]
		withRealStop(p)

		// A teardown of the peer, such as after the remote closed it, stops it
		// while it registers
		sw := NewMultiplexSwitch(nil)
		sw.peers = &mockSet{
			addFn: func(PeerConn) error {
				require.NoError(t, p.Stop())

				return nil
			},
		}

		logs := captureLogs(sw)

		assert.Equal(t, 1, dialRejected(t, sw, p))

		// Whatever stopped the peer logged why, and on the dial side a
		// connection replaced or closed by the remote while being added is a
		// designed outcome of a simultaneous open
		assert.Contains(t, logs.String(), `level=INFO msg="unable to add peer"`)
		assert.Contains(t, logs.String(), errPeerStopped.Error())
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

	t.Run("a reconnect clears the backoff", func(t *testing.T) {
		t.Parallel()

		addr := generateNetAddr(t, 1)[0]

		sw := NewMultiplexSwitch(
			&mockTransport{},
			WithPersistentPeers([]*types.NetAddress{addr}),
		)
		sw.redialInterval = 10 * time.Millisecond

		go sw.runRedialLoop(t.Context())

		// popped pops the queued dial, and reports whether it was due right
		// away (due) or scheduled after a backoff (!due)
		popped := func(due bool) func() bool {
			return func() bool {
				item := sw.persistentDialQueue.Pop()

				return item != nil && !item.Time.After(time.Now()) == due
			}
		}

		// The first pass queues a dial due right away
		require.Eventually(t, popped(true), 5*time.Second, 5*time.Millisecond)

		// That dial went nowhere, so the next one waits for the backoff
		require.Eventually(t, popped(false), 5*time.Second, 5*time.Millisecond)

		// The peer connects: once it drops, its next dial is due right away
		sw.events.Notify(events.PeerConnectedEvent{PeerID: addr.ID})

		require.Eventually(t, popped(true), 5*time.Second, 5*time.Millisecond)
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

		// The first dial after a disconnect is due right away
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

			// The dial connects and the connection drops at once, and the
			// redial loop clears the attempts on PeerConnected. The rate limit
			// comes from the loop calling this method once per tick, pinned at
			// loop level by "a reconnect clears the backoff"
			delete(attempts, addr.ID)
		}
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

// orderedIDs returns two distinct node IDs, the lower one first
func orderedIDs(t *testing.T) (types.ID, types.ID) {
	t.Helper()

	a, b := types.GenerateNodeKey().ID(), types.GenerateNodeKey().ID()
	require.NotEqual(t, a, b)

	if b < a {
		a, b = b, a
	}

	return a, b
}

// peerWithID returns a mock peer carrying the given ID, in the given direction
func peerWithID(t *testing.T, id types.ID, outbound bool) *mock.Peer {
	t.Helper()

	p := mock.GeneratePeers(t, 1)[0]
	p.IDFn = func() types.ID { return id }
	p.IsOutboundFn = func() bool { return outbound }

	return p
}

// switchWithID returns a switch whose own node ID is the given one
func switchWithID(id types.ID, opts ...SwitchOption) *MultiplexSwitch {
	return NewMultiplexSwitch(
		&mockTransport{
			netAddressFn: func() types.NetAddress {
				return types.NetAddress{ID: id}
			},
		},
		opts...,
	)
}

// withRealStop gives a mock peer the Stop of the service it embeds: once
// started, the first Stop stops it, so IsRunning reports false, and every later
// one returns service.ErrAlreadyStopped
func withRealStop(p *mock.Peer) {
	p.StopFn = p.BaseService.Stop
}

// drainEvents returns the events already delivered to evCh
func drainEvents(evCh <-chan events.Event) []events.Event {
	var drained []events.Event

	for {
		select {
		case ev := <-evCh:
			drained = append(drained, ev)
		default:
			return drained
		}
	}
}

// countPeerEvents drains the events already delivered to evCh, and counts the
// peer connections and disconnections among them
func countPeerEvents(evCh <-chan events.Event) (connected, disconnected int) {
	for _, ev := range drainEvents(evCh) {
		switch ev.Type() {
		case events.PeerConnected:
			connected++
		case events.PeerDisconnected:
			disconnected++
		}
	}

	return connected, disconnected
}

// acceptOnce returns an accept function that hands out p once, then blocks
// until the context ends
func acceptOnce(p PeerConn) func(context.Context, PeerBehavior) (PeerConn, error) {
	incoming := make(chan PeerConn, 1)
	incoming <- p

	return func(ctx context.Context, _ PeerBehavior) (PeerConn, error) {
		select {
		case p := <-incoming:
			return p, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// acceptSwitch returns a switch whose own node ID is the given one, and whose
// transport accepts the incoming connection once
func acceptSwitch(id types.ID, incoming PeerConn, opts ...SwitchOption) *MultiplexSwitch {
	return NewMultiplexSwitch(
		&mockTransport{
			netAddressFn: func() types.NetAddress {
				return types.NetAddress{ID: id}
			},
			acceptFn: acceptOnce(incoming),
		},
		opts...,
	)
}

// closedSignal returns a channel closed the first time p's connection is
// closed
func closedSignal(p *mock.Peer) <-chan struct{} {
	var (
		closed = make(chan struct{})
		once   sync.Once
	)

	p.CloseConnFn = func() error {
		once.Do(func() { close(closed) })

		return nil
	}

	return closed
}

// awaitClosed fails the test with msg if closed is not closed in time
func awaitClosed(t *testing.T, closed <-chan struct{}, msg string) {
	t.Helper()

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-closed:
	case <-timer.C:
		t.Fatal(msg)
	}
}

func TestMultiplexSwitch_KeepsRegistered(t *testing.T) {
	t.Parallel()

	lower, upper := orderedIDs(t)

	testTable := []struct {
		name               string
		local, remote      types.ID
		registeredOutbound bool
		incomingOutbound   bool
		wantKept           bool
	}{
		{"same direction, both inbound", lower, upper, false, false, true},
		{"same direction, both outbound", lower, upper, true, true, true},
		{"our ID lower, registered outbound", lower, upper, true, false, true},
		{"our ID lower, registered inbound", lower, upper, false, true, false},
		{"our ID higher, registered inbound", upper, lower, false, true, true},
		{"our ID higher, registered outbound", upper, lower, true, false, false},
		{"equal IDs", lower, lower, true, false, true},
		{"own ID unknown", "", upper, false, true, true},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				sw         = switchWithID(testCase.local)
				registered = peerWithID(t, testCase.remote, testCase.registeredOutbound)
				incoming   = peerWithID(t, testCase.remote, testCase.incomingOutbound)
			)

			assert.Equal(t, testCase.wantKept, sw.keepsRegistered(registered, incoming))
		})
	}
}

func TestMultiplexSwitch_RegisterPeer(t *testing.T) {
	t.Parallel()

	t.Run("a peer with no registered connection is added", func(t *testing.T) {
		t.Parallel()

		lower, upper := orderedIDs(t)

		var (
			sw = switchWithID(upper)
			p  = peerWithID(t, lower, false)
		)

		replaced, err := sw.registerPeer(p)

		require.NoError(t, err)
		assert.Nil(t, replaced)
		assert.Same(t, p, sw.peers.Get(lower))
	})

	t.Run("a connection the tie-break refuses leaves the registered one", func(t *testing.T) {
		t.Parallel()

		lower, upper := orderedIDs(t)

		var (
			// Our ID is the lower one, so our outbound connection is kept
			sw       = switchWithID(lower)
			ours     = peerWithID(t, upper, true)
			incoming = peerWithID(t, upper, false)
		)

		require.NoError(t, sw.peers.Add(ours))

		replaced, err := sw.registerPeer(incoming)

		require.ErrorIs(t, err, errDuplicatePeer)
		assert.Nil(t, replaced)
		assert.Same(t, ours, sw.peers.Get(upper))
		assert.EqualValues(t, 1, sw.peers.NumOutbound())
		assert.EqualValues(t, 0, sw.peers.NumInbound())
	})

	t.Run("a connection the tie-break keeps replaces the registered one", func(t *testing.T) {
		t.Parallel()

		lower, upper := orderedIDs(t)

		var (
			// Our ID is the higher one, so the peer's connection is kept
			sw     = switchWithID(upper)
			ours   = peerWithID(t, lower, true)
			theirs = peerWithID(t, lower, false)
		)

		require.NoError(t, sw.peers.Add(ours))

		replaced, err := sw.registerPeer(theirs)

		require.NoError(t, err)
		assert.Same(t, ours, replaced)
		assert.Same(t, theirs, sw.peers.Get(lower))
		assert.EqualValues(t, 0, sw.peers.NumOutbound())
		assert.EqualValues(t, 1, sw.peers.NumInbound())
	})
}

// requireRollbackAnnouncesOneDisconnect has a connection of a peer registered,
// then a replacing connection of the same peer fail to be added, and requires
// the peer to be gone with one connect and one disconnect announced for it.
// wire sets up how the remote's close of the replacing connection lands, and
// returns the connection to add and a function to run once its add returned
func requireRollbackAnnouncesOneDisconnect(
	t *testing.T,
	wire func(sw *MultiplexSwitch, ours, theirs *mock.Peer) (replacing PeerConn, settle func()),
) {
	t.Helper()

	lower, upper := orderedIDs(t)

	var (
		// Our ID is the higher one, so the peer's connection is kept
		sw     = switchWithID(upper)
		ours   = peerWithID(t, lower, true)
		theirs = peerWithID(t, lower, false)
	)

	withRealStop(ours)
	withRealStop(theirs)

	replacing, settle := wire(sw, ours, theirs)

	evCh, unsubFn := sw.Subscribe(func(events.Event) bool { return true })
	defer unsubFn()

	require.NoError(t, sw.addPeer(ours))
	require.ErrorIs(t, sw.addPeer(replacing), errPeerStopped)

	if settle != nil {
		settle()
	}

	assert.False(t, sw.peers.Has(lower))

	var connected, disconnected int

	for _, ev := range drainEvents(evCh) {
		switch ev := ev.(type) {
		case events.PeerConnectedEvent:
			connected++
		case events.PeerDisconnectedEvent:
			disconnected++

			assert.Equal(t, lower, ev.PeerID, "the disconnect names the peer")
		}
	}

	assert.Equal(t, 1, connected)
	assert.Equal(t, 1, disconnected)
}

func TestMultiplexSwitch_AddPeerSimultaneousOpen(t *testing.T) {
	t.Parallel()

	t.Run("the replaced connection is torn down before the new one reaches reactors", func(t *testing.T) {
		t.Parallel()

		lower, upper := orderedIDs(t)

		var (
			calls []string

			// Our ID is the higher one, so the peer's connection is kept
			ours   = peerWithID(t, lower, true)
			theirs = peerWithID(t, lower, false)

			label = func(p PeerConn) string {
				if p == PeerConn(ours) {
					return "ours"
				}

				return "theirs"
			}

			sw           *MultiplexSwitch
			oursClosed   bool
			oursReported bool
			oursReason   error
		)

		ours.CloseConnFn = func() error {
			oursClosed = true

			// A recv routine still running reports the closed socket, once,
			// as MConnection.stopForError does
			if ours.IsRunning() && !oursReported {
				oursReported = true
				sw.StopPeerForError(ours, errors.New("EOF"))
			}

			return nil
		}
		withRealStop(ours)

		reactor := &mockReactor{
			initPeerFn: func(p PeerConn) PeerConn {
				calls = append(calls, "InitPeer "+label(p))

				return p
			},
			addPeerFn: func(p PeerConn) { calls = append(calls, "AddPeer "+label(p)) },
			removePeerFn: func(p PeerConn, reason any) {
				calls = append(calls, "RemovePeer "+label(p))

				if p == PeerConn(ours) {
					oursReason, _ = reason.(error)
				}
			},
		}

		sw = switchWithID(upper, WithReactor("mock", reactor))

		logs := captureLogs(sw)

		evCh, unsubFn := sw.Subscribe(func(events.Event) bool { return true })
		defer unsubFn()

		require.NoError(t, sw.addPeer(ours))
		require.NoError(t, sw.addPeer(theirs))

		wantCalls := []string{
			"InitPeer ours",
			"AddPeer ours",
			"InitPeer theirs",
			"RemovePeer ours",
			"AddPeer theirs",
		}

		assert.Equal(t, wantCalls, calls)

		assert.True(t, oursClosed)
		assert.False(t, ours.IsRunning())
		assert.Contains(t, logs.String(), "replacing connection to resolve a simultaneous open")
		assert.Contains(t, logs.String(), "kept=inbound")
		assert.Same(t, theirs, sw.peers.Get(lower))

		// Each connection announced itself once; the replaced one announces
		// no disconnect, since the peer stays connected
		connected, disconnected := countPeerEvents(evCh)

		assert.Equal(t, 2, connected)
		assert.Zero(t, disconnected)

		// A late error on the replaced connection, such as its recv routine
		// reporting the closed socket, leaves the new entry alone and tears
		// nothing down a second time: the replacement owns the teardown
		sw.StopPeerForError(ours, errors.New("EOF"))

		assert.Same(t, theirs, sw.peers.Get(lower))
		assert.Equal(t, wantCalls, calls)
		assert.ErrorIs(t, oursReason, errSimultaneousOpen)

		connected, disconnected = countPeerEvents(evCh)

		assert.Zero(t, connected)
		assert.Zero(t, disconnected)
	})

	outboundLimitTable := []struct {
		name       string
		persistent bool
		wantErr    bool
	}{
		{"a replacing outbound connection is refused at the outbound limit", false, true},
		{"a persistent peer's replacing outbound connection passes the outbound limit", true, false},
	}

	for _, testCase := range outboundLimitTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			lower, upper := orderedIDs(t)

			// Our ID is the lower one, so our outbound connection would be
			// kept, but no outbound slot is left
			opts := []SwitchOption{WithMaxOutboundPeers(0)}
			if testCase.persistent {
				opts = append(opts, WithPersistentPeers([]*types.NetAddress{{ID: upper}}))
			}

			var (
				sw     = switchWithID(lower, opts...)
				theirs = peerWithID(t, upper, false)
				ours   = peerWithID(t, upper, true)
			)

			require.NoError(t, sw.peers.Add(theirs))

			if testCase.wantErr {
				require.ErrorIs(t, sw.addPeer(ours), errMaxOutboundPeers)
				assert.Same(t, theirs, sw.peers.Get(upper))

				return
			}

			require.NoError(t, sw.addPeer(ours))
			assert.Same(t, ours, sw.peers.Get(upper))
		})
	}

	t.Run("a replacement that fails to start leaves the registered connection", func(t *testing.T) {
		t.Parallel()

		lower, upper := orderedIDs(t)

		var (
			sw     = switchWithID(upper)
			ours   = peerWithID(t, lower, true)
			theirs = peerWithID(t, lower, false)
		)

		require.NoError(t, sw.peers.Add(ours))

		// Starting an already started peer fails
		require.NoError(t, theirs.Start())

		require.Error(t, sw.addPeer(theirs))
		assert.Same(t, ours, sw.peers.Get(lower))
	})

	t.Run("a replacing connection stopped before its registration announces one disconnect", func(t *testing.T) {
		t.Parallel()

		requireRollbackAnnouncesOneDisconnect(t, func(sw *MultiplexSwitch, _, theirs *mock.Peer) (PeerConn, func()) {
			replacing := &peerErrorOnStart{Peer: theirs}

			// The remote closes theirs as soon as it starts, so its teardown
			// runs while ours still holds the entry
			replacing.startFn = func() error {
				if err := theirs.Start(); err != nil {
					return err
				}

				sw.StopPeerForError(replacing, errors.New("EOF"))

				return nil
			}

			return replacing, nil
		})
	})

	t.Run("a replacing connection whose teardown removes it before the rollback announces one disconnect", func(t *testing.T) {
		t.Parallel()

		requireRollbackAnnouncesOneDisconnect(t, func(sw *MultiplexSwitch, ours, theirs *mock.Peer) (PeerConn, func()) {
			// The remote closes theirs while addPeer tears ours down, after
			// theirs registered and before the rollback check. Its teardown
			// removes the entry theirs holds before ours's teardown looks at it
			ours.CloseConnFn = func() error {
				sw.StopPeerForError(theirs, errors.New("EOF"))

				return nil
			}

			return theirs, nil
		})
	})

	t.Run("a replacing connection the rollback removes before its teardown does announces one disconnect", func(t *testing.T) {
		t.Parallel()

		requireRollbackAnnouncesOneDisconnect(t, func(sw *MultiplexSwitch, ours, theirs *mock.Peer) (PeerConn, func()) {
			var (
				theirsStopped = make(chan struct{})
				release       = make(chan struct{})
				tornDown      = make(chan struct{})
				releaseOnce   sync.Once
			)

			// The teardown of theirs stays blocked until release closes, so a
			// failed requirement must not leave its goroutine behind
			releaseTeardown := func() { releaseOnce.Do(func() { close(release) }) }
			t.Cleanup(releaseTeardown)

			// The teardown of theirs stops it, then holds on its socket close
			// until the rollback has run
			theirs.CloseConnFn = func() error {
				close(theirsStopped)
				<-release

				return nil
			}

			// The remote closes theirs while addPeer tears ours down, after
			// theirs registered. Its teardown runs on its own goroutine, as a
			// recv routine's report does, and ours's teardown resumes once
			// theirs is stopped
			ours.CloseConnFn = func() error {
				go func() {
					defer close(tornDown)

					sw.StopPeerForError(theirs, errors.New("EOF"))
				}()

				awaitClosed(t, theirsStopped, "the teardown of theirs did not stop it")

				return nil
			}

			return theirs, func() {
				releaseTeardown()
				awaitClosed(t, tornDown, "the teardown of theirs did not finish")
			}
		})
	})

	t.Run("the rollback leaves the entry of a connection that replaced the stopped one", func(t *testing.T) {
		t.Parallel()

		lower, upper := orderedIDs(t)

		var (
			// Our ID is the higher one, so the peer's connection replaces ours
			sw     = switchWithID(upper)
			ours   = &peerReplacedBeforeRunCheck{Peer: peerWithID(t, lower, true)}
			theirs = peerWithID(t, lower, false)
		)

		// theirs replaces ours, and stops it, between the registration of ours
		// and the rollback check
		ours.replaceFn = func() {
			assert.Same(t, ours, sw.peers.Get(lower), "ours is checked after its registration")
			assert.NoError(t, sw.addPeer(theirs))
		}

		require.ErrorIs(t, sw.addPeer(ours), errPeerStopped)

		// The rollback of ours removes only an entry ours holds
		require.True(t, sw.peers.Has(lower), "the rollback of ours removed the entry of theirs")
		assert.Same(t, theirs, sw.peers.Get(lower))
	})

	t.Run("a concurrent teardown of the replaced connection keeps the new entry", func(t *testing.T) {
		t.Parallel()

		lower, upper := orderedIDs(t)

		for range 200 {
			var (
				sw     = switchWithID(upper)
				ours   = peerWithID(t, lower, true)
				theirs = peerWithID(t, lower, false)
				wg     sync.WaitGroup
			)

			require.NoError(t, sw.peers.Add(ours))

			wg.Go(func() { sw.StopPeerForError(ours, errors.New("EOF")) })
			wg.Go(func() { assert.NoError(t, sw.addPeer(theirs)) })
			wg.Wait()

			assert.Same(t, theirs, sw.peers.Get(lower))
		}
	})
}

func TestMultiplexSwitch_HasPeerFromIP(t *testing.T) {
	t.Parallel()

	var (
		lower, upper = orderedIDs(t)
		ip           = net.ParseIP("127.0.0.1")
		otherIP      = net.ParseIP("127.0.0.2")
	)

	// peerAt returns a peer with the given ID, connected from the given IP
	peerAt := func(id types.ID, ip net.IP) *mock.Peer {
		p := peerWithID(t, id, false)
		p.RemoteIPFn = func() net.IP { return ip }

		return p
	}

	testTable := []struct {
		name   string
		peers  []*mock.Peer
		ip     net.IP
		except types.ID
		want   bool
	}{
		{"a nil IP is held by no peer", []*mock.Peer{peerAt(upper, ip)}, nil, lower, false},
		{"a different peer on the IP holds it", []*mock.Peer{peerAt(upper, ip)}, ip, lower, true},
		{"the excepted peer's own connection does not hold it", []*mock.Peer{peerAt(lower, ip)}, ip, lower, false},
		{
			"a different peer holds it beside the excepted peer's connection",
			[]*mock.Peer{peerAt(lower, ip), peerAt(upper, ip)},
			ip, lower, true,
		},
		{"a different peer on another IP does not hold it", []*mock.Peer{peerAt(upper, otherIP)}, ip, lower, false},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sw := switchWithID(types.GenerateNodeKey().ID())

			for _, p := range testCase.peers {
				require.NoError(t, sw.peers.Add(p))
			}

			assert.Equal(t, testCase.want, sw.hasPeerFromIP(testCase.ip, testCase.except))
		})
	}
}

func TestMultiplexSwitch_AcceptLoop_SimultaneousOpen(t *testing.T) {
	t.Parallel()

	// Each row has us at the higher ID, so the peer's inbound connection wins
	// against our outbound one to the same peer, when that one is registered
	guardTable := []struct {
		name string
		// opts returns the switch options, given the peer's ID
		opts func(peer types.ID) []SwitchOption
		// unregistered leaves our outbound connection out of the peer set
		unregistered bool
		// sharedIP puts both connections to the peer on one IP
		sharedIP bool
		// otherOnIP registers a second peer on that IP
		otherOnIP    bool
		wantReplaced bool
	}{
		{
			// No inbound slot left, the duplicate-IP guard on (the default),
			// and the peer is persistent
			name: "a persistent peer's winning inbound connection replaces ours at the inbound limit",
			opts: func(peer types.ID) []SwitchOption {
				return []SwitchOption{
					WithMaxInboundPeers(0),
					WithPersistentPeers([]*types.NetAddress{{ID: peer}}),
				}
			},
			sharedIP:     true,
			wantReplaced: true,
		},
		{
			// No inbound slot left, and the peer is persistent, but it has no
			// registered connection to replace
			name: "a persistent peer's inbound connection is refused at the inbound limit when nothing is registered",
			opts: func(peer types.ID) []SwitchOption {
				return []SwitchOption{
					WithMaxInboundPeers(0),
					WithPersistentPeers([]*types.NetAddress{{ID: peer}}),
				}
			},
			unregistered: true,
		},
		{
			// The peer is not persistent, so its replacement needs an inbound slot
			name: "a winning inbound connection is refused at the inbound limit",
			opts: func(types.ID) []SwitchOption {
				return []SwitchOption{WithMaxInboundPeers(0)}
			},
		},
		{
			// The duplicate-IP guard is on, and the only peer from that IP is
			// the connection being replaced
			name:         "a winning inbound connection from the replaced connection's IP replaces it",
			sharedIP:     true,
			wantReplaced: true,
		},
		{
			name:      "a winning inbound connection is refused while another peer holds its IP",
			sharedIP:  true,
			otherOnIP: true,
		},
	}

	for _, testCase := range guardTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			lower, upper := orderedIDs(t)

			var (
				ip = net.ParseIP("127.0.0.1")

				// We dialed the peer while it dialed us
				ours   = peerWithID(t, lower, true)
				theirs = peerWithID(t, lower, false)
				closed = closedSignal(theirs)
			)

			if testCase.sharedIP {
				ours.RemoteIPFn = func() net.IP { return ip }
				theirs.RemoteIPFn = func() net.IP { return ip }
			}

			var opts []SwitchOption
			if testCase.opts != nil {
				opts = testCase.opts(lower)
			}

			sw := acceptSwitch(upper, theirs, opts...)

			if !testCase.unregistered {
				require.NoError(t, sw.peers.Add(ours))
			}

			if testCase.otherOnIP {
				other := mock.GeneratePeers(t, 1)[0]
				other.RemoteIPFn = func() net.IP { return ip }

				require.NoError(t, sw.peers.Add(other))
			}

			go sw.runAcceptLoop(t.Context())

			if testCase.wantReplaced {
				require.Eventually(t, func() bool {
					return sw.peers.Get(lower) == PeerConn(theirs)
				}, 5*time.Second, 10*time.Millisecond)

				return
			}

			awaitClosed(t, closed, "the connection was not refused")

			if testCase.unregistered {
				assert.False(t, sw.peers.Has(lower))

				return
			}

			assert.Same(t, ours, sw.peers.Get(lower))
		})
	}

	t.Run("a losing inbound connection is rejected and closed", func(t *testing.T) {
		t.Parallel()

		lower, upper := orderedIDs(t)

		var (
			// Our ID is the lower one, so our outbound connection is kept
			ours   = peerWithID(t, upper, true)
			theirs = peerWithID(t, upper, false)
			closed = closedSignal(theirs)
		)

		sw := acceptSwitch(lower, theirs)

		logs := captureLogs(sw)

		require.NoError(t, sw.peers.Add(ours))

		go sw.runAcceptLoop(t.Context())

		awaitClosed(t, closed, "the losing connection was not closed")

		assert.Contains(t, logs.String(), "Ignoring inbound connection: already connected")
		assert.Contains(t, logs.String(), "kept=outbound")

		assert.Same(t, ours, sw.peers.Get(upper))
	})

	t.Run("a same-direction inbound connection is rejected", func(t *testing.T) {
		t.Parallel()

		lower, upper := orderedIDs(t)

		var (
			// The peer dials us again while its first connection is live
			first  = peerWithID(t, lower, false)
			second = peerWithID(t, lower, false)
			closed = closedSignal(second)
		)

		sw := acceptSwitch(upper, second)

		require.NoError(t, sw.peers.Add(first))

		go sw.runAcceptLoop(t.Context())

		awaitClosed(t, closed, "the second connection was not closed")

		assert.Same(t, first, sw.peers.Get(lower))
	})
}

func TestMultiplexSwitch_AcceptLoopStopsRefusedPeerBeforeClosingIt(t *testing.T) {
	t.Parallel()

	lower, upper := orderedIDs(t)

	var (
		removed  atomic.Int64
		reported bool

		// The connection loses registration after it was started
		mockSet = &mockSet{
			addFn: func(PeerConn) error { return errDuplicatePeer },
		}

		reactor = &mockReactor{
			removePeerFn: func(PeerConn, any) { removed.Add(1) },
		}

		p         = peerWithID(t, upper, false)
		closed    = make(chan struct{})
		closeOnce sync.Once
	)

	withRealStop(p)

	sw := acceptSwitch(lower, p, WithReactor("mock", reactor))
	sw.peers = mockSet

	logs := captureLogs(sw)

	// A recv routine still running reports the closed socket, once, as
	// MConnection.stopForError does
	p.CloseConnFn = func() error {
		defer closeOnce.Do(func() { close(closed) })

		if p.IsRunning() && !reported {
			reported = true
			sw.StopPeerForError(p, errors.New("EOF"))
		}

		return nil
	}

	go sw.runAcceptLoop(t.Context())

	awaitClosed(t, closed, "the refused connection was not closed")

	require.Eventually(t, func() bool {
		return strings.Contains(logs.String(), "error while adding peer")
	}, 5*time.Second, time.Millisecond)

	// addPeer's unwind is the only teardown of the refused connection
	assert.EqualValues(t, 1, removed.Load())
}

func TestMultiplexSwitch_DialPeerTearsDownRefusedPeerOnce(t *testing.T) {
	t.Parallel()

	// The recv routine of a connection refused at registration reports the
	// remote closing it, once. Each row has the report land at a different
	// point of addPeer's teardown of that connection
	testTable := []struct {
		name string
		// atStop has the report win addPeer's stop of the connection, instead
		// of landing while addPeer gives back its reactor state
		atStop bool
	}{
		{"reported while its reactor state is given back", false},
		{"reported as it is stopped", true},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				removed  int
				reported bool

				sw *MultiplexSwitch
				p  = mock.GeneratePeers(t, 1)[0]
			)

			report := func() {
				if !reported {
					reported = true
					sw.StopPeerForError(p, errors.New("EOF"))
				}
			}

			p.StopFn = func() error {
				if testCase.atStop {
					report()
				}

				return p.BaseService.Stop()
			}

			reactor := &mockReactor{
				removePeerFn: func(PeerConn, any) {
					removed++

					if !testCase.atStop {
						report()
					}
				},
			}

			sw = NewMultiplexSwitch(nil, WithReactor("mock", reactor))

			// The connection loses registration after it was started
			sw.peers = &mockSet{
				addFn: func(PeerConn) error { return errDuplicatePeer },
			}

			logs := captureLogs(sw)

			dialRejected(t, sw, p)

			// Whichever stops the refused connection gives back its reactor
			// state, and the other tears nothing down
			assert.True(t, reported)
			assert.Equal(t, 1, removed)
			assert.False(t, p.IsRunning())
			assert.NotContains(t, logs.String(), "unable to gracefully stop peer")
		})
	}
}

// captureLogs makes the switch log into the returned buffer
func captureLogs(sw *MultiplexSwitch) *lockedBuffer {
	logs := &lockedBuffer{}
	sw.SetLogger(slog.New(slog.NewTextHandler(logs, nil)))

	return logs
}

// lockedBuffer is a bytes.Buffer a logger can write to from another goroutine
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
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
