package conn

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// writeRawPacketMsg writes a PacketMsg without asserting on the result, for
// writers that are expected to block (and possibly fail) once we stop reading.
func writeRawPacketMsg(w net.Conn, chID, eof byte, data []byte) error {
	var buf bytes.Buffer
	if _, err := amino.MarshalAnySizedWriter(&buf, PacketMsg{ChannelID: chID, EOF: eof, Bytes: data}); err != nil {
		return err
	}

	_, err := w.Write(buf.Bytes())

	return err
}

// TestRecvAssemblyDeadlineNotChargedForReactorStall is the counterpart to
// TestRecvAssemblyTimeoutNotResetByDribble: the deadline must fire for delay the
// peer causes, and must NOT fire for delay we cause ourselves.
//
// recvRoutine calls onReceive inline, and real reactors block there for
// unbounded time -- the mempool's Receive waits on the same mutex ApplyBlock
// holds across commit and recheck of the whole mempool, and the consensus
// reactor's does a blocking send into a peerMsgQueue that receiveRoutine may not
// be draining. While that lasts nothing on the connection is read, so an
// honest peer half way through a message on another channel would be torn down
// for a stall on our side.
func TestRecvAssemblyDeadlineNotChargedForReactorStall(t *testing.T) {
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

	const (
		assemblyTimeout = 300 * time.Millisecond
		reactorStall    = 4 * assemblyTimeout
	)

	// Channel 0x01's reactor blocks for well over the assembly deadline, once.
	var once sync.Once
	receivedCh := make(chan []byte, 2)
	onReceive := func(chID byte, msgBytes []byte) {
		if chID == 0x01 {
			once.Do(func() { time.Sleep(reactorStall) })
		}

		receivedCh <- msgBytes
	}

	cfg := quietPingConfig()
	cfg.RecvAssemblyTimeout = assemblyTimeout

	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1, RecvMessageCapacity: 4096},
		{ID: 0x02, Priority: 1, SendQueueCapacity: 1, RecvMessageCapacity: 4096},
	}, onReceive, onError, cfg)
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	// An honest peer: it opens a message on 0x02, hands us a complete message on
	// 0x01 (which parks recvRoutine in the slow reactor), then finishes 0x02.
	// Its writes block while we are not reading, exactly like TCP backpressure.
	writeDone := make(chan error, 1)
	go func() {
		for _, p := range []struct {
			chID, eof byte
			data      []byte
		}{
			{0x02, 0x00, []byte("part-1")},
			{0x01, 0x01, []byte("whole")},
			{0x02, 0x00, []byte("part-2")},
			{0x02, 0x01, []byte("part-3")},
		} {
			if err := writeRawPacketMsg(server, p.chID, p.eof, p.data); err != nil {
				writeDone <- err

				return
			}
		}

		writeDone <- nil
	}()

	// Both messages must arrive, and the connection must survive.
	var got [][]byte
	for len(got) < 2 {
		select {
		case b := <-receivedCh:
			got = append(got, b)
		case err := <-errorsCh:
			t.Fatalf("connection torn down by our own reactor stall: %v", err)
		case <-time.After(reactorStall + 5*time.Second):
			t.Fatal("messages were not delivered")
		}
	}

	assert.Equal(t, []byte("whole"), got[0])
	assert.Equal(t, []byte("part-1part-2part-3"), got[1],
		"the peer's message must complete, not be cut short")

	require.NoError(t, <-writeDone)

	// And the deadline still works afterwards: a fresh message that the peer
	// never finishes is dropped as before.
	require.NoError(t, writeRawPacketMsg(server, 0x02, 0x00, []byte("orphan")))

	select {
	case err := <-errorsCh:
		assert.Contains(t, err.Error(), "assembly timeout")
	case <-time.After(5 * time.Second):
		t.Fatal("the assembly deadline stopped working after a credited stall")
	}
}
