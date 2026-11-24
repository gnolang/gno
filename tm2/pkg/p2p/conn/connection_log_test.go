package conn

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// maxBadPacketLog bounds what one message may put in the log, and what an
// undecodable packet puts in the error the switch logs again when it stops the
// peer.
const maxBadPacketLog = 1024

// lockedBuffer is a log sink a test reads while the connection's routines write.
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

// startLoggedConnection starts a connection over one end of a pipe that logs to
// the returned buffer and sends the error it stops with on the returned channel.
func startLoggedConnection(t *testing.T, server net.Conn) (*lockedBuffer, <-chan error) {
	t.Helper()

	stopErr := make(chan error, 1)
	c := createMConnectionWithCallbacks(t, server, func(byte, []byte) {}, func(err error) { stopErr <- err })

	logged := &lockedBuffer{}

	c.SetLogger(slog.New(slog.NewTextHandler(logged, &slog.HandlerOptions{Level: slog.LevelError})))
	require.NoError(t, c.Start())
	t.Cleanup(func() { _ = c.Stop() })

	return logged, stopErr
}

// TestRecvRoutineBoundsLogOfBadPacket sends a packet whose 1024-byte type URL
// amino quotes in hex, and bounds the log and the stop error.
func TestRecvRoutineBoundsLogOfBadPacket(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer client.Close()

	logged, stopErr := startLoggedConnection(t, server)

	var typeURL bytes.Buffer
	require.NoError(t, amino.EncodeByteSlice(&typeURL, bytes.Repeat([]byte{0xff}, defaultMaxPacketMsgPayloadSize)))

	body := append([]byte{0x0a}, typeURL.Bytes()...)

	go func() { _ = amino.EncodeByteSlice(client, body) }()

	select {
	case err := <-stopErr:
		require.Contains(t, err.Error(), "invalid type_url")
		require.LessOrEqual(t, len(err.Error()), maxBadPacketLog)
		require.Contains(t, logged.String(), "recvRoutine")
		require.LessOrEqual(t, len(logged.String()), maxBadPacketLog)
	case <-time.After(5 * time.Second):
		t.Fatal("the connection did not stop")
	}
}

// TestRecvRoutineLogsNoFailureOnClose closes the peer's end, which stops the
// connection with EOF and logs no receive failure at Error level.
func TestRecvRoutineLogsNoFailureOnClose(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()

	logged, stopErr := startLoggedConnection(t, server)

	require.NoError(t, client.Close())

	select {
	case err := <-stopErr:
		require.ErrorIs(t, err, io.EOF)
		require.NotContains(t, logged.String(), "recvRoutine")
	case <-time.After(5 * time.Second):
		t.Fatal("the connection did not stop")
	}
}

// TestSendBoundsLogOfFailedSend sends to a peer that never reads until the
// channel's queue is full, so the next send times out, and bounds the line it
// logs. It runs alone: its send routine blocks on the pipe for
// defaultSendTimeout, which TestMConnectionPingPongs's leak check would count.
func TestSendBoundsLogOfFailedSend(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	c := NewMConnectionWithConfig(server, []*ChannelDescriptor{{ID: 0x01, Priority: 1, SendQueueCapacity: 1}},
		func(byte, []byte) {}, func(error) {}, DefaultMConnConfig())

	logged := &lockedBuffer{}

	c.SetLogger(slog.New(slog.NewTextHandler(logged, &slog.HandlerOptions{Level: slog.LevelDebug})))
	require.NoError(t, c.Start())
	defer c.Stop()

	// The first message blocks the send routine on the pipe, the second fills
	// the queue, and the third waits out defaultSendTimeout.
	msg := make([]byte, 1<<20)
	require.True(t, c.Send(0x01, msg))
	require.True(t, c.Send(0x01, msg))
	require.False(t, c.Send(0x01, msg))

	var line string
	for l := range strings.SplitSeq(logged.String(), "\n") {
		if strings.Contains(l, "Send failed") {
			line = l
		}
	}
	require.NotEmpty(t, line, "no send failure logged")
	require.LessOrEqual(t, len(line), maxBadPacketLog)
}
