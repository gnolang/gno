package conn

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// writePacketMsg marshals and writes a single PacketMsg to w.
func writePacketMsg(t *testing.T, w net.Conn, chID, eof byte, data []byte) {
	t.Helper()
	var buf bytes.Buffer
	_, err := amino.MarshalAnySizedWriter(&buf, PacketMsg{ChannelID: chID, EOF: eof, Bytes: data})
	require.NoError(t, err)
	_, err = w.Write(buf.Bytes())
	require.NoError(t, err)
}

// quietPingConfig returns a config whose ping/pong machinery is effectively
// disabled, so the only thing that can tear a connection down is the defense
// under test (not the 45s pong timeout).
func quietPingConfig() MConnConfig {
	cfg := DefaultMConnConfig()
	cfg.PingInterval = 30 * time.Minute
	cfg.PongTimeout = 20 * time.Minute // must stay < PingInterval
	return cfg
}

// TestRecvAssemblyTimeout verifies that a peer which starts a message with a
// partial packet (EOF=0) and never completes it is disconnected once the
// assembly deadline elapses — even while it stays responsive to pings.
func TestRecvAssemblyTimeout(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	errorsCh := make(chan error, 1)
	onError := func(r error) {
		select {
		case errorsCh <- r:
		default:
		}
	}

	cfg := quietPingConfig()
	cfg.RecvAssemblyTimeout = 200 * time.Millisecond
	chDescs := []*ChannelDescriptor{{ID: 0x01, Priority: 1, SendQueueCapacity: 1, RecvMessageCapacity: 1024}}
	mconn := NewMConnectionWithConfig(client, chDescs, func(byte, []byte) {}, onError, cfg)
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	// Start a message but never finish it.
	writePacketMsg(t, server, 0x01, 0x00, []byte("partial"))

	select {
	case err := <-errorsCh:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "assembly timeout")
	case <-time.After(3 * time.Second):
		t.Fatal("expected an assembly timeout error")
	}
}

// TestRecvAssemblyTimeoutNotResetByDribble is the key regression test: an
// caller keeps sending partial packets (here, empty ones — the exact dribble
// the earlier proposed fix was affected by) to try to keep resetting the
// deadline. The deadline is anchored to the first partial packet and must fire
// regardless, tearing the connection down.
func TestRecvAssemblyTimeoutNotResetByDribble(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	errorsCh := make(chan error, 1)
	onError := func(r error) {
		select {
		case errorsCh <- r:
		default:
		}
	}

	cfg := quietPingConfig()
	cfg.RecvAssemblyTimeout = 400 * time.Millisecond
	chDescs := []*ChannelDescriptor{{ID: 0x01, Priority: 1, SendQueueCapacity: 1, RecvMessageCapacity: 1 << 20}}
	mconn := NewMConnectionWithConfig(client, chDescs, func(byte, []byte) {}, onError, cfg)
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	// First partial packet starts the (anchored) deadline.
	writePacketMsg(t, server, 0x01, 0x00, []byte("start"))

	// Keep dribbling empty partial packets every 100ms, well under the 400ms
	// deadline. Stop after ~1.2s (3x the deadline). Writing to a net.Pipe blocks
	// until read, so once the connection is torn down the write fails; that is
	// fine and expected, so we ignore write results in this goroutine.
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		deadline := time.After(1200 * time.Millisecond)
		for {
			select {
			case <-deadline:
				return
			case <-ticker.C:
				var buf bytes.Buffer
				if _, err := amino.MarshalAnySizedWriter(&buf, PacketMsg{ChannelID: 0x01, EOF: 0x00}); err != nil {
					return
				}
				if _, err := server.Write(buf.Bytes()); err != nil {
					return // connection torn down — expected
				}
			}
		}
	}()

	select {
	case err := <-errorsCh:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "assembly timeout")
	case <-time.After(3 * time.Second):
		t.Fatal("deadline was reset by dribbled packets — connection was never torn down")
	}
	<-done
}

// TestRecvAssemblyTimeoutClearedOnCompletion verifies that completing a message
// (EOF=1) cancels the deadline, so a well-behaved peer that keeps sending
// complete messages is never disconnected.
func TestRecvAssemblyTimeoutClearedOnCompletion(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	receivedCh := make(chan []byte, 8)
	errorsCh := make(chan error, 1)
	onError := func(r error) {
		select {
		case errorsCh <- r:
		default:
		}
	}
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1, RecvMessageCapacity: 1 << 20},
	}, func(_ byte, b []byte) { receivedCh <- b }, onError, func() MConnConfig {
		cfg := quietPingConfig()
		cfg.RecvAssemblyTimeout = 200 * time.Millisecond
		return cfg
	}())
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	// Send several two-packet messages, each spaced well beyond the deadline.
	// Because each message completes, the deadline is cleared every time and the
	// connection must survive.
	for i := 0; i < 3; i++ {
		writePacketMsg(t, server, 0x01, 0x00, []byte("part-"))
		writePacketMsg(t, server, 0x01, 0x01, []byte("end"))
		select {
		case <-receivedCh:
		case err := <-errorsCh:
			t.Fatalf("unexpected disconnect: %v", err)
		case <-time.After(2 * time.Second):
			t.Fatal("message was not received")
		}
		time.Sleep(300 * time.Millisecond) // longer than the deadline
	}

	select {
	case err := <-errorsCh:
		t.Fatalf("connection was disconnected despite completing every message: %v", err)
	default:
	}
	require.True(t, mconn.IsRunning())
}

// TestMaxRecvBufferBudget verifies the cross-channel total recving budget: two
// channels, each individually under their own RecvMessageCapacity, together
// exceed the connection-wide budget and the connection is torn down.
func TestMaxRecvBufferBudget(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	errorsCh := make(chan error, 1)
	onError := func(r error) {
		select {
		case errorsCh <- r:
		default:
		}
	}

	cfg := quietPingConfig()
	cfg.RecvAssemblyTimeout = 30 * time.Minute // isolate the budget from the deadline
	cfg.MaxRecvBufferBytes = 100
	chDescs := []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1, RecvMessageCapacity: 80},
		{ID: 0x02, Priority: 1, SendQueueCapacity: 1, RecvMessageCapacity: 80},
	}
	mconn := NewMConnectionWithConfig(client, chDescs, func(byte, []byte) {}, onError, cfg)
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	// 60 bytes on ch 0x01 (under budget), then 50 more on ch 0x02 → 110 > 100.
	writePacketMsg(t, server, 0x01, 0x00, make([]byte, 60))
	writePacketMsg(t, server, 0x02, 0x00, make([]byte, 50))

	select {
	case err := <-errorsCh:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "total recving buffer budget exceeded")
	case <-time.After(3 * time.Second):
		t.Fatal("expected the total recving budget to be enforced")
	}
}

// writeMessage sends a single message of total bytes to the given channel,
// split into 1KB partial packets, and waits for it to be assembled.
func writeMessage(t *testing.T, w net.Conn, received <-chan []byte, chID byte, total int) {
	t.Helper()

	for sent := 0; sent < total; sent += 1024 {
		size := min(1024, total-sent)

		eof := byte(0x00)
		if sent+size >= total {
			eof = 0x01
		}

		writePacketMsg(t, w, chID, eof, make([]byte, size))
	}

	select {
	case b := <-received:
		require.Len(t, b, total)
	case <-time.After(2 * time.Second):
		t.Fatal("message was not received")
	}
}

// TestRecvingBackingArrayFreedWhenNoLongerNeeded verifies that a grown backing
// array is not pinned on the channel for the life of the connection: a channel
// that saw a single outsized message hands the memory back on the next message
// that does not need it.
//
// The release is deliberately deferred to that next message rather than done
// immediately on completion -- see TestRecvingBufferReusedForRepeatedLargeMessages
// for why.
func TestRecvingBackingArrayFreedWhenNoLongerNeeded(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	receivedCh := make(chan []byte, 1)
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1, RecvBufferCapacity: 4096, RecvMessageCapacity: 1 << 16},
	}, func(_ byte, b []byte) { receivedCh <- b }, func(error) {}, quietPingConfig())
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	ch := mconn.channelsIdx[0x01]

	// An 8KB message grows the 4KB backing array past its configured capacity.
	// It is kept, because the message that just completed was using it.
	writeMessage(t, server, receivedCh, 0x01, 8192)

	assert.Equal(t, 0, len(ch.recving), "recving should be empty after completion")
	assert.Greater(t, cap(ch.recving), ch.desc.RecvBufferCapacity,
		"the array a large message is still using must be kept, not reallocated")

	// An ordinary message no longer needs that array, so it is handed back.
	writeMessage(t, server, receivedCh, 0x01, 512)

	assert.Equal(t, 0, len(ch.recving), "recving should be empty after completion")
	assert.Equal(t, ch.desc.RecvBufferCapacity, cap(ch.recving),
		"the grown array must be released once a message no longer needs it")
	assert.Equal(t, 0, mconn.recvBufferBytes, "total recv buffer accounting should return to zero")
}

// TestRecvingBufferReusedForRepeatedLargeMessages is a regression test for the
// cost of releasing the backing array too eagerly. Releasing it on every message
// that merely grew past RecvBufferCapacity puts a realloc-and-regrow on the path
// of every such message -- and that is the steady state on the channels carrying
// the largest messages: blockchain and consensus data configure 200KB against
// multi-MB blocks, the mempool channel leaves it at the 4KB default against
// MaxTxBytes-sized txs. Measured on a 2MB message with a 200KB capacity, that
// costs 2.50ms, 9.07MB and 10 allocs, against 47.6us and no allocations here.
func TestRecvingBufferReusedForRepeatedLargeMessages(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	receivedCh := make(chan []byte, 1)
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1, RecvBufferCapacity: 4096, RecvMessageCapacity: 1 << 16},
	}, func(_ byte, b []byte) { receivedCh <- b }, func(error) {}, quietPingConfig())
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	ch := mconn.channelsIdx[0x01]

	// Every message is larger than RecvBufferCapacity, so each one would pay a
	// realloc and a regrow if the array were released on completion.
	var firstArray *byte

	for i := range 3 {
		writeMessage(t, server, receivedCh, 0x01, 8192)

		require.Equal(t, 0, len(ch.recving))

		array := &ch.recving[:1][0]
		if i == 0 {
			firstArray = array

			continue
		}

		assert.Same(t, firstArray, array,
			"a channel steadily carrying large messages must reuse its backing array")
	}

	assert.Equal(t, 0, mconn.recvBufferBytes, "total recv buffer accounting should return to zero")
}

// TestRecvingBufferReusedWhenNotGrown is the counterpart to the test above: a
// message that fits within RecvBufferCapacity must reuse the existing backing
// array rather than allocating a fresh one. Every gossip message takes this
// path, so re-allocating here would put a RecvBufferCapacity-sized allocation on
// the p2p hot path.
func TestRecvingBufferReusedWhenNotGrown(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	receivedCh := make(chan []byte, 1)
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1, RecvBufferCapacity: 4096, RecvMessageCapacity: 1 << 16},
	}, func(_ byte, b []byte) { receivedCh <- b }, func(error) {}, quietPingConfig())
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	ch := mconn.channelsIdx[0x01]

	// Send two messages that each fit comfortably inside RecvBufferCapacity, and
	// check the backing array is the same one across both.
	var firstArray *byte
	for i := range 2 {
		writePacketMsg(t, server, 0x01, 0x00, make([]byte, 512))
		writePacketMsg(t, server, 0x01, 0x01, make([]byte, 512))

		select {
		case b := <-receivedCh:
			require.Len(t, b, 1024)
		case <-time.After(2 * time.Second):
			t.Fatal("message was not received")
		}

		require.Equal(t, 0, len(ch.recving))
		require.Equal(t, ch.desc.RecvBufferCapacity, cap(ch.recving))

		array := &ch.recving[:1][0]
		if i == 0 {
			firstArray = array
			continue
		}
		assert.Same(t, firstArray, array,
			"a message that fits in RecvBufferCapacity must reuse the backing array")
	}

	assert.Equal(t, 0, mconn.recvBufferBytes, "total recv buffer accounting should return to zero")
}
