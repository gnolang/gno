package upstream_test

// Layer 1 tests: protowire walks of amino-emitted bytes.
//
// For each upstream-shaped type, encode a value via tm2 amino and decode
// the resulting bytes with google.golang.org/protobuf/encoding/protowire
// (the stdlib protobuf low-level wire walker). Assert that:
//   - Each field appears at the expected field number
//   - Each field's wire-type matches upstream Tendermint v0.34's .proto schema
//   - Each value round-trips through the protowire decoder
//
// This is the lightest of the three test layers: no .pb.go fixtures
// needed, no live tmkms binary. It catches the most common regressions
// (wrong wire-type from a tag drift, wrong field number from struct field
// reordering) at unit-test cost.

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
)

// codec returns a fresh sealed codec with the upstream package registered.
func codec(t *testing.T) *amino.Codec {
	t.Helper()
	cdc := amino.NewCodec()
	cdc.RegisterPackage(upstream.Package)
	cdc.Seal()
	return cdc
}

// fieldDesc captures the expected wire shape of one field.
type fieldDesc struct {
	num  protowire.Number
	typ  protowire.Type
	name string
}

// walkFields decodes a protobuf-shaped byte stream and returns each field's
// number/type/value-bytes in encounter order. Errors fail the test.
func walkFields(t *testing.T, bz []byte) []struct {
	num  protowire.Number
	typ  protowire.Type
	val  []byte
	uval uint64 // value for varint/i32/i64; raw bytes-length for length-delimited
} {
	t.Helper()
	var out []struct {
		num  protowire.Number
		typ  protowire.Type
		val  []byte
		uval uint64
	}
	cur := bz
	for len(cur) > 0 {
		num, typ, n := protowire.ConsumeTag(cur)
		require.Greater(t, n, 0, "ConsumeTag failed at %x", cur)
		cur = cur[n:]

		var entry struct {
			num  protowire.Number
			typ  protowire.Type
			val  []byte
			uval uint64
		}
		entry.num = num
		entry.typ = typ

		switch typ {
		case protowire.VarintType:
			u, m := protowire.ConsumeVarint(cur)
			require.Greater(t, m, 0)
			entry.uval = u
			entry.val = cur[:m]
			cur = cur[m:]
		case protowire.Fixed32Type:
			u, m := protowire.ConsumeFixed32(cur)
			require.Greater(t, m, 0)
			entry.uval = uint64(u)
			entry.val = cur[:m]
			cur = cur[m:]
		case protowire.Fixed64Type:
			u, m := protowire.ConsumeFixed64(cur)
			require.Greater(t, m, 0)
			entry.uval = u
			entry.val = cur[:m]
			cur = cur[m:]
		case protowire.BytesType:
			b, m := protowire.ConsumeBytes(cur)
			require.Greater(t, m, 0)
			entry.uval = uint64(len(b))
			entry.val = b
			cur = cur[m:]
		default:
			t.Fatalf("unsupported wire-type %v at field %d", typ, num)
		}
		out = append(out, entry)
	}
	return out
}

// assertSchema checks that the fields encountered in bz are a (possibly
// proper) subset of want. amino skips zero-value fields, so a struct field
// being absent is allowed; but anything PRESENT must match the expected
// (num, typ).
func assertSchema(t *testing.T, bz []byte, want []fieldDesc) {
	t.Helper()
	got := walkFields(t, bz)
	wantByNum := map[protowire.Number]fieldDesc{}
	for _, w := range want {
		wantByNum[w.num] = w
	}
	for _, g := range got {
		w, ok := wantByNum[g.num]
		assert.True(t, ok, "unexpected field number %d", g.num)
		if ok {
			assert.Equal(t, w.typ, g.typ,
				"field %d (%s): wire-type mismatch — got %v, want %v",
				g.num, w.name, g.typ, w.typ)
		}
	}
}

// ---- PartSetHeader ---------------------------------------------------------
//
// upstream's types.proto:
//
//	message PartSetHeader {
//	  uint32 total = 1;
//	  bytes  hash  = 2;
//	}
func TestUpstream_PartSetHeader_Schema(t *testing.T) {
	t.Parallel()
	cdc := codec(t)
	v := upstream.PartSetHeader{Total: 42, Hash: []byte{0xab, 0xcd}}
	bz, err := cdc.Marshal(&v)
	require.NoError(t, err)

	assertSchema(t, bz, []fieldDesc{
		{num: 1, typ: protowire.VarintType, name: "total"},
		{num: 2, typ: protowire.BytesType, name: "hash"},
	})

	// Verify field 1 is varint and decodes to 42.
	got := walkFields(t, bz)
	require.Len(t, got, 2)
	assert.Equal(t, uint64(42), got[0].uval, "total = 42")
	assert.Equal(t, []byte{0xab, 0xcd}, got[1].val, "hash bytes")
}

// ---- BlockID --------------------------------------------------------------
//
//	message BlockID {
//	  bytes         hash             = 1;
//	  PartSetHeader part_set_header  = 2;
//	}
func TestUpstream_BlockID_Schema(t *testing.T) {
	t.Parallel()
	cdc := codec(t)
	v := upstream.BlockID{
		Hash:          []byte{0x01, 0x02, 0x03},
		PartSetHeader: upstream.PartSetHeader{Total: 7, Hash: []byte{0xff}},
	}
	bz, err := cdc.Marshal(&v)
	require.NoError(t, err)

	assertSchema(t, bz, []fieldDesc{
		{num: 1, typ: protowire.BytesType, name: "hash"},
		{num: 2, typ: protowire.BytesType, name: "part_set_header"},
	})
}

// ---- Vote ----------------------------------------------------------------
//
//	message Vote {
//	  SignedMsgType type              = 1;  // varint
//	  int64         height            = 2;  // varint (plain, NOT zigzag)
//	  int32         round             = 3;  // varint (plain)
//	  BlockID       block_id          = 4;  // length-delimited
//	  google.protobuf.Timestamp ts   = 5;  // length-delimited
//	  bytes         validator_address = 6;  // length-delimited
//	  int32         validator_index   = 7;  // varint (plain)
//	  bytes         signature         = 8;  // length-delimited
//	}
func TestUpstream_Vote_Schema(t *testing.T) {
	t.Parallel()
	cdc := codec(t)
	v := upstream.Vote{
		Type:             types.PrecommitType,
		Height:           42,
		Round:            3,
		BlockID:          upstream.BlockID{Hash: []byte{0xaa}},
		Timestamp:        time.Unix(1700000000, 0).UTC(),
		ValidatorAddress: make([]byte, 20),
		ValidatorIndex:   5,
		Signature:        []byte{0xde, 0xad, 0xbe, 0xef},
	}
	bz, err := cdc.Marshal(&v)
	require.NoError(t, err)

	assertSchema(t, bz, []fieldDesc{
		{num: 1, typ: protowire.VarintType, name: "type"},
		{num: 2, typ: protowire.VarintType, name: "height"},
		{num: 3, typ: protowire.VarintType, name: "round"},
		{num: 4, typ: protowire.BytesType, name: "block_id"},
		{num: 5, typ: protowire.BytesType, name: "timestamp"},
		{num: 6, typ: protowire.BytesType, name: "validator_address"},
		{num: 7, typ: protowire.VarintType, name: "validator_index"},
		{num: 8, typ: protowire.BytesType, name: "signature"},
	})
}

// Negative round encodes as a 10-byte plain varint (sign-extended through
// uint64). If the field were emitting zigzag (sint32), -1 would be 1 byte
// (0x01) — this test catches that regression.
func TestUpstream_Vote_NegativeRoundPlainVarint(t *testing.T) {
	t.Parallel()
	cdc := codec(t)
	v := upstream.Vote{
		Type:             types.PrecommitType,
		Height:           1,
		Round:            -1,
		ValidatorAddress: make([]byte, 20),
	}
	bz, err := cdc.Marshal(&v)
	require.NoError(t, err)

	got := walkFields(t, bz)
	for _, g := range got {
		if g.num == 3 { // round
			assert.Equal(t, protowire.VarintType, g.typ)
			// Plain varint of int32(-1) sign-extends to int64(-1) →
			// uint64(0xFFFFFFFFFFFFFFFF) → 10 bytes.
			assert.Equal(t, 10, len(g.val), "round=-1 must be 10 bytes plain varint, not zigzag")
			assert.Equal(t, uint64(0xFFFFFFFFFFFFFFFF), g.uval)
			return
		}
	}
	t.Fatal("round field not found in encoded bytes")
}

// ---- Proposal ------------------------------------------------------------
//
//	message Proposal {
//	  SignedMsgType type      = 1; // varint
//	  int64         height    = 2; // varint (plain)
//	  int32         round     = 3; // varint (plain)
//	  int32         pol_round = 4; // varint (plain)
//	  BlockID       block_id  = 5; // length-delimited
//	  google.protobuf.Timestamp ts = 6; // length-delimited
//	  bytes         signature = 7; // length-delimited
//	}
func TestUpstream_Proposal_Schema(t *testing.T) {
	t.Parallel()
	cdc := codec(t)
	v := upstream.Proposal{
		Type:      types.ProposalType,
		Height:    100,
		Round:     2,
		POLRound:  -1,
		BlockID:   upstream.BlockID{Hash: []byte{0x11}},
		Timestamp: time.Unix(1700000000, 0).UTC(),
		Signature: []byte{0xca, 0xfe},
	}
	bz, err := cdc.Marshal(&v)
	require.NoError(t, err)

	assertSchema(t, bz, []fieldDesc{
		{num: 1, typ: protowire.VarintType, name: "type"},
		{num: 2, typ: protowire.VarintType, name: "height"},
		{num: 3, typ: protowire.VarintType, name: "round"},
		{num: 4, typ: protowire.VarintType, name: "pol_round"},
		{num: 5, typ: protowire.BytesType, name: "block_id"},
		{num: 6, typ: protowire.BytesType, name: "timestamp"},
		{num: 7, typ: protowire.BytesType, name: "signature"},
	})

	// pol_round = -1: 10 bytes plain varint.
	got := walkFields(t, bz)
	for _, g := range got {
		if g.num == 4 {
			assert.Equal(t, protowire.VarintType, g.typ)
			assert.Equal(t, 10, len(g.val), "pol_round=-1 must be plain varint")
			assert.Equal(t, uint64(0xFFFFFFFFFFFFFFFF), g.uval)
			return
		}
	}
	t.Fatal("pol_round field not found")
}

// ---- Translator round-trip ------------------------------------------------

func TestTranslator_Vote_RoundTrip(t *testing.T) {
	t.Parallel()
	orig := &types.Vote{
		Type:             types.PrecommitType,
		Height:           42,
		Round:            3,
		BlockID:          types.BlockID{Hash: []byte{0xaa}},
		Timestamp:        time.Unix(1700000000, 0).UTC(),
		ValidatorAddress: makeAddr(t, 0x55),
		ValidatorIndex:   7,
		Signature:        []byte{0xde, 0xad},
	}

	up := upstream.FromTM2Vote(orig)
	require.NotNil(t, up)
	assert.Equal(t, orig.Type, up.Type)
	assert.Equal(t, orig.Height, up.Height)
	assert.Equal(t, int32(orig.Round), up.Round)
	assert.Equal(t, int32(orig.ValidatorIndex), up.ValidatorIndex)
	assert.Equal(t, orig.ValidatorAddress[:], up.ValidatorAddress)

	back := upstream.ToTM2Vote(up)
	require.NotNil(t, back)
	assert.Equal(t, *orig, *back)
}

func TestTranslator_Proposal_RoundTrip(t *testing.T) {
	t.Parallel()
	orig := &types.Proposal{
		Type:      types.ProposalType,
		Height:    100,
		Round:     2,
		POLRound:  -1,
		BlockID:   types.BlockID{Hash: []byte{0x11}, PartsHeader: types.PartSetHeader{Total: 5, Hash: []byte{0xff}}},
		Timestamp: time.Unix(1700000000, 0).UTC(),
		Signature: []byte{0xca, 0xfe},
	}

	up := upstream.FromTM2Proposal(orig)
	back := upstream.ToTM2Proposal(up)
	assert.Equal(t, *orig, *back)
}

func TestTranslator_NegativeTotalRejected(t *testing.T) {
	t.Parallel()
	psh := types.PartSetHeader{Total: -1, Hash: []byte{0xff}}
	assert.Panics(t, func() { upstream.FromTM2PartSetHeader(psh) },
		"negative Total must panic — upstream uint32 cannot represent it")
}

// Privval message schema tests removed: privval-protocol messages
// (PubKeyRequest, SignVoteRequest, Message envelope, etc.) are now
// protoc-generated upstreampb types, not amino-encoded. Their wire-format
// is correct by construction (proto.Marshal). Tests for those live in
// stdlibproto_test.go (round-trip via protobuf-go) or msgs_test.go
// (wrap/unwrap via WrapMsg/UnwrapMsg).

// ---- helpers --------------------------------------------------------------

func makeAddr(t *testing.T, b byte) types.Address {
	t.Helper()
	var addr types.Address
	for i := range addr {
		addr[i] = b
	}
	return addr
}
