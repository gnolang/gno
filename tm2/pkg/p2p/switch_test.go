package p2p

import (
	"context"
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

func TestMultiplexSwitch_Broadcast(t *testing.T) {
	t.Parallel()

	var (
		wg sync.WaitGroup

		expectedChID = byte(10)
		expectedData = []byte("broadcast data")

		peers = mock.GeneratePeers(t, 10)
		sw    = NewSwitch(nil)
	)

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
		sw    = NewSwitch(nil)
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
				removeFn: func(removedPeer Peer) {
					assert.Equal(t, p.ID(), removedPeer.ID())
				},
			}

			sw = NewSwitch(mockTransport)
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
				removeFn: func(removedPeer Peer) {
					assert.Equal(t, p.ID(), removedPeer.ID())
				},
				netAddressFn: func() types.NetAddress {
					return types.NetAddress{}
				},
			}

			sw = NewSwitch(mockTransport)
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
		sw.dialQueue.Has(p.NodeInfo().NetAddress)
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
				) (Peer, error) {
					peerDialed = true

					return nil, nil
				},
			}

			sw = NewSwitch(mockTransport)
		)

		sw.peers = mockSet

		// Prepare the dial queue
		sw.dialQueue.Push(dial.Item{
			Time:    dialTime,
			Address: p.NodeInfo().NetAddress,
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
				) (Peer, error) {
					peerDialed = true

					cancelFn()

					ch <- struct{}{}

					return nil, errors.New("invalid dial")
				},
			}

			sw = NewSwitch(mockTransport)
		)

		sw.peers = mockSet

		// Prepare the dial queue
		sw.dialQueue.Push(dial.Item{
			Time:    dialTime,
			Address: p.NodeInfo().NetAddress,
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
				) (Peer, error) {
					peerDialed = true

					cancelFn()

					ch <- struct{}{}

					return p, nil
				},
			}

			sw = NewSwitch(mockTransport)
		)

		// Prepare the dial queue
		sw.dialQueue.Push(dial.Item{
			Time:    dialTime,
			Address: p.NodeInfo().NetAddress,
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

	// TODO implement
}
