package upstream

// secret_connection_padding_test.go pins PR #5717 review finding 2:
//
// Write frames each chunk into a buffer taken from the shared global
// go-buffer-pool (pool.Get), which is NOT zeroed. It writes only the 4-byte
// length prefix and chunkLength bytes of payload, but sendAead.Seal encrypts
// the *entire* scTotalFrameSize frame. For any chunk shorter than
// scDataMaxSize, the tail [scDataLenSize+chunkLength:] carries whatever stale
// plaintext was last in that pooled buffer — and because the pool is shared
// process-wide (including with tm2/pkg/p2p/conn's chain-p2p SecretConnection),
// that can be unrelated plaintext, which then gets sealed and transmitted to
// the authenticated KMS peer as frame padding.
//
// The fix clears the unused tail before Seal. This test primes the pool with a
// sentinel pattern, performs a short Write, decrypts the captured frame, and
// asserts the padding is all zero. It FAILS on the pre-fix code (padding ==
// sentinel) and PASSES once the tail is cleared.

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	pool "github.com/libp2p/go-buffer-pool"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"
)

// captureConn is an io.ReadWriteCloser whose Write appends to a buffer, so the
// test can inspect the exact sealed bytes Write puts on the wire.
type captureConn struct{ buf *bytes.Buffer }

func (c captureConn) Write(p []byte) (int, error) { return c.buf.Write(p) }
func (captureConn) Read([]byte) (int, error)      { return 0, io.EOF }
func (captureConn) Close() error                  { return nil }

func TestSecretConnection_Write_ZeroesFramePadding(t *testing.T) {
	t.Parallel()

	// Prime the shared pool's bucket for scTotalFrameSize with a non-zero
	// sentinel so a leaking Write would seal the sentinel into the padding.
	// (On the fixed code the buffer is cleared regardless, so this test never
	// flakes on correct code; priming only lets it catch a regression.)
	const sentinel = 0xAA
	primed := make([][]byte, 0, 8)
	for range 8 {
		b := pool.Get(scTotalFrameSize)
		for i := range b {
			b[i] = sentinel
		}
		primed = append(primed, b)
	}
	for _, b := range primed {
		pool.Put(b)
	}

	// A SecretConnection with a known AEAD and an all-zero send nonce, writing
	// into a capture buffer. We only exercise the send path.
	key := make([]byte, scAEADKeySize)
	aead, err := chacha20poly1305.New(key)
	require.NoError(t, err)

	var wire bytes.Buffer
	sc := &SecretConnection{
		sendAead:  aead,
		sendNonce: new([scAEADNonceSize]byte),
		conn:      captureConn{&wire},
	}

	// Snapshot the nonce Seal will use (Write increments it afterwards).
	var nonce [scAEADNonceSize]byte
	copy(nonce[:], sc.sendNonce[:])

	chunk := []byte{0x01, 0x02, 0x03}
	n, err := sc.Write(chunk)
	require.NoError(t, err)
	require.Equal(t, len(chunk), n)

	// Decrypt the single sealed frame off the wire.
	plain, err := aead.Open(nil, nonce[:], wire.Bytes(), nil)
	require.NoError(t, err)
	require.Len(t, plain, scTotalFrameSize)

	// Length prefix and payload are intact...
	require.Equal(t, uint32(len(chunk)), binary.LittleEndian.Uint32(plain[:scDataLenSize]))
	require.Equal(t, chunk, plain[scDataLenSize:scDataLenSize+len(chunk)])

	// ...and the padding tail must be zero, not stale pooled plaintext.
	padding := plain[scDataLenSize+len(chunk):]
	require.Equal(t, make([]byte, len(padding)), padding,
		"frame padding leaked stale pooled plaintext instead of being zeroed")
}
