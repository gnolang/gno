package upstream_test

// protoio_overread_test.go pins PR #5717 review finding 1:
//
// NewDelimitedReader wraps the underlying conn in a *bufio.Reader and is
// created fresh per message, then discarded. Reading the varint length
// prefix calls bufio.fill(), which pulls up to 4KB off the conn in a single
// Read. When the peer has already written the *next* frame (TCP coalescing —
// both frames sit in the kernel socket buffer before we read), those bytes
// are slurped into the bufio buffer and thrown away when the reader is
// dropped. The next read stage reads directly off the raw conn (in the real
// handshake, SecretConnection.Read does io.ReadFull(sc.conn, ...) for the
// sealed AuthSig frame) and starves.
//
// This is exactly the shareEphPubKey -> shareAuthSignature sequence in
// secret_connection.go: shareEphPubKey reads the 32-byte ephemeral pubkey
// frame with a throwaway DelimitedReader over the raw conn, then the AuthSig
// frame is read off the conn directly. Over io.Pipe (the existing tests) each
// Write is delivered to its own Read so the bug is invisible; over real TCP
// it manifests as intermittent handshake failures.
//
// Both subtests below FAIL on the current bufio-based implementation and PASS
// once ReadMsg consumes exactly one message's bytes without reading ahead.

import (
	"bytes"
	"io"
	"net"
	"slices"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// makeEphFrame builds a varint-length-prefixed BytesValue frame exactly as
// shareEphPubKey writes the ephemeral pubkey on the wire.
func makeEphFrame(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	var eph [32]byte
	for i := range eph {
		eph[i] = byte(i + 1) // arbitrary non-zero, irrelevant to framing
	}
	if _, err := upstream.NewDelimitedWriter(&buf).WriteMsg(&wrapperspb.BytesValue{Value: eph[:]}); err != nil {
		t.Fatalf("frame eph pubkey: %v", err)
	}
	return buf.Bytes()
}

// TestDelimitedReader_DoesNotOverReadIntoNextFrame_InMemory is the
// deterministic core proof. The underlying reader hands back the eph frame
// AND the following bytes in a single Read (models a coalesced TCP segment);
// after reading exactly one message, the trailing bytes must still be
// retrievable from the same reader.
func TestDelimitedReader_DoesNotOverReadIntoNextFrame_InMemory(t *testing.T) {
	t.Parallel()

	ephFrame := makeEphFrame(t)
	// Stand-in for the start of the sealed AuthSig frame that, in the real
	// handshake, is read off the conn directly by SecretConnection.Read.
	next := []byte("NEXT-FRAME-BYTES-MUST-SURVIVE-THE-FIRST-READ")

	combined := slices.Concat(ephFrame, next)
	// bytes.Reader satisfies one Read with as many bytes as fit in the
	// caller's buffer, so bufio.fill() pulls the whole thing at once —
	// exactly the coalescing case.
	r := bytes.NewReader(combined)

	var bv wrapperspb.BytesValue
	n, err := upstream.NewDelimitedReader(r, 1024*1024).ReadMsg(&bv)
	require.NoError(t, err)
	require.Len(t, bv.Value, 32, "ephemeral pubkey should round-trip")
	require.Equal(t, len(ephFrame), n, "ReadMsg should report consuming exactly the eph frame")

	// The next stage reads directly from the same underlying reader.
	got := make([]byte, len(next))
	_, err = io.ReadFull(r, got)
	require.NoError(t, err, "trailing frame bytes were swallowed by the discarded bufio buffer")
	require.Equal(t, next, got, "trailing frame bytes must survive the first ReadMsg")
}

// TestDelimitedReader_DoesNotOverReadIntoNextFrame_TCP reproduces the same
// failure over a real loopback TCP connection: the peer writes the eph frame
// and the next chunk back-to-back before we read, so both land in the socket
// buffer and bufio.fill() coalesces them.
func TestDelimitedReader_DoesNotOverReadIntoNextFrame_TCP(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	ephFrame := makeEphFrame(t)
	next := []byte("NEXT-FRAME-BYTES-MUST-SURVIVE-THE-FIRST-READ")

	written := make(chan struct{})
	go func() {
		c, aerr := ln.Accept()
		if aerr != nil {
			return
		}
		defer c.Close()
		// One Write of both frames -> single coalesced segment on loopback.
		_, _ = c.Write(slices.Concat(ephFrame, next))
		close(written)
		// Hold the conn open so the reader side sees the buffered bytes
		// rather than an EOF-driven flush.
		time.Sleep(time.Second)
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	// Ensure both frames are in our socket buffer before the first read.
	<-written
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))

	var bv wrapperspb.BytesValue
	_, err = upstream.NewDelimitedReader(conn, 1024*1024).ReadMsg(&bv)
	require.NoError(t, err)
	require.Len(t, bv.Value, 32)

	got := make([]byte, len(next))
	_, err = io.ReadFull(conn, got)
	require.NoError(t, err, "trailing frame bytes were swallowed by the discarded bufio buffer")
	require.Equal(t, next, got, "trailing frame bytes must survive the first ReadMsg")
}
