package upstream_test

// msgs_test.go: verify the WrapMsg/UnwrapMsg envelope helpers and end-to-end
// round-trip via google.golang.org/protobuf. This is the load-bearing test
// for what was previously Phase 3 BUG-1 (Message envelope was not a real
// oneof) and BUG-2 (PubKey emitted Any/TypeURL bytes, not upstream's
// PublicKey oneof). Both are fixed by routing through upstreampb.

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// WrapMsg + Marshal + Unmarshal + UnwrapMsg round-trip yields the same
// concrete message back. Exercise each Sum variant.
func TestMsgs_Wrap_Unwrap_RoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   interface{}
	}{
		{"PubKeyRequest", &upstreampb.PubKeyRequest{ChainId: "test"}},
		{"SignVoteRequest", &upstreampb.SignVoteRequest{Vote: &upstreampb.Vote{Height: 42}, ChainId: "test"}},
		{"SignProposalRequest", &upstreampb.SignProposalRequest{Proposal: &upstreampb.Proposal{Height: 1, PolRound: -1}, ChainId: "test"}},
		{"PingRequest", &upstreampb.PingRequest{}},
		{"PingResponse", &upstreampb.PingResponse{}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			env := upstream.WrapMsg(c.in)
			bz, err := proto.Marshal(env)
			require.NoError(t, err)

			var got upstreampb.Message
			require.NoError(t, proto.Unmarshal(bz, &got))

			out, err := upstream.UnwrapMsg(&got)
			require.NoError(t, err)
			// Compare via proto.Equal for protobuf message identity.
			assert.True(t, proto.Equal(c.in.(proto.Message), out.(proto.Message)),
				"round-trip mismatch: in=%v out=%v", c.in, out)
		})
	}
}

// Empty Message envelope decodes to a nil Sum and UnwrapMsg returns an
// error rather than a nil-pointer-deref panic.
func TestMsgs_UnwrapMsg_EmptyRejected(t *testing.T) {
	t.Parallel()
	_, err := upstream.UnwrapMsg(nil)
	require.Error(t, err)
	_, err = upstream.UnwrapMsg(&upstreampb.Message{})
	require.Error(t, err)
}

// Sentinel: only one Sum branch is encoded per Message. Build a Message
// with PubKeyRequest set, marshal/unmarshal, confirm only that branch is
// populated. (proto3 oneof guarantees this; we lock it in as a regression
// test against any future hand-rolled marshaler.)
func TestMsgs_OneofExclusive(t *testing.T) {
	t.Parallel()
	in := upstream.WrapMsg(&upstreampb.PingRequest{})
	bz, err := proto.Marshal(in)
	require.NoError(t, err)

	var out upstreampb.Message
	require.NoError(t, proto.Unmarshal(bz, &out))
	// Exactly the Ping_Request variant should be set.
	_, ok := out.Sum.(*upstreampb.Message_PingRequest)
	assert.True(t, ok, "expected Sum to be Message_PingRequest, got %T", out.Sum)
}

// ---- PubKey conversion (the BUG-2 fix) -----------------------------------

func TestPubKeyToProto_Ed25519_RoundTrip(t *testing.T) {
	t.Parallel()
	priv := ed25519.GenPrivKey()
	pk := priv.PubKey()

	pbk, err := upstream.PubKeyToProto(pk)
	require.NoError(t, err)

	// Wire-encode. Field 1 = Ed25519 (32 bytes).
	bz, err := proto.Marshal(pbk)
	require.NoError(t, err)

	var dec upstreampb.PublicKey
	require.NoError(t, proto.Unmarshal(bz, &dec))

	got, err := upstream.PubKeyFromProto(&dec)
	require.NoError(t, err)
	assert.Equal(t, pk.Bytes(), got.Bytes())

	// Specifically: bytes do NOT contain a TypeURL (`/tm.PubKeyEd25519`).
	// That was BUG-2 — amino-encoded crypto.PubKey emitted Any-shape.
	// upstreampb.PublicKey emits exactly: `0a 20 <32 bytes>` (~34 bytes).
	assert.Less(t, len(bz), 50, "PublicKey wire encoding must be ~34 bytes (no Any/TypeURL)")
	assert.NotContains(t, string(bz), "/tm.PubKey", "must not contain amino-style TypeURL")
}

func TestPubKeyToProto_Secp256k1_RoundTrip(t *testing.T) {
	t.Parallel()
	priv := secp256k1.GenPrivKey()
	pk := priv.PubKey()

	pbk, err := upstream.PubKeyToProto(pk)
	require.NoError(t, err)

	bz, err := proto.Marshal(pbk)
	require.NoError(t, err)

	var dec upstreampb.PublicKey
	require.NoError(t, proto.Unmarshal(bz, &dec))

	got, err := upstream.PubKeyFromProto(&dec)
	require.NoError(t, err)
	assert.Equal(t, pk.Bytes(), got.Bytes())
}

func TestPubKeyFromProto_NilOrEmpty(t *testing.T) {
	t.Parallel()
	_, err := upstream.PubKeyFromProto(nil)
	require.Error(t, err)
	_, err = upstream.PubKeyFromProto(&upstreampb.PublicKey{})
	require.Error(t, err)
}
