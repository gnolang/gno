package conn

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// tcpPair returns a connected pair over loopback TCP. Unlike net.Pipe this has
// kernel send/recv buffers, so a write does not block until the far side reads
// -- which matters whenever a test needs sendRoutine to keep running while
// recvRoutine is stalled.
func tcpPair(t *testing.T) (server, client net.Conn) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	dialed := make(chan net.Conn, 1)
	go func() {
		c, dialErr := net.Dial("tcp", ln.Addr().String())
		if dialErr != nil {
			close(dialed)

			return
		}
		dialed <- c
	}()

	client, err = ln.Accept()
	require.NoError(t, err)

	server = <-dialed
	require.NotNil(t, server)

	t.Cleanup(func() {
		_ = server.Close()
		_ = client.Close()
	})

	return server, client
}

// TestPongTimeoutNotChargedForReactorStall is the recv-assembly deadline's twin
// at the other end of the same read loop.
//
// A pong is only observed by recvRoutine, and recvRoutine calls onReceive
// inline. While a reactor blocks there the peer's pong sits unread in the
// socket, so a pong window shorter than the stall expires and the connection is
// dropped for "pong timeout" -- blaming a peer that answered on time.
func TestPongTimeoutNotChargedForReactorStall(t *testing.T) {
	t.Parallel()

	server, client := tcpPair(t)

	errorsCh := make(chan error, 1)
	onError := func(r error) {
		select {
		case errorsCh <- r:
		default:
		}
	}

	const (
		pingInterval = 200 * time.Millisecond
		pongTimeout  = 150 * time.Millisecond
		reactorStall = 5 * pongTimeout
	)

	// The first message parks recvRoutine well past the pong window.
	var once sync.Once
	stalling := make(chan struct{})
	onReceive := func(byte, []byte) {
		once.Do(func() {
			close(stalling)
			time.Sleep(reactorStall)
		})
	}

	cfg := DefaultMConnConfig()
	cfg.PingInterval = pingInterval
	cfg.PongTimeout = pongTimeout

	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1, RecvMessageCapacity: 4096},
	}, onReceive, onError, cfg)
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	// A responsive peer: answers every ping with a pong, promptly.
	pongs := make(chan struct{}, 64)
	go func() {
		if err := writeRawPacketMsg(server, 0x01, 0x01, []byte("go")); err != nil {
			return
		}

		for {
			var packet Packet
			if _, err := amino.UnmarshalSizedReader(server, &packet, int64(mconn._maxPacketMsgSize)); err != nil {
				return
			}

			if _, ok := packet.(PacketPing); !ok {
				continue
			}

			if _, err := amino.MarshalAnySizedWriter(server, PacketPong{}); err != nil {
				return
			}

			select {
			case pongs <- struct{}{}:
			default:
			}
		}
	}()

	<-stalling

	// Across the whole stall plus a couple of ping intervals, the connection
	// must survive: the peer is answering, we are the ones not reading.
	select {
	case err := <-errorsCh:
		t.Fatalf("connection torn down by our own reactor stall: %v", err)
	case <-time.After(reactorStall + 3*pingInterval):
	}

	assert.True(t, mconn.IsRunning(), "connection should still be up")
	assert.NotEmpty(t, pongs, "test setup: the peer should have answered pings")
}
