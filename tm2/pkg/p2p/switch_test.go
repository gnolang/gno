package p2p

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/p2p/dial"
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

	require.NoError(t, sw.OnStart())
	t.Cleanup(sw.OnStop)

	// Create a new peer set
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
		sw.peers.Add(p)
	}

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
		sw.peers.Add(p)
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
		sw.peers.Add(p)

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
				netAddressFn: func() types.NetAddress {
					return types.NetAddress{}
				},
			}

			sw = NewMultiplexSwitch(mockTransport)
		)

		// Make sure the peer is persistent
		p.IsPersistentFn = func() bool {
			return true
		}

		p.IsOutboundFn = func() bool {
			return false
		}

		// Create a new peer set
		sw.peers = newSet()

		// Save the single peer
		sw.peers.Add(p)

		// Stop and remove the peer
		sw.StopPeerForError(p, nil)

		// Make sure the peer is removed
		assert.False(t, sw.peers.Has(p.ID()))

		// Make sure the peer is in the dial queue
		sw.dialQueue.Has(p.SocketAddr())
	})
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

			peerDialed bool

			p        = mock.GeneratePeers(t, 1)[0]
			dialTime = time.Now().Add(-5 * time.Second) // in the past

			mockTransport = &mockTransport{
				dialFn: func(
					_ context.Context,
					_ types.NetAddress,
					_ PeerBehavior,
				) (PeerConn, error) {
					peerDialed = true

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

		require.True(t, sw.Peers().Has(p.ID()))

		assert.True(t, peerDialed)
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

			peerRemoved bool

			p = mock.GeneratePeers(t, 1)[0]

			mockTransport = &mockTransport{
				acceptFn: func(_ context.Context, _ PeerBehavior) (PeerConn, error) {
					return p, nil
				},
				removeFn: func(removedPeer PeerConn) {
					require.Equal(t, p.ID(), removedPeer.ID())

					peerRemoved = true

					ch <- struct{}{}
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
		case <-ch:
		case <-time.After(5 * time.Second):
		}

		assert.True(t, peerRemoved)
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

			peerAdded bool

			p = mock.GeneratePeers(t, 1)[0]

			mockTransport = &mockTransport{
				acceptFn: func(_ context.Context, _ PeerBehavior) (PeerConn, error) {
					return p, nil
				},
			}

			ps = &mockSet{
				numInboundFn: func() uint64 {
					return maxInbound - 1 // available slot
				},
				addFn: func(peer PeerConn) {
					require.Equal(t, p.ID(), peer.ID())

					peerAdded = true

					ch <- struct{}{}
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
		case <-ch:
		case <-time.After(5 * time.Second):
		}

		assert.True(t, peerAdded)
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
					if !sw.dialQueue.Has(missingAddr) {
						continue
					}

					cancelFn()

					return
				}
			}
		}()

		wg.Wait()

		require.True(t, sw.dialQueue.Has(missingAddr))
		assert.Equal(t, missingAddr, sw.dialQueue.Peek().Address)
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
				PeerID: addr.ID,
			}
		}

		sw := NewMultiplexSwitch(mockTransport)

		// Dial the peers
		sw.DialPeers(p.SocketAddr())

		// Make sure the peer wasn't actually dialed
		assert.False(t, sw.dialQueue.Has(p.SocketAddr()))
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
