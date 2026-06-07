package upstream_test

// Layer 2: full stdlib-protobuf round-trip via generated .pb.go fixtures.
//
// For each upstream type, encode a value via tm2 amino, then DECODE the
// resulting bytes via google.golang.org/protobuf/proto.Unmarshal into the
// stdlib-generated message struct (in testdata/upstreampb/). Assert
// field-by-field equality. Then re-marshal via stdlib protobuf and assert
// byte-identical to the amino output.
//
// This is the definitive interop test: byte-equal round-trip with stdlib
// protobuf is exactly what tmkms (and any other upstream-protocol consumer)
// will see when decoding tm2's emitted bytes.
//
// The generated .pb.go was produced from upstream.proto in the same
// testdata/ directory; see that file's header for regeneration steps.

import (
	"bytes"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// assertBytesEqual compares two byte slices for content equality, treating
// nil and empty as equal (which assert.Equal does not — it uses
// reflect.DeepEqual which distinguishes []byte(nil) from []byte{}).
func assertBytesEqual(t *testing.T, want, got []byte, msgAndArgs ...any) {
	t.Helper()
	assert.True(t, bytes.Equal(want, got), msgAndArgs...)
}

// ---- PartSetHeader --------------------------------------------------------

func TestStdlibProto_PartSetHeader(t *testing.T) {
	t.Parallel()
	cdc := codec(t)

	cases := []upstream.PartSetHeader{
		{Total: 0, Hash: nil},
		{Total: 1, Hash: []byte{0xab, 0xcd}},
		{Total: 1000, Hash: []byte("0123456789abcdef")},
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)

			var pb upstreampb.PartSetHeader
			require.NoError(t, proto.Unmarshal(amBz, &pb), "case %d", i)

			assert.Equal(t, c.Total, pb.Total)
			assert.Equal(t, c.Hash, pb.Hash)

			// Re-encode via stdlib proto; must be byte-identical to amino.
			pbBz, err := proto.Marshal(&pb)
			require.NoError(t, err)
			assertBytesEqual(t, amBz, pbBz, "amino bytes != stdlib-proto re-encoded bytes")
		})
	}
}

// ---- BlockID --------------------------------------------------------------

func TestStdlibProto_BlockID(t *testing.T) {
	t.Parallel()
	cdc := codec(t)

	cases := []upstream.BlockID{
		{},
		{Hash: []byte{0xaa, 0xbb}},
		{
			Hash:          []byte("blockhash"),
			PartSetHeader: upstream.PartSetHeader{Total: 5, Hash: []byte{0xff}},
		},
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)

			var pb upstreampb.BlockID
			require.NoError(t, proto.Unmarshal(amBz, &pb), "case %d", i)

			assert.Equal(t, c.Hash, pb.Hash)
			if c.PartSetHeader.Total != 0 || len(c.PartSetHeader.Hash) > 0 {
				require.NotNil(t, pb.PartSetHeader, "case %d: PartSetHeader should be set", i)
				assert.Equal(t, c.PartSetHeader.Total, pb.PartSetHeader.Total)
				assert.Equal(t, c.PartSetHeader.Hash, pb.PartSetHeader.Hash)
			}

			pbBz, err := proto.Marshal(&pb)
			require.NoError(t, err)
			assertBytesEqual(t, amBz, pbBz)
		})
	}
}

// ---- Vote ----------------------------------------------------------------

func TestStdlibProto_Vote(t *testing.T) {
	t.Parallel()
	cdc := codec(t)

	cases := []upstream.Vote{
		{
			Type:             types.PrecommitType,
			Height:           42,
			Round:            3,
			BlockID:          upstream.BlockID{Hash: []byte{0xaa}},
			ValidatorAddress: bytesOf20(0x55),
			ValidatorIndex:   7,
			Signature:        []byte{0xde, 0xad},
		},
		// Negative round — verifies plain-varint sign-extension produces
		// 10 bytes that stdlib protobuf decodes as int32(-1).
		{
			Type:             types.PrecommitType,
			Height:           1,
			Round:            -1,
			ValidatorAddress: bytesOf20(0x00),
		},
		// All-zero scalars: stdlib proto3 omits zero values, so amino must
		// also omit them or this test would catch the divergence.
		{ValidatorAddress: bytesOf20(0xff)},
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)

			var pb upstreampb.Vote
			require.NoError(t, proto.Unmarshal(amBz, &pb), "case %d", i)

			assert.Equal(t, upstreampb.SignedMsgType(c.Type), pb.Type)
			assert.Equal(t, c.Height, pb.Height)
			assert.Equal(t, c.Round, pb.Round)
			assert.Equal(t, c.ValidatorAddress, pb.ValidatorAddress)
			assert.Equal(t, c.ValidatorIndex, pb.ValidatorIndex)
			assert.Equal(t, c.Signature, pb.Signature)

			pbBz, err := proto.Marshal(&pb)
			require.NoError(t, err)
			assertBytesEqual(t, amBz, pbBz, "case %d: amino bytes != stdlib-proto re-encoded bytes", i)
		})
	}
}

// ---- Proposal ------------------------------------------------------------

func TestStdlibProto_Proposal(t *testing.T) {
	t.Parallel()
	cdc := codec(t)

	cases := []upstream.Proposal{
		{
			Type:      types.ProposalType,
			Height:    100,
			Round:     2,
			POLRound:  -1,
			BlockID:   upstream.BlockID{Hash: []byte{0x11}},
			Signature: []byte{0xca, 0xfe},
		},
		{Type: types.ProposalType, Height: 1, Round: 0, POLRound: -1},
		{Type: types.ProposalType, Height: 1, Round: -1, POLRound: -1}, // negatives everywhere
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)

			var pb upstreampb.Proposal
			require.NoError(t, proto.Unmarshal(amBz, &pb), "case %d", i)

			assert.Equal(t, upstreampb.SignedMsgType(c.Type), pb.Type)
			assert.Equal(t, c.Height, pb.Height)
			assert.Equal(t, c.Round, pb.Round)
			assert.Equal(t, c.POLRound, pb.PolRound)
			assert.Equal(t, c.Signature, pb.Signature)

			pbBz, err := proto.Marshal(&pb)
			require.NoError(t, err)
			assertBytesEqual(t, amBz, pbBz)
		})
	}
}

// ---- CanonicalProposal --------------------------------------------------

func TestStdlibProto_CanonicalProposal(t *testing.T) {
	t.Parallel()
	cdc := amino.NewCodec()
	cdc.RegisterPackage(types.Package)
	cdc.Seal()

	cases := []types.CanonicalProposal{
		{Type: types.ProposalType, Height: 1, Round: 1, POLRound: -1, ChainID: ""},
		{Type: types.ProposalType, Height: 100, Round: 2, POLRound: 5, ChainID: "test-chain"},
		{
			Type:     types.ProposalType,
			Height:   42,
			Round:    3,
			POLRound: 1,
			BlockID: types.CanonicalBlockID{
				Hash:        []byte{0x01, 0x02},
				PartsHeader: types.CanonicalPartSetHeader{Total: 7, Hash: []byte{0xff, 0xee}},
			},
			ChainID: "gno.land",
		},
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)

			var pb upstreampb.CanonicalProposal
			require.NoError(t, proto.Unmarshal(amBz, &pb), "case %d", i)

			assert.Equal(t, upstreampb.SignedMsgType(c.Type), pb.Type)
			assert.Equal(t, c.Height, pb.Height)
			assert.Equal(t, c.Round, pb.Round)
			assert.Equal(t, c.POLRound, pb.PolRound)
			assert.Equal(t, c.ChainID, pb.ChainId)
			if len(c.BlockID.Hash) > 0 || c.BlockID.PartsHeader.Total != 0 {
				require.NotNil(t, pb.BlockId)
				assert.Equal(t, c.BlockID.Hash, pb.BlockId.Hash)
				if c.BlockID.PartsHeader.Total != 0 {
					require.NotNil(t, pb.BlockId.PartSetHeader)
					assert.Equal(t, c.BlockID.PartsHeader.Total, pb.BlockId.PartSetHeader.Total)
					assert.Equal(t, c.BlockID.PartsHeader.Hash, pb.BlockId.PartSetHeader.Hash)
				}
			}

			pbBz, err := proto.Marshal(&pb)
			require.NoError(t, err)
			assertBytesEqual(t, amBz, pbBz, "case %d: canonical proposal amino bytes != stdlib-proto re-encoded bytes", i)
		})
	}
}

// ---- CanonicalVote ------------------------------------------------------

func TestStdlibProto_CanonicalVote(t *testing.T) {
	t.Parallel()
	cdc := amino.NewCodec()
	cdc.RegisterPackage(types.Package)
	cdc.Seal()

	cases := []types.CanonicalVote{
		{Type: types.PrecommitType, Height: 1, Round: 1, ChainID: ""},
		{Type: types.PrevoteType, Height: 100, Round: 0, ChainID: "test-chain"},
		{
			Type:   types.PrecommitType,
			Height: 42,
			Round:  3,
			BlockID: types.CanonicalBlockID{
				Hash:        []byte{0x01, 0x02},
				PartsHeader: types.CanonicalPartSetHeader{Total: 7, Hash: []byte{0xff, 0xee}},
			},
			ChainID: "gno.land",
		},
	}
	for i, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			amBz, err := cdc.Marshal(&c)
			require.NoError(t, err)

			var pb upstreampb.CanonicalVote
			require.NoError(t, proto.Unmarshal(amBz, &pb), "case %d", i)

			assert.Equal(t, upstreampb.SignedMsgType(c.Type), pb.Type)
			assert.Equal(t, c.Height, pb.Height)
			assert.Equal(t, c.Round, pb.Round)
			assert.Equal(t, c.ChainID, pb.ChainId)

			pbBz, err := proto.Marshal(&pb)
			require.NoError(t, err)
			assertBytesEqual(t, amBz, pbBz)
		})
	}
}

// ---- Timestamp interop --------------------------------------------------

// A non-zero timestamp on a Vote should round-trip as
// google.protobuf.Timestamp{seconds, nanos}.
func TestStdlibProto_Vote_Timestamp(t *testing.T) {
	t.Parallel()
	cdc := codec(t)

	ts := time.Unix(1700000000, 123456789).UTC()
	v := upstream.Vote{
		Type:             types.PrecommitType,
		Height:           1,
		Round:            1,
		Timestamp:        ts,
		ValidatorAddress: bytesOf20(0x00),
	}
	amBz, err := cdc.Marshal(&v)
	require.NoError(t, err)

	var pb upstreampb.Vote
	require.NoError(t, proto.Unmarshal(amBz, &pb))
	require.NotNil(t, pb.Timestamp)
	wantPb := timestamppb.New(ts)
	assert.Equal(t, wantPb.Seconds, pb.Timestamp.Seconds)
	assert.Equal(t, wantPb.Nanos, pb.Timestamp.Nanos)

	pbBz, err := proto.Marshal(&pb)
	require.NoError(t, err)
	assertBytesEqual(t, amBz, pbBz)
}

func bytesOf20(b byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = b
	}
	return out
}
