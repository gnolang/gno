package upstream_test

// secret_connection_compat_test.go: byte-format and round-trip
// verification for the upstream-compat SecretConnection used on the
// tmkms-listener path (this package's MakeSecretConnection, ported
// from cometbft v0.34).
//
// Two distinct check kinds:
//
//  1. **Old-path divergence canary** — tm2/pkg/p2p/conn's
//     SecretConnection (amino-based authSigMessage) is wire-different
//     from upstream by 2 bytes. We pin that divergence so a later
//     edit can't silently "fix" it (which would change chain p2p).
//
//  2. **New-path positive checks** — the upstream-compat
//     SecretConnection in this package speaks the v0.34 AuthSigMessage
//     shape. We assert the proto-encoded bytes match upstream byte-
//     for-byte, and that two halves successfully handshake against
//     each other end-to-end.
//
// The handshake has two pre-AEAD writes (ephemeral pubkey exchange)
// and two post-AEAD writes (auth-sig message exchange). The encoding
// MUST match upstream byte-for-byte or the handshake fails — tmkms
// would either deserialize garbage and disconnect, or deserialize
// successfully into the wrong values and produce a key mismatch.
//
// What "upstream" emits for each message:
//
//	ephemeral pubkey:
//	    protoio-delimited gogoproto BytesValue{Value: pub[:]}
//	    wire bytes: varint(34) + 0x0a + 0x20 + pub[:]
//	    (BytesValue tag 1 wire-type 2, length 32, then 32 raw bytes)
//
//	AuthSigMessage:
//	    protoio-delimited AuthSigMessage{
//	        pub_key: PublicKey{Ed25519: pub[:]},
//	        sig:     signature,
//	    }
//	    inner PublicKey: tag 1 wire-type 2, len 32 → [0x0a, 0x20, ...32 bytes...] = 34 bytes
//	    AuthSigMessage.pub_key field: tag 1 wire-type 2, len 34 → [0x0a, 0x22, ...34 bytes...] = 36 bytes
//	    AuthSigMessage.sig    field: tag 2 wire-type 2, len 64 → [0x12, 0x40, ...64 bytes...] = 66 bytes
//	    body: 102 bytes
//	    delimited: varint(102) + 102 bytes = 103 bytes
//
// What tm2 emits today:
//
//	ephemeral pubkey:
//	    amino.MarshalSizedWriter on *[32]byte (size-prefix + raw bytes)
//
//	AuthSigMessage:
//	    amino.MarshalSizedWriter on the unexported authSigMessage{Key,Sig}
//	    where Key is ed25519.PubKeyEd25519 (an amino-registered type
//	    that emits a 4-byte Disambiguator/Prefix ahead of its data).
//
// These tests capture both byte streams and ASSERT EQUALITY with the
// upstream-shape expectation. A failure here means tm2's
// SecretConnection handshake is wire-incompatible with tmkms; the
// tmkms-listener path needs either a fix to secret_connection.go or
// a separate upstream-compat handshake variant.
//
// Notes on scope:
//   - These tests do NOT spin up a real connection; they call the
//     same encoding paths used by shareEphPubKey / shareAuthSignature
//     directly with deterministic inputs.
//   - secret_connection.go's authSigMessage struct is unexported. We
//     replicate its shape locally here (same fields, same order, same
//     types) to drive amino through the identical code path. If
//     amino's reflective encoder produces the same bytes for the
//     replica as for the original, the assertion is faithful.

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// upstreamEphPubKeyBytes returns the on-the-wire representation an
// upstream Tendermint v0.34 SecretConnection writes for the ephemeral
// pubkey exchange, given a 32-byte ephemeral pubkey.
func upstreamEphPubKeyBytes(pub [32]byte) []byte {
	// gogoproto BytesValue { bytes value = 1; } body:
	//   tag 1, wire-type 2 (length-delimited) = 0x0a
	//   length varint = 0x20 (32)
	//   32 raw bytes
	body := append([]byte{0x0a, 0x20}, pub[:]...) // 34 bytes
	// protoio delimited writer prepends a varint of body length.
	out := appendVarint(nil, uint64(len(body)))
	return append(out, body...)
}

// upstreamAuthSigMessageBytes returns the on-the-wire representation
// of an AuthSigMessage{pub_key: PublicKey{Ed25519: pub}, sig: sig}
// per upstream Tendermint v0.34's protobuf schema, written via
// protoio's length-delimited writer.
func upstreamAuthSigMessageBytes(pub [32]byte, sig []byte) []byte {
	// PublicKey { oneof sum { bytes ed25519 = 1; } }
	// inner ed25519 entry: tag 1 wt 2, len 32, 32 bytes
	innerPub := append([]byte{0x0a, 0x20}, pub[:]...) // 34 bytes
	// AuthSigMessage { PublicKey pub_key = 1; bytes sig = 2; }
	body := []byte{0x0a, byte(len(innerPub))}
	body = append(body, innerPub...)
	body = append(body, 0x12)
	body = appendVarint(body, uint64(len(sig)))
	body = append(body, sig...)
	out := appendVarint(nil, uint64(len(body)))
	return append(out, body...)
}

func appendVarint(b []byte, v uint64) []byte {
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

// tm2EphPubKeyBytes returns what tm2's secret_connection.go emits
// for the ephemeral pubkey exchange, by replicating the call site
// (amino.MarshalSizedWriter on a *[32]byte).
func tm2EphPubKeyBytes(t *testing.T, pub [32]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	_, err := amino.MarshalSizedWriter(&buf, &pub)
	require.NoError(t, err)
	return buf.Bytes()
}

// tm2AuthSigMessageStandIn mirrors the unexported authSigMessage in
// tm2/pkg/p2p/conn/secret_connection.go: same field types, same
// order. Amino encodes by reflection, so the bytes match what the
// real call site produces.
type tm2AuthSigMessageStandIn struct {
	Key ed25519.PubKeyEd25519
	Sig []byte
}

func tm2AuthSigMessageBytes(t *testing.T, pub [32]byte, sig []byte) []byte {
	t.Helper()
	var pk ed25519.PubKeyEd25519
	copy(pk[:], pub[:])
	var buf bytes.Buffer
	_, err := amino.MarshalSizedWriter(&buf, tm2AuthSigMessageStandIn{Key: pk, Sig: sig})
	require.NoError(t, err)
	return buf.Bytes()
}

func TestSecretConnectionWire_EphemeralPubKey_MatchesUpstream(t *testing.T) {
	t.Parallel()

	var pub [32]byte
	for i := range pub {
		pub[i] = byte(i)
	}

	want := upstreamEphPubKeyBytes(pub)
	got := tm2EphPubKeyBytes(t, pub)

	if !bytes.Equal(want, got) {
		t.Logf("upstream-expected: %s (%d bytes)", hex.EncodeToString(want), len(want))
		t.Logf("tm2 emits:         %s (%d bytes)", hex.EncodeToString(got), len(got))
		t.Errorf("ephemeral-pubkey wire format diverges from upstream Tendermint v0.34 — " +
			"tm2 SecretConnection cannot complete a handshake with tmkms over this listener path " +
			"unless secret_connection.go is reworked to emit gogoproto BytesValue framing")
	}
}

// TestSecretConnectionWire_AuthSigMessage_KnownDivergence documents
// the 2-byte gap between tm2's AuthSigMessage encoding and upstream
// Tendermint v0.34's. tm2 amino-encodes the unexported authSigMessage
// {Key ed25519.PubKeyEd25519, Sig []byte} where Key is a registered
// ed25519 type; this emits the pubkey as a direct length-delimited
// field at tag 1 (no PublicKey oneof wrapper). Upstream wraps the
// ed25519 in a PublicKey oneof message, adding two bytes (0x0a 0x22)
// of inner framing.
//
// **Consequence**: tm2's SecretConnection auth handshake is wire-
// incompatible with tmkms. The tmkms-listener path (Phase 4-5)
// completes ephemeral-pubkey exchange and DH, then deadlocks in
// shareAuthSignature because the protobuf decoder on the tmkms side
// can't parse our amino-shaped bytes (and vice versa).
//
// Fix path (deferred to a follow-up; out of scope for Phase 6
// verification): introduce an upstream-compat handshake variant for
// the tmkms-listener path that uses gogoproto AuthSigMessage with
// the PublicKey oneof. Cannot blanket-fix tm2/pkg/p2p/conn/
// secret_connection.go without breaking p2p wire compatibility.
//
// **This test is the canary**: if anyone ever fixes the divergence
// (intentionally or by accident), this test FAILS — flagging that
// the fix path needs to be evaluated chain-wide.
func TestSecretConnectionWire_AuthSigMessage_KnownDivergence(t *testing.T) {
	t.Parallel()

	var pub [32]byte
	for i := range pub {
		pub[i] = byte(0xa0 + i%16)
	}
	sig := make([]byte, 64)
	for i := range sig {
		sig[i] = byte(0x10 + i%16)
	}

	upstreamBytes := upstreamAuthSigMessageBytes(pub, sig)
	tm2Bytes := tm2AuthSigMessageBytes(t, pub, sig)

	t.Logf("upstream-expected: %s (%d bytes)", hex.EncodeToString(upstreamBytes), len(upstreamBytes))
	t.Logf("tm2 emits:         %s (%d bytes)",  hex.EncodeToString(tm2Bytes),     len(tm2Bytes))

	// The expected divergence is exactly the missing PublicKey oneof
	// wrapper: 0x0a 0x22 ahead of the inner [0x0a, 0x20, ...32 bytes...]
	// pub_key entry, plus the corresponding 2-byte bump in the outer
	// length prefix.
	require.NotEqual(t, upstreamBytes, tm2Bytes,
		"tm2 AuthSigMessage now matches upstream — the known divergence has been fixed; "+
			"update Phase 6 docs and remove the tmkms-listener compat shim")
	require.Equal(t, len(upstreamBytes)-2, len(tm2Bytes),
		"divergence size has changed; the documented 2-byte PublicKey-oneof wrapper "+
			"may no longer fully describe the shape of the gap")
}

// TestSecretConnectionWire_AuthSigMessage_Tm2Emit pins exactly what
// bytes tm2 emits for a known input. Useful as a starting point for
// the upstream-compat shim (one needs to know what to translate
// from).
func TestSecretConnectionWire_AuthSigMessage_Tm2Emit(t *testing.T) {
	t.Parallel()

	var pub [32]byte
	for i := range pub {
		pub[i] = byte(i)
	}
	sig := make([]byte, 64)
	for i := range sig {
		sig[i] = byte(0xff - i)
	}

	got := tm2AuthSigMessageBytes(t, pub, sig)

	// Hand-derived expectation, mirroring tm2's amino on
	// authSigMessage{Key, Sig}:
	//   varint(body_len)
	//   tag 1 wt 2 (Key field), len 32, 32 raw pub bytes
	//   tag 2 wt 2 (Sig field), len 64, 64 raw sig bytes
	body := []byte{0x0a, 0x20}
	body = append(body, pub[:]...)
	body = append(body, 0x12, 0x40)
	body = append(body, sig...)
	want := appendVarint(nil, uint64(len(body)))
	want = append(want, body...)

	assert.Equal(t, want, got,
		"tm2's amino encoding of authSigMessage{Key ed25519.PubKeyEd25519, Sig []byte} drifted")
}

// TestUpstreamSecretConnection_AuthSigMessage_MatchesUpstream checks
// the NEW path: the upstreampb.AuthSigMessage proto used by this
// package's MakeSecretConnection emits bytes byte-identical to what
// upstream Tendermint v0.34's tendermint.p2p.AuthSigMessage emits.
// Combined with the self-handshake test below, this is the
// positive-side proof that the listener path is wire-compatible
// with tmkms.
func TestUpstreamSecretConnection_AuthSigMessage_MatchesUpstream(t *testing.T) {
	t.Parallel()

	var pub [32]byte
	for i := range pub {
		pub[i] = byte(0xa0 + i%16)
	}
	sig := make([]byte, 64)
	for i := range sig {
		sig[i] = byte(0x10 + i%16)
	}

	// What tm2's NEW path emits: upstreampb.AuthSigMessage marshaled
	// via google.golang.org/protobuf/proto.Marshal, length-delimited
	// the way protoio does on the wire.
	asm := &upstreampb.AuthSigMessage{
		PubKey: &upstreampb.PublicKey{
			Sum: &upstreampb.PublicKey_Ed25519{Ed25519: pub[:]},
		},
		Sig: sig,
	}
	body, err := proto.Marshal(asm)
	require.NoError(t, err)
	got := append(appendVarint(nil, uint64(len(body))), body...)

	want := upstreamAuthSigMessageBytes(pub, sig)

	if !bytes.Equal(want, got) {
		t.Logf("upstream-expected: %s (%d bytes)", hex.EncodeToString(want), len(want))
		t.Logf("upstreampb emits:  %s (%d bytes)", hex.EncodeToString(got), len(got))
	}
	assert.Equal(t, want, got,
		"upstream-compat AuthSigMessage encoding must match upstream Tendermint v0.34 byte-for-byte")
}

// TestUpstreamSecretConnection_SelfHandshake runs MakeSecretConnection
// from both ends of an io.Pipe, asserts both succeed, and exercises
// the encrypted Read/Write loop. Implicitly proves the entire
// adapted handshake (Merlin transcript + HKDF + AuthSigMessage) is
// internally consistent.
func TestUpstreamSecretConnection_SelfHandshake(t *testing.T) {
	t.Parallel()

	fooConn, barConn := pipeConnPair()
	defer fooConn.Close()
	defer barConn.Close()

	fooPriv := ed25519.GenPrivKey()
	barPriv := ed25519.GenPrivKey()

	type out struct {
		sc  *upstream.SecretConnection
		err error
	}
	fooCh := make(chan out, 1)
	barCh := make(chan out, 1)
	go func() {
		sc, err := upstream.MakeSecretConnection(fooConn, fooPriv)
		fooCh <- out{sc, err}
	}()
	go func() {
		sc, err := upstream.MakeSecretConnection(barConn, barPriv)
		barCh <- out{sc, err}
	}()

	fooRes := <-fooCh
	barRes := <-barCh
	require.NoError(t, fooRes.err, "foo handshake failed")
	require.NoError(t, barRes.err, "bar handshake failed")

	// Each side should know the other's consensus pubkey.
	assert.True(t, fooRes.sc.RemotePubKey().Equals(barPriv.PubKey()),
		"foo's view of remote pubkey must match bar's actual pubkey")
	assert.True(t, barRes.sc.RemotePubKey().Equals(fooPriv.PubKey()),
		"bar's view of remote pubkey must match foo's actual pubkey")

	// Round-trip a payload.
	want := []byte("hello from upstream-compat secret_connection")
	go func() { _, _ = fooRes.sc.Write(want) }()
	got := make([]byte, len(want))
	_, err := io.ReadFull(barRes.sc, got)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

// pipeConnPair returns two halves of an in-memory connection that
// satisfy io.ReadWriteCloser. SecretConnection only needs the
// io.ReadWriteCloser surface for the handshake; net.Conn-typed
// methods are not exercised in this test.
type pipeConn struct {
	*io.PipeReader
	*io.PipeWriter
}

func (p pipeConn) Close() error {
	_ = p.PipeReader.Close()
	return p.PipeWriter.Close()
}

func pipeConnPair() (a, b pipeConn) {
	aReader, bWriter := io.Pipe()
	bReader, aWriter := io.Pipe()
	return pipeConn{aReader, aWriter}, pipeConn{bReader, bWriter}
}

// TestSecretConnectionWire_VarintHelperSelfCheck guards the helper
// upstreamEphPubKeyBytes / upstreamAuthSigMessageBytes use.
func TestSecretConnectionWire_VarintHelperSelfCheck(t *testing.T) {
	t.Parallel()
	// 34 = 0x22 (single byte)
	require.Equal(t, []byte{0x22}, appendVarint(nil, 34))
	// 102 = 0x66 (single byte)
	require.Equal(t, []byte{0x66}, appendVarint(nil, 102))
	// 128 = 0x80 0x01 (two bytes — sets the continuation bit)
	require.Equal(t, []byte{0x80, 0x01}, appendVarint(nil, 128))
	// Cross-check vs binary.AppendUvarint (Go 1.19+) for sanity.
	for _, v := range []uint64{0, 1, 127, 128, 16383, 16384, 65535, 1 << 20} {
		got := appendVarint(nil, v)
		want := binary.AppendUvarint(nil, v)
		assert.Equal(t, want, got, "varint(%d) mismatch", v)
	}
}
