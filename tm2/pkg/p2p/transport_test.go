package p2p

import (
	"context"
	"net"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateNetAddr generates dummy net addresses
func generateNetAddr(t *testing.T, count int) []types.NetAddress {
	addrs := make([]types.NetAddress, 0, count)

	for i := 0; i < count; i++ {
		key := types.GenerateNodeKey()

		// Grab a random port
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		addr, err := types.NewNetAddress(key.ID(), ln.Addr())
		require.NoError(t, err)

		addrs = append(addrs, *addr)
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

		require.NoError(t, transport.Listen(addr))
		defer func() {
			require.NoError(t, transport.Close())
		}()

		netAddr := transport.NetAddress()
		assert.False(t, netAddr.Equals(addr))
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

		require.NoError(t, transport.Listen(addr))
		defer func() {
			require.NoError(t, transport.Close())
		}()

		netAddr := transport.NetAddress()
		assert.True(t, netAddr.Equals(addr))
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

		p, err := transport.Accept(context.Background(), nil)

		assert.Nil(t, p)
		assert.ErrorIs(
			t,
			err,
			errTransportInactive,
		)
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
		require.NoError(t, transport.Listen(addr))

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
		require.NoError(t, transport.Listen(addr))

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

	t.Run("peer accepted", func(t *testing.T) {
		t.Parallel()

		var (
			mCfg   = conn.DefaultMConnConfig()
			logger = log.NewNoopLogger()
			addr   = generateNetAddr(t, 1)[0]

			mockConn = &mockConn{
				remoteAddrFn: func() net.Addr {
					return &net.TCPAddr{
						IP:   addr.IP,
						Port: int(addr.Port),
					}
				},
			}
			mockListener = &mockListener{
				acceptFn: func() (net.Conn, error) {
					return mockConn, nil
				},
			}

			pi = peerInfo{
				addr: &addr,
				conn: mockConn,
				nodeInfo: types.NodeInfo{
					NetAddress: &addr,
				},
			}

			peerBehavior = &reactorPeerBehavior{
				chDescs:      make([]*conn.ChannelDescriptor, 0),
				reactorsByCh: make(map[byte]Reactor),
				handlePeerErrFn: func(_ Peer, err error) {
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

		transport := NewMultiplexTransport(types.NodeInfo{}, types.NodeKey{}, mCfg, logger)

		// Set the listener
		transport.listener = mockListener

		// Prepare the peer info for accepting
		transport.peerCh <- pi

		// Accept the peer
		p, err := transport.Accept(context.Background(), peerBehavior)
		require.NoError(t, err)

		assert.Equal(t, pi.nodeInfo, p.NodeInfo())
	})
}

func TestMultiplexTransport_Dial(t *testing.T) {
	t.Parallel()

	// TODO implement
}

func TestMultiplexTransport_Listen(t *testing.T) {
	t.Parallel()

	// TODO implement
}
