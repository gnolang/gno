package upstream_test

// Layer 2 / Layer 3 hybrid: hand-build the expected upstream-protobuf wire
// bytes for each type using only google.golang.org/protobuf/encoding/protowire
// primitives — no generated .pb.go needed — then assert tm2's amino emission
// is byte-identical.
//
// This is the definitive byte-compatibility test. If tm2's amino output for
// `upstream.Vote{...}` equals what's emitted by a hand-rolled encoder
// following upstream Tendermint v0.34's types.proto, then by construction
// stdlib protobuf decoders (and tmkms) can decode it.
//
// The hand-rolled encoders mirror upstream's gogoproto-generated Marshal
// methods at the wire level, including:
//   - Tag-byte emission via protowire.AppendTag(num, wireType)
//   - Varint encoding for plain int64/int32/uint32 fields
//   - sfixed64 encoding for Height/Round in canonical types
//   - Length-delimited bytes/string/nested-message encoding
//   - google.protobuf.Timestamp = nested message with seconds + nanos
//   - SignedMsgType = varint enum (numeric value matches upstream)
//
// If a future amino change drifts from upstream's wire format, these tests
// fail with byte-level diffs that point straight at the regression.

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
)

// ---- upstream-shaped hand-rolled encoders ---------------------------------

// appendVarintField appends a varint-typed field at fieldNum with value v.
// Skip if v == 0 (proto3 default-value omission, matching amino's behavior).
func appendVarintField(buf []byte, fieldNum protowire.Number, v uint64) []byte {
	if v == 0 {
		return buf
	}
	buf = protowire.AppendTag(buf, fieldNum, protowire.VarintType)
	return protowire.AppendVarint(buf, v)
}

// appendInt64VarintField is for plain int64 fields. Negative values
// sign-extend through uint64 → 10-byte varint.
func appendInt64VarintField(buf []byte, fieldNum protowire.Number, v int64) []byte {
	if v == 0 {
		return buf
	}
	buf = protowire.AppendTag(buf, fieldNum, protowire.VarintType)
	return protowire.AppendVarint(buf, uint64(v))
}

// appendInt32VarintField is for plain int32 fields. Sign-extends int32 →
// int64 → uint64 before varint, matching upstream's wire convention.
func appendInt32VarintField(buf []byte, fieldNum protowire.Number, v int32) []byte {
	if v == 0 {
		return buf
	}
	buf = protowire.AppendTag(buf, fieldNum, protowire.VarintType)
	return protowire.AppendVarint(buf, uint64(int64(v)))
}

// appendSFixed64Field is for sfixed64 fields (Height/Round in canonical types).
// Always 8 bytes. Skip if v == 0 — matches amino's WriteEmpty=false default.
func appendSFixed64Field(buf []byte, fieldNum protowire.Number, v int64) []byte {
	if v == 0 {
		return buf
	}
	buf = protowire.AppendTag(buf, fieldNum, protowire.Fixed64Type)
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint64(tmp, uint64(v))
	return append(buf, tmp...)
}

// appendBytesField appends a length-delimited bytes field. Skip if empty.
func appendBytesField(buf []byte, fieldNum protowire.Number, v []byte) []byte {
	if len(v) == 0 {
		return buf
	}
	buf = protowire.AppendTag(buf, fieldNum, protowire.BytesType)
	return protowire.AppendBytes(buf, v)
}

// appendStringField appends a length-delimited string field. Skip if empty.
func appendStringField(buf []byte, fieldNum protowire.Number, v string) []byte {
	if v == "" {
		return buf
	}
	buf = protowire.AppendTag(buf, fieldNum, protowire.BytesType)
	return protowire.AppendString(buf, v)
}

// appendMessageField appends a length-delimited submessage field by length-
// prefixing the inner bytes. Skip if inner is empty.
func appendMessageField(buf []byte, fieldNum protowire.Number, inner []byte) []byte {
	if len(inner) == 0 {
		return buf
	}
	buf = protowire.AppendTag(buf, fieldNum, protowire.BytesType)
	return protowire.AppendBytes(buf, inner)
}

// encodeTimestamp emits google.protobuf.Timestamp as a nested message with
// fields seconds=int64-varint(1), nanos=int32-varint(2). amino does NOT
// omit Go's zero time.Time{} — that's seconds=-62135596800 (year 0001),
// not seconds=0. Only fully proto3-default (seconds=0 AND nanos=0) is
// omitted, which corresponds to Unix epoch — but amino encodes that with
// no field-2 nanos and no field-1 seconds (both zero), so the inner
// buffer is empty and the enclosing message field gets skipped.
func encodeTimestamp(t time.Time) []byte {
	var inner []byte
	inner = appendInt64VarintField(inner, 1, t.Unix())
	inner = appendInt32VarintField(inner, 2, int32(t.Nanosecond()))
	return inner
}

// upstreamWirePartSetHeader emits PartSetHeader per upstream's types.proto:
//
//	uint32 total = 1; bytes hash = 2;
func upstreamWirePartSetHeader(p upstream.PartSetHeader) []byte {
	var buf []byte
	buf = appendVarintField(buf, 1, uint64(p.Total))
	buf = appendBytesField(buf, 2, p.Hash)
	return buf
}

// upstreamWireBlockID per upstream's types.proto:
//
//	bytes hash = 1; PartSetHeader part_set_header = 2;
func upstreamWireBlockID(b upstream.BlockID) []byte {
	var buf []byte
	buf = appendBytesField(buf, 1, b.Hash)
	buf = appendMessageField(buf, 2, upstreamWirePartSetHeader(b.PartSetHeader))
	return buf
}

// ---- Tests ----------------------------------------------------------------

func TestUpstreamWire_PartSetHeader_ByteIdentical(t *testing.T) {
	t.Parallel()
	cdc := codec(t)

	cases := []upstream.PartSetHeader{
		{Total: 0, Hash: nil},
		{Total: 1, Hash: []byte{0xab}},
		{Total: 1000, Hash: []byte("0123456789abcdef")},
		{Total: 1<<31 - 1, Hash: nil}, // max int32 as uint32 — still in range
	}
	for _, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)
			wantBz := upstreamWirePartSetHeader(c)
			assert.Equal(t, wantBz, amBz, "PartSetHeader %+v: amino bytes diverge from upstream-spec bytes", c)
		})
	}
}

func TestUpstreamWire_BlockID_ByteIdentical(t *testing.T) {
	t.Parallel()
	cdc := codec(t)

	cases := []upstream.BlockID{
		{},
		{Hash: []byte{0x11, 0x22}},
		{
			Hash:          []byte("blockhash-bytes"),
			PartSetHeader: upstream.PartSetHeader{Total: 5, Hash: []byte{0xff}},
		},
	}
	for _, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)
			wantBz := upstreamWireBlockID(c)
			assert.Equal(t, wantBz, amBz)
		})
	}
}

func TestUpstreamWire_Vote_ByteIdentical(t *testing.T) {
	t.Parallel()
	cdc := codec(t)

	mkVote := func(round, idx int32, polNeg bool) upstream.Vote {
		v := upstream.Vote{
			Type:             types.PrecommitType,
			Height:           42,
			Round:            round,
			BlockID:          upstream.BlockID{Hash: []byte{0xaa}},
			ValidatorAddress: make([]byte, 20),
			ValidatorIndex:   idx,
		}
		_ = polNeg
		return v
	}

	cases := []upstream.Vote{
		mkVote(0, 0, false),
		mkVote(3, 7, false),
		mkVote(-1, 0, true),  // negative round → 10-byte plain varint
		mkVote(1, -1, false), // negative index → 10-byte plain varint
		{Type: types.PrevoteType, Height: 1, Round: 1, ValidatorAddress: make([]byte, 20)},
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)
			want := upstreamWireVote(c)
			assert.Equal(t, want, amBz, "case %d: %+v", i, c)
		})
	}
}

func upstreamWireVote(v upstream.Vote) []byte {
	var buf []byte
	buf = appendVarintField(buf, 1, uint64(v.Type))
	buf = appendInt64VarintField(buf, 2, v.Height)
	buf = appendInt32VarintField(buf, 3, v.Round)
	buf = appendMessageField(buf, 4, upstreamWireBlockID(v.BlockID))
	buf = appendMessageField(buf, 5, encodeTimestamp(v.Timestamp))
	buf = appendBytesField(buf, 6, v.ValidatorAddress)
	buf = appendInt32VarintField(buf, 7, v.ValidatorIndex)
	buf = appendBytesField(buf, 8, v.Signature)
	return buf
}

func TestUpstreamWire_Proposal_ByteIdentical(t *testing.T) {
	t.Parallel()
	cdc := codec(t)

	cases := []upstream.Proposal{
		{Type: types.ProposalType, Height: 1, Round: 0, POLRound: -1},
		{Type: types.ProposalType, Height: 100, Round: 2, POLRound: 1, BlockID: upstream.BlockID{Hash: []byte{0x11}}},
		{Type: types.ProposalType, Height: 1, Round: -1, POLRound: -1}, // negatives
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)
			want := upstreamWireProposal(c)
			assert.Equal(t, want, amBz, "case %d: %+v", i, c)
		})
	}
}

func upstreamWireProposal(p upstream.Proposal) []byte {
	var buf []byte
	buf = appendVarintField(buf, 1, uint64(p.Type))
	buf = appendInt64VarintField(buf, 2, p.Height)
	buf = appendInt32VarintField(buf, 3, p.Round)
	buf = appendInt32VarintField(buf, 4, p.POLRound)
	buf = appendMessageField(buf, 5, upstreamWireBlockID(p.BlockID))
	buf = appendMessageField(buf, 6, encodeTimestamp(p.Timestamp))
	buf = appendBytesField(buf, 7, p.Signature)
	return buf
}

// ---- Canonical types (encoded via tm2/pkg/bft/types codec) ----------------

// upstreamWireCanonicalProposal mirrors upstream's canonical.proto:
//
//	SignedMsgType type = 1;
//	sfixed64      height = 2;
//	sfixed64      round = 3;
//	int64         pol_round = 4;
//	CanonicalBlockID block_id = 5;
//	google.protobuf.Timestamp timestamp = 6;
//	string        chain_id = 7;
func upstreamWireCanonicalProposal(p types.CanonicalProposal) []byte {
	var buf []byte
	buf = appendVarintField(buf, 1, uint64(p.Type))
	buf = appendSFixed64Field(buf, 2, p.Height)
	buf = appendSFixed64Field(buf, 3, p.Round)
	buf = appendInt64VarintField(buf, 4, p.POLRound)
	buf = appendMessageField(buf, 5, upstreamWireCanonicalBlockID(p.BlockID))
	buf = appendMessageField(buf, 6, encodeTimestamp(p.Timestamp))
	buf = appendStringField(buf, 7, p.ChainID)
	return buf
}

func upstreamWireCanonicalBlockID(b types.CanonicalBlockID) []byte {
	var buf []byte
	buf = appendBytesField(buf, 1, b.Hash)
	buf = appendMessageField(buf, 2, upstreamWireCanonicalPartSetHeader(b.PartsHeader))
	return buf
}

func upstreamWireCanonicalPartSetHeader(p types.CanonicalPartSetHeader) []byte {
	var buf []byte
	buf = appendVarintField(buf, 1, uint64(p.Total))
	buf = appendBytesField(buf, 2, p.Hash)
	return buf
}

func TestUpstreamWire_CanonicalProposal_ByteIdentical(t *testing.T) {
	t.Parallel()
	cdc := amino.NewCodec()
	cdc.RegisterPackage(types.Package)
	cdc.Seal()

	cases := []types.CanonicalProposal{
		{Type: types.ProposalType, Height: 1, Round: 1, POLRound: -1, ChainID: ""},
		{Type: types.ProposalType, Height: 100, Round: 2, POLRound: 5, ChainID: "test-chain-id"},
		{Type: types.ProposalType, Height: 1, Round: 1, POLRound: -1, BlockID: types.CanonicalBlockID{
			Hash:        []byte{0x01, 0x02},
			PartsHeader: types.CanonicalPartSetHeader{Total: 7, Hash: []byte{0xff, 0xee}},
		}},
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)
			want := upstreamWireCanonicalProposal(c)
			assert.Equal(t, want, amBz, "case %d: %+v", i, c)
		})
	}
}

// CanonicalVote — same shape as Proposal minus pol_round, with field
// numbers adjusted (block_id=4, timestamp=5, chain_id=6).
func upstreamWireCanonicalVote(v types.CanonicalVote) []byte {
	var buf []byte
	buf = appendVarintField(buf, 1, uint64(v.Type))
	buf = appendSFixed64Field(buf, 2, v.Height)
	buf = appendSFixed64Field(buf, 3, v.Round)
	buf = appendMessageField(buf, 4, upstreamWireCanonicalBlockID(v.BlockID))
	buf = appendMessageField(buf, 5, encodeTimestamp(v.Timestamp))
	buf = appendStringField(buf, 6, v.ChainID)
	return buf
}

func TestUpstreamWire_CanonicalVote_ByteIdentical(t *testing.T) {
	t.Parallel()
	cdc := amino.NewCodec()
	cdc.RegisterPackage(types.Package)
	cdc.Seal()

	cases := []types.CanonicalVote{
		{Type: types.PrecommitType, Height: 1, Round: 1, ChainID: ""},
		{Type: types.PrevoteType, Height: 100, Round: 0, ChainID: "test-chain-id"},
		{Type: types.PrecommitType, Height: 42, Round: 3, BlockID: types.CanonicalBlockID{
			Hash:        []byte{0x01, 0x02},
			PartsHeader: types.CanonicalPartSetHeader{Total: 7, Hash: []byte{0xff, 0xee}},
		}, ChainID: "gno.land"},
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)
			want := upstreamWireCanonicalVote(c)
			assert.Equal(t, want, amBz, "case %d: %+v", i, c)
		})
	}
}
