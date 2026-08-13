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
// attacker keeps sending partial packets (here, empty ones — the exact dribble
// the earlier proposed fix was vulnerable to) to try to keep resetting the
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

// TestRecvingBackingArrayFreedOnCompletion verifies that a completed message
// does not pin its grown backing array on the channel: after receiving a
// message larger than RecvBufferCapacity, the recving slice is reallocated back
// to RecvBufferCapacity rather than retaining the large array.
func TestRecvingBackingArrayFreedOnCompletion(t *testing.T) {
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

	// A 8KB message in 1KB chunks grows the 4KB backing array past its initial cap.
	const total = 8192
	for sent := 0; sent < total; sent += 1024 {
		eof := byte(0x00)
		if sent+1024 >= total {
			eof = 0x01
		}
		writePacketMsg(t, server, 0x01, eof, make([]byte, 1024))
	}

	select {
	case b := <-receivedCh:
		require.Len(t, b, total)
	case <-time.After(2 * time.Second):
		t.Fatal("message was not received")
	}

	ch := mconn.channelsIdx[0x01]
	assert.Equal(t, 0, len(ch.recving), "recving should be empty after completion")
	assert.Equal(t, ch.desc.RecvBufferCapacity, cap(ch.recving),
		"backing array should be reallocated to RecvBufferCapacity, not the grown capacity")
	assert.Equal(t, 0, mconn.recvBufferBytes, "total recv buffer accounting should return to zero")
}
