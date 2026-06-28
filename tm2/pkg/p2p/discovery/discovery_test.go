package discovery

import (
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/p2p/mock"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestStore creates an empty peer store backed by a temp file.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	s, err := NewStore(filepath.Join(t.TempDir(), peerStoreFile), types.NetAddress{})
	require.NoError(t, err)

	return s
}

func TestReactor_DiscoveryRequest(t *testing.T) {
	t.Parallel()

	var (
		notifCh = make(chan struct{}, 1)

		capturedSend []byte

		mockPeer = &mock.Peer{
			SendFn: func(chID byte, data []byte) bool {
				require.Equal(t, Channel, chID)

				capturedSend = data

				notifCh <- struct{}{}

				return true
			},
		}

		ps = &mockPeerSet{
			listFn: func() []p2p.PeerConn {
				return []p2p.PeerConn{mockPeer}
			},
		}

		mockSwitch = &mockSwitch{
			peersFn: func() p2p.PeerSet {
				return ps
			},
		}
	)

	r := NewReactor(
		newTestStore(t),
		WithDiscoveryInterval(10*time.Millisecond),
	)

	// Set the mock switch
	r.SetSwitch(mockSwitch)

	// Start the discovery service
	require.NoError(t, r.Start())
	t.Cleanup(func() {
		require.NoError(t, r.Stop())
	})

	select {
	case <-notifCh:
	case <-time.After(5 * time.Second):
	}

	// Make sure the adequate message was captured
	require.NotNil(t, capturedSend)

	// Parse the message
	var msg Message

	require.NoError(t, amino.Unmarshal(capturedSend, &msg))

	// Make sure the base message is valid
	require.NoError(t, msg.ValidateBasic())

	_, ok := msg.(*Request)

	require.True(t, ok)
}

func TestReactor_DiscoveryResponse(t *testing.T) {
	t.Parallel()

	t.Run("discovery request received", func(t *testing.T) {
		t.Parallel()

		var (
			peers   = mock.GeneratePeers(t, 50)
			notifCh = make(chan struct{}, 1)

			capturedSend []byte

			mockPeer = &mock.Peer{
				SendFn: func(chID byte, data []byte) bool {
					require.Equal(t, Channel, chID)

					capturedSend = data

					notifCh <- struct{}{}

					return true
				},
			}

			ps = &mockPeerSet{
				listFn: func() []p2p.PeerConn {
					listed := make([]p2p.PeerConn, 0, len(peers))

					for _, peer := range peers {
						listed = append(listed, peer)
					}

					return listed
				},
				numInboundFn: func() uint64 {
					return uint64(len(peers))
				},
			}

			mockSwitch = &mockSwitch{
				peersFn: func() p2p.PeerSet {
					return ps
				},
			}
		)

		r := NewReactor(
			newTestStore(t),
			WithDiscoveryInterval(10*time.Millisecond),
		)

		// Set the mock switch
		r.SetSwitch(mockSwitch)

		// Prepare the message
		req := &Request{}

		preparedReq, err := amino.MarshalAny(req)
		require.NoError(t, err)

		// Receive the message
		r.Receive(Channel, mockPeer, preparedReq)

		select {
		case <-notifCh:
		case <-time.After(5 * time.Second):
		}

		// Make sure the adequate message was captured
		require.NotNil(t, capturedSend)

		// Parse the message
		var msg Message

		require.NoError(t, amino.Unmarshal(capturedSend, &msg))

		// Make sure the base message is valid
		require.NoError(t, msg.ValidateBasic())

		resp, ok := msg.(*Response)
		require.True(t, ok)

		// Make sure the peers are valid
		require.Len(t, resp.Peers, maxPeersShared)

		require.True(t, slices.ContainsFunc(resp.Peers, func(addr *types.NetAddress) bool {
			for _, localP := range peers {
				if localP.NodeInfo().DialAddress().Equals(*addr) {
					return true
				}
			}

			return false
		}))
	})

	t.Run("empty peers on discover", func(t *testing.T) {
		t.Parallel()

		var (
			capturedSend []byte

			mockPeer = &mock.Peer{
				SendFn: func(chID byte, data []byte) bool {
					require.Equal(t, Channel, chID)

					capturedSend = data

					return true
				},
			}

			ps = &mockPeerSet{
				listFn: func() []p2p.PeerConn {
					return make([]p2p.PeerConn, 0)
				},
			}

			mockSwitch = &mockSwitch{
				peersFn: func() p2p.PeerSet {
					return ps
				},
			}
		)

		r := NewReactor(
			newTestStore(t),
			WithDiscoveryInterval(10*time.Millisecond),
		)

		// Set the mock switch
		r.SetSwitch(mockSwitch)

		// Prepare the message
		req := &Request{}

		preparedReq, err := amino.MarshalAny(req)
		require.NoError(t, err)

		// Receive the message
		r.Receive(Channel, mockPeer, preparedReq)

		// Make sure no message was captured
		assert.Nil(t, capturedSend)
	})

	t.Run("private peers not shared", func(t *testing.T) {
		t.Parallel()

		var (
			publicPeers  = 1
			privatePeers = 50

			peers   = mock.GeneratePeers(t, publicPeers+privatePeers)
			notifCh = make(chan struct{}, 1)

			capturedSend []byte

			mockPeer = &mock.Peer{
				SendFn: func(chID byte, data []byte) bool {
					require.Equal(t, Channel, chID)

					capturedSend = data

					notifCh <- struct{}{}

					return true
				},
			}

			ps = &mockPeerSet{
				listFn: func() []p2p.PeerConn {
					listed := make([]p2p.PeerConn, 0, len(peers))

					for _, peer := range peers {
						listed = append(listed, peer)
					}

					return listed
				},
				numInboundFn: func() uint64 {
					return uint64(len(peers))
				},
			}

			mockSwitch = &mockSwitch{
				peersFn: func() p2p.PeerSet {
					return ps
				},
			}
		)

		// Mark all except the last X peers as private
		for _, peer := range peers[:privatePeers] {
			peer.IsPrivateFn = func() bool {
				return true
			}
		}

		r := NewReactor(
			newTestStore(t),
			WithDiscoveryInterval(10*time.Millisecond),
		)

		// Set the mock switch
		r.SetSwitch(mockSwitch)

		// Prepare the message
		req := &Request{}

		preparedReq, err := amino.MarshalAny(req)
		require.NoError(t, err)

		// Receive the message
		r.Receive(Channel, mockPeer, preparedReq)

		select {
		case <-notifCh:
		case <-time.After(5 * time.Second):
		}

		// Make sure the adequate message was captured
		require.NotNil(t, capturedSend)

		// Parse the message
		var msg Message

		require.NoError(t, amino.Unmarshal(capturedSend, &msg))

		// Make sure the base message is valid
		require.NoError(t, msg.ValidateBasic())

		resp, ok := msg.(*Response)
		require.True(t, ok)

		// Make sure the peers are valid
		require.Len(t, resp.Peers, publicPeers)

		require.True(t, slices.ContainsFunc(resp.Peers, func(addr *types.NetAddress) bool {
			for _, localP := range peers {
				if localP.NodeInfo().DialAddress().Equals(*addr) {
					return true
				}
			}

			return false
		}))
	})

	t.Run("peer response received", func(t *testing.T) {
		t.Parallel()

		var (
			peers   = mock.GeneratePeers(t, 50)
			notifCh = make(chan struct{}, 1)

			capturedDials []*types.NetAddress

			ps = &mockPeerSet{
				listFn: func() []p2p.PeerConn {
					listed := make([]p2p.PeerConn, 0, len(peers))

					for _, peer := range peers {
						listed = append(listed, peer)
					}

					return listed
				},
				numInboundFn: func() uint64 {
					return uint64(len(peers))
				},
			}

			mockSwitch = &mockSwitch{
				peersFn: func() p2p.PeerSet {
					return ps
				},
				dialPeersFn: func(addresses ...*types.NetAddress) {
					capturedDials = append(capturedDials, addresses...)

					notifCh <- struct{}{}
				},
			}
		)

		r := NewReactor(
			newTestStore(t),
			WithDiscoveryInterval(10*time.Millisecond),
		)

		// Set the mock switch
		r.SetSwitch(mockSwitch)

		// Prepare the addresses
		peerAddrs := make([]*types.NetAddress, 0, len(peers))

		for _, p := range peers {
			peerAddrs = append(peerAddrs, p.NodeInfo().DialAddress())
		}

		// Prepare the message
		req := &Response{
			Peers: peerAddrs,
		}

		preparedReq, err := amino.MarshalAny(req)
		require.NoError(t, err)

		// Receive the message
		r.Receive(Channel, &mock.Peer{}, preparedReq)

		select {
		case <-notifCh:
		case <-time.After(5 * time.Second):
		}

		// Make sure the correct peers were dialed
		assert.Equal(t, capturedDials, peerAddrs)
	})

	t.Run("invalid peer response received", func(t *testing.T) {
		t.Parallel()

		var (
			peers = mock.GeneratePeers(t, 50)

			capturedDials []*types.NetAddress

			ps = &mockPeerSet{
				listFn: func() []p2p.PeerConn {
					listed := make([]p2p.PeerConn, 0, len(peers))

					for _, peer := range peers {
						listed = append(listed, peer)
					}

					return listed
				},
				numInboundFn: func() uint64 {
					return uint64(len(peers))
				},
			}

			mockSwitch = &mockSwitch{
				peersFn: func() p2p.PeerSet {
					return ps
				},
				dialPeersFn: func(addresses ...*types.NetAddress) {
					capturedDials = append(capturedDials, addresses...)
				},
			}
		)

		r := NewReactor(
			newTestStore(t),
			WithDiscoveryInterval(10*time.Millisecond),
		)

		// Set the mock switch
		r.SetSwitch(mockSwitch)

		// Prepare the message
		req := &Response{
			Peers: make([]*types.NetAddress, 0), // empty
		}

		preparedReq, err := amino.MarshalAny(req)
		require.NoError(t, err)

		// Receive the message
		r.Receive(Channel, &mock.Peer{}, preparedReq)

		// Make sure no peers were dialed
		assert.Empty(t, capturedDials)
	})
}

func TestReactor_Store_PersistsDiscoveredPeers(t *testing.T) {
	t.Parallel()

	var (
		peers = mock.GeneratePeers(t, 5)

		storePath = filepath.Join(t.TempDir(), peerStoreFile)

		ps = &mockPeerSet{
			listFn: func() []p2p.PeerConn {
				return make([]p2p.PeerConn, 0)
			},
		}

		mockSwitch = &mockSwitch{
			peersFn: func() p2p.PeerSet {
				return ps
			},
		}
	)

	store, err := NewStore(storePath, types.NetAddress{})
	require.NoError(t, err)

	r := NewReactor(
		store,
		WithDiscoveryInterval(time.Hour), // avoid background discovery
	)

	r.SetSwitch(mockSwitch)

	// Prepare the discovered peer addresses
	peerAddrs := make([]*types.NetAddress, 0, len(peers))
	for _, p := range peers {
		peerAddrs = append(peerAddrs, p.NodeInfo().DialAddress())
	}

	// Prepare and send a discovery Response
	resp := &Response{Peers: peerAddrs}

	preparedResp, err := amino.MarshalAny(resp)
	require.NoError(t, err)

	r.Receive(Channel, &mock.Peer{}, preparedResp)

	// The store should now contain all discovered peers
	assert.Equal(t, len(peers), store.Size())

	// Stopping the reactor should flush to disk
	r.OnStop()

	// Reload the store from disk and verify persistence
	reloaded, err := NewStore(storePath, types.NetAddress{})
	require.NoError(t, err)

	assert.Equal(t, len(peers), reloaded.Size())
}

func TestReactor_Store_DialsPersistedPeersOnStart(t *testing.T) {
	t.Parallel()

	var (
		peers = mock.GeneratePeers(t, 5)

		storePath = filepath.Join(t.TempDir(), peerStoreFile)

		notifCh = make(chan struct{}, 1)

		ps = &mockPeerSet{
			listFn: func() []p2p.PeerConn {
				return make([]p2p.PeerConn, 0)
			},
		}

		capturedDials []*types.NetAddress

		mockSwitch = &mockSwitch{
			peersFn: func() p2p.PeerSet {
				return ps
			},
			dialPeersFn: func(addresses ...*types.NetAddress) {
				capturedDials = append(capturedDials, addresses...)

				notifCh <- struct{}{}
			},
		}
	)

	// Pre-populate the store with discovered peers
	store, err := NewStore(storePath, types.NetAddress{})
	require.NoError(t, err)

	peerAddrs := make([]*types.NetAddress, 0, len(peers))
	for _, p := range peers {
		peerAddrs = append(peerAddrs, p.NodeInfo().DialAddress())
	}

	store.AddPeers(peerAddrs...)
	require.NoError(t, store.Save())

	// Create a fresh store that loads from disk
	loadedStore, err := NewStore(storePath, types.NetAddress{})
	require.NoError(t, err)
	require.Equal(t, len(peers), loadedStore.Size())

	r := NewReactor(
		loadedStore,
		WithDiscoveryInterval(time.Hour), // avoid background discovery
	)

	r.SetSwitch(mockSwitch)

	// OnStart should dial the persisted peers
	require.NoError(t, r.OnStart())
	t.Cleanup(func() {
		r.OnStop()
	})

	select {
	case <-notifCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for persisted peers to be dialed")
	}

	// All persisted peers should have been dialed
	assert.Len(t, capturedDials, len(peers))
}
