package p2p

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/versionset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateNetAddr generates dummy net addresses
func generateNetAddr(t *testing.T, count int) []*types.NetAddress {
	t.Helper()

	addrs := make([]*types.NetAddress, 0, count)

	for range count {
		key := types.GenerateNodeKey()

		// Grab a random port
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		addr, err := types.NewNetAddress(key.ID(), ln.Addr())
		require.NoError(t, err)

		addrs = append(addrs, addr)
	}

	return addrs
}

func TestMultiplexTransport_NetAddress(t *testing.T) {
	t.Parallel()

	t.Run("transport not active", func(t *testing.T) {
		t.Parallel()

		var (
			ni     = types.NodeInfo{}
			nk     = types.NodeKey{}
			mCfg   = conn.DefaultMConnConfig()
			logger = log.NewNoopLogger()
		)

		transport := NewMultiplexTransport(ni, nk, mCfg, logger)
		addr := transport.NetAddress()

		assert.Error(t, addr.Validate())
	})

	t.Run("active transport on random port", func(t *testing.T) {
		t.Parallel()

		var (
			ni     = types.NodeInfo{}
			nk     = types.NodeKey{}
			mCfg   = conn.DefaultMConnConfig()
			logger = log.NewNoopLogger()
			addr   = generateNetAddr(t, 1)[0]
		)

		addr.Port = 0 // random port

		transport := NewMultiplexTransport(ni, nk, mCfg, logger)

		require.NoError(t, transport.Listen(*addr))
		defer func() {
			require.NoError(t, transport.Close())
		}()

		netAddr := transport.NetAddress()
		assert.False(t, netAddr.Equals(*addr))
		assert.NoError(t, netAddr.Validate())
	})

	t.Run("active transport on specific port", func(t *testing.T) {
		t.Parallel()

		var (
			ni     = types.NodeInfo{}
			nk     = types.NodeKey{}
			mCfg   = conn.DefaultMConnConfig()
			logger = log.NewNoopLogger()
			addr   = generateNetAddr(t, 1)[0]
		)

		addr.Port = 4123 // specific port

		transport := NewMultiplexTransport(ni, nk, mCfg, logger)

		require.NoError(t, transport.Listen(*addr))
		defer func() {
			require.NoError(t, transport.Close())
		}()

		netAddr := transport.NetAddress()
		assert.True(t, netAddr.Equals(*addr))
		assert.NoError(t, netAddr.Validate())
	})
}

func TestMultiplexTransport_Accept(t *testing.T) {
	t.Parallel()

	t.Run("inactive transport", func(t *testing.T) {
		t.Parallel()

		var (
			ni     = types.NodeInfo{}
			nk     = types.NodeKey{}
			mCfg   = conn.DefaultMConnConfig()
			logger = log.NewNoopLogger()
		)

		transport := NewMultiplexTransport(ni, nk, mCfg, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		p, err := transport.Accept(ctx, nil)
		assert.Nil(t, p)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("transport closed", func(t *testing.T) {
		t.Parallel()

		var (
			ni     = types.NodeInfo{}
			nk     = types.NodeKey{}
			mCfg   = conn.DefaultMConnConfig()
			logger = log.NewNoopLogger()
			addr   = generateNetAddr(t, 1)[0]
		)

		addr.Port = 0

		transport := NewMultiplexTransport(ni, nk, mCfg, logger)

		// Start the transport
		require.NoError(t, transport.Listen(*addr))

		// Stop the transport
		require.NoError(t, transport.Close())

		p, err := transport.Accept(context.Background(), nil)

		assert.Nil(t, p)
		assert.ErrorIs(
			t,
			err,
			errTransportClosed,
		)
	})

	t.Run("context canceled", func(t *testing.T) {
		t.Parallel()

		var (
			ni     = types.NodeInfo{}
			nk     = types.NodeKey{}
			mCfg   = conn.DefaultMConnConfig()
			logger = log.NewNoopLogger()
			addr   = generateNetAddr(t, 1)[0]
		)

		addr.Port = 0

		transport := NewMultiplexTransport(ni, nk, mCfg, logger)

		// Start the transport
		require.NoError(t, transport.Listen(*addr))

		ctx, cancelFn := context.WithCancel(context.Background())
		cancelFn()

		p, err := transport.Accept(ctx, nil)

		assert.Nil(t, p)
		assert.ErrorIs(
			t,
			err,
			context.Canceled,
		)
	})

	t.Run("peer ID mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			network = "dev"
			mCfg    = conn.DefaultMConnConfig()
			logger  = log.NewNoopLogger()
			keys    = []*types.NodeKey{
				types.GenerateNodeKey(),
				types.GenerateNodeKey(),
			}

			peerBehavior = &reactorPeerBehavior{
				chDescs:      make([]*conn.ChannelDescriptor, 0),
				reactorsByCh: make(map[byte]Reactor),
				handlePeerErrFn: func(_ PeerConn, err error) {
					require.NoError(t, err)
				},
				isPersistentPeerFn: func(_ types.ID) bool {
					return false
				},
				isPrivatePeerFn: func(_ types.ID) bool {
					return false
				},
			}
		)

		peers := make([]*MultiplexTransport, 0, len(keys))

		for index, key := range keys {
			addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
			require.NoError(t, err)

			// Hijack the key value
			id := types.GenerateNodeKey().ID()

			na, err := types.NewNetAddress(id, addr)
			require.NoError(t, err)

			ni := types.NodeInfo{
				Network:    network, // common network
				NetAddress: na,
				Version:    "v1.0.0-rc.0",
				Moniker:    fmt.Sprintf("node-%d", index),
				VersionSet: make(versionset.VersionSet, 0), // compatible version set
				Channels:   []byte{42},                     // common channel
			}

			// Create a fresh transport
			tr := NewMultiplexTransport(ni, *key, mCfg, logger)

			// Start the transport
			require.NoError(t, tr.Listen(*na))

			t.Cleanup(func() {
				assert.NoError(t, tr.Close())
			})

			peers = append(
				peers,
				tr,
			)
		}

		// Make peer 1 --dial--> peer 2, and handshake.
		// This "upgrade" should fail because the peer shared a different
		// peer ID than what they actually used for the secret connection
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		p, err := peers[0].Dial(ctx, peers[1].netAddr, peerBehavior)
		assert.ErrorIs(t, err, errPeerIDNodeInfoMismatch)
		require.Nil(t, p)
	})

	t.Run("incompatible peers", func(t *testing.T) {
		t.Parallel()

		var (
			network = "dev"
			mCfg    = conn.DefaultMConnConfig()
			logger  = log.NewNoopLogger()
			keys    = []*types.NodeKey{
				types.GenerateNodeKey(),
				types.GenerateNodeKey(),
			}

			peerBehavior = &reactorPeerBehavior{
				chDescs:      make([]*conn.ChannelDescriptor, 0),
				reactorsByCh: make(map[byte]Reactor),
				handlePeerErrFn: func(_ PeerConn, err error) {
					require.NoError(t, err)
				},
				isPersistentPeerFn: func(_ types.ID) bool {
					return false
				},
				isPrivatePeerFn: func(_ types.ID) bool {
					return false
				},
			}
		)

		peers := make([]*MultiplexTransport, 0, len(keys))

		for index, key := range keys {
			addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
			require.NoError(t, err)

			id := key.ID()

			na, err := types.NewNetAddress(id, addr)
			require.NoError(t, err)

			chainID := network

			if index%2 == 0 {
				chainID = "totally-random-network"
			}

			ni := types.NodeInfo{
				Network:    chainID,
				NetAddress: na,
				Version:    "v1.0.0-rc.0",
				Moniker:    fmt.Sprintf("node-%d", index),
				VersionSet: make(versionset.VersionSet, 0), // compatible version set
				Channels:   []byte{42},                     // common channel
			}

			// Create a fresh transport
			tr := NewMultiplexTransport(ni, *key, mCfg, logger)

			// Start the transport
			require.NoError(t, tr.Listen(*na))

			t.Cleanup(func() {
				assert.NoError(t, tr.Close())
			})

			peers = append(
				peers,
				tr,
			)
		}

		// Make peer 1 --dial--> peer 2, and handshake.
		// This "upgrade" should fail because the peer shared a different
		// peer ID than what they actually used for the secret connection
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		p, err := peers[0].Dial(ctx, peers[1].netAddr, peerBehavior)
		assert.ErrorIs(t, err, errIncompatibleNodeInfo)
		require.Nil(t, p)
	})

	t.Run("dialed peer ID mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			network = "dev"
			mCfg    = conn.DefaultMConnConfig()
			logger  = log.NewNoopLogger()
			keys    = []*types.NodeKey{
				types.GenerateNodeKey(),
				types.GenerateNodeKey(),
			}

			peerBehavior = &reactorPeerBehavior{
				chDescs:      make([]*conn.ChannelDescriptor, 0),
				reactorsByCh: make(map[byte]Reactor),
				handlePeerErrFn: func(_ PeerConn, err error) {
					require.NoError(t, err)
				},
				isPersistentPeerFn: func(_ types.ID) bool {
					return false
				},
				isPrivatePeerFn: func(_ types.ID) bool {
					return false
				},
			}
		)

		peers := make([]*MultiplexTransport, 0, len(keys))

		for index, key := range keys {
			addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
			require.NoError(t, err)

			na, err := types.NewNetAddress(key.ID(), addr)
			require.NoError(t, err)

			ni := types.NodeInfo{
				Network:    network, // common network
				NetAddress: na,
				Version:    "v1.0.0-rc.0",
				Moniker:    fmt.Sprintf("node-%d", index),
				VersionSet: make(versionset.VersionSet, 0), // compatible version set
				Channels:   []byte{42},                     // common channel
			}

			// Create a fresh transport
			tr := NewMultiplexTransport(ni, *key, mCfg, logger)

			// Start the transport
			require.NoError(t, tr.Listen(*na))

			t.Cleanup(func() {
				assert.NoError(t, tr.Close())
			})

			peers = append(
				peers,
				tr,
			)
		}

		// Make peer 1 --dial--> peer 2, and handshake
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		p, err := peers[0].Dial(
			ctx,
			types.NetAddress{
				ID:   types.GenerateNodeKey().ID(), // mismatched ID
				IP:   peers[1].netAddr.IP,
				Port: peers[1].netAddr.Port,
			},
			peerBehavior,
		)
		assert.ErrorIs(t, err, errPeerIDDialMismatch)
		assert.Nil(t, p)
	})

	t.Run("valid peer accepted", func(t *testing.T) {
		t.Parallel()

		var (
			network = "dev"
			mCfg    = conn.DefaultMConnConfig()
			logger  = log.NewNoopLogger()
			keys    = []*types.NodeKey{
				types.GenerateNodeKey(),
				types.GenerateNodeKey(),
			}

			peerBehavior = &reactorPeerBehavior{
				chDescs:      make([]*conn.ChannelDescriptor, 0),
				reactorsByCh: make(map[byte]Reactor),
				handlePeerErrFn: func(_ PeerConn, err error) {
					require.NoError(t, err)
				},
				isPersistentPeerFn: func(_ types.ID) bool {
					return false
				},
				isPrivatePeerFn: func(_ types.ID) bool {
					return false
				},
			}
		)

		peers := make([]*MultiplexTransport, 0, len(keys))

		for index, key := range keys {
			addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
			require.NoError(t, err)

			na, err := types.NewNetAddress(key.ID(), addr)
			require.NoError(t, err)

			ni := types.NodeInfo{
				Network:    network, // common network
				NetAddress: na,
				Version:    "v1.0.0-rc.0",
				Moniker:    fmt.Sprintf("node-%d", index),
				VersionSet: make(versionset.VersionSet, 0), // compatible version set
				Channels:   []byte{42},                     // common channel
			}

			// Create a fresh transport
			tr := NewMultiplexTransport(ni, *key, mCfg, logger)

			// Start the transport
			require.NoError(t, tr.Listen(*na))

			t.Cleanup(func() {
				assert.NoError(t, tr.Close())
			})

			peers = append(
				peers,
				tr,
			)
		}

		// Make peer 1 --dial--> peer 2, and handshake
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		p, err := peers[0].Dial(ctx, peers[1].netAddr, peerBehavior)
		require.NoError(t, err)
		require.NotNil(t, p)

		// Make sure the new peer info is valid
		assert.Equal(t, peers[1].netAddr.ID, p.ID())

		assert.Equal(t, peers[1].nodeInfo.Channels, p.NodeInfo().Channels)
		assert.Equal(t, peers[1].nodeInfo.Moniker, p.NodeInfo().Moniker)
		assert.Equal(t, peers[1].nodeInfo.Network, p.NodeInfo().Network)

		// Attempt to dial again, expect the dial to fail
		// because the connection is already active
		dialedPeer, err := peers[0].Dial(ctx, peers[1].netAddr, peerBehavior)
		require.ErrorIs(t, err, errDuplicateConnection)
		assert.Nil(t, dialedPeer)

		// Remove the peer
		peers[0].Remove(p)
	})
}
