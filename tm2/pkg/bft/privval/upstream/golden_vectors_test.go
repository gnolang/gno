package upstream_test

// Frozen-byte golden vectors for upstream-compat types.
//
// The other test layers (types_test.go schema walks, upstreamwire_test.go
// hand-rolled encoders, stdlibproto_test.go protobuf-go round-trips) all
// compute the "expected" bytes at runtime. They catch encoder/decoder
// disagreement, but not silent drift in BOTH sides of the comparison.
// These hex-frozen tests are the last line: if tm2's wire format drifts,
// they fail against the frozen bytes even if the other layers stay
// self-consistent.
//
// !!! IMPORTANT: vectors here come from TWO DIFFERENT ORACLES !!!
//
//   - upstream/* vectors are oracle-derived from upstreampb's
//     protoc-generated Marshal (`proto.Marshal(...)`). These are a
//     genuine third-party byte-equivalence check against upstream
//     Tendermint v0.34's wire format.
//
//   - canonical/* vectors are SELF-derived from tm2 amino itself, at
//     the time PR #5625 was verified against a real tmkms 0.15.0
//     binary on the wire. They are NOT recomputed against upstream
//     protoc; they pin tm2 amino against its own past output. The
//     upstream-equivalence claim for canonical types rests on (a) the
//     schema/wire layers in the sibling tests, and (b) that one-time
//     tmkms 0.15.0 wire interop check at #5625.
//
// Re-capturing: emit bytes for upstream/* types via upstreampb's
// protoc-generated Marshal (`proto.Marshal(...)`); emit bytes for the
// canonical/* types via tm2 amino with the bft/types package registered.
// Both encoders are deterministic for the inputs below.
//
// One nuance for the upstream/* vectors: upstreampb wraps Timestamp as a
// nullable `*timestamppb.Timestamp`, so a nil pointer omits the field
// outright. tm2 amino uses a value `time.Time` and only omits the field
// when seconds=0 AND nanos=0 (Unix epoch). To keep both encoders byte-
// identical on the "no timestamp" path, vectors below use
// `time.Unix(0, 0).UTC()` rather than Go's zero `time.Time{}` (which
// would encode as Year-0001, seconds=-62135596800).

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upstreamRefTimestamp is the reference timestamp used for golden vectors:
// 2023-11-14 22:13:20 UTC = Unix 1_700_000_000. Frozen to keep the bytes
// reproducible.
var upstreamRefTimestamp = time.Unix(1_700_000_000, 0).UTC()

// upstreamGolden runs the assertion: encode v via tm2 amino, compare the
// resulting bytes to the hex-frozen golden vector.
func upstreamGolden(t *testing.T, name, goldenHex string, v any) {
	t.Helper()
	want, err := hex.DecodeString(goldenHex)
	require.NoErrorf(t, err, "%s: invalid golden hex", name)

	cdc := codec(t)
	got, err := cdc.Marshal(v)
	require.NoErrorf(t, err, "%s: amino.Marshal failed", name)

	// bytes.Equal treats nil and empty as equal; assert.Equal does not.
	assert.Truef(t, bytes.Equal(want, got),
		"%s: amino bytes drift from upstream golden\n  want %x\n  got  %x",
		name, want, got)
}

// canonicalGolden is the same as upstreamGolden but registers bft/types,
// not the upstream package — used for CanonicalProposal / CanonicalVote.
func canonicalGolden(t *testing.T, name, goldenHex string, v any) {
	t.Helper()
	want, err := hex.DecodeString(goldenHex)
	require.NoErrorf(t, err, "%s: invalid golden hex", name)

	cdc := amino.NewCodec()
	cdc.RegisterPackage(types.Package)
	cdc.Seal()
	got, err := cdc.Marshal(v)
	require.NoErrorf(t, err, "%s: amino.Marshal failed", name)

	// bytes.Equal treats nil and empty as equal; assert.Equal does not.
	assert.Truef(t, bytes.Equal(want, got),
		"%s: amino bytes drift from upstream golden\n  want %x\n  got  %x",
		name, want, got)
}

// ---- upstream/* (protoc-emitted golden) -----------------------------------

func TestGolden_Upstream_PartSetHeader(t *testing.T) {
	t.Parallel()

	upstreamGolden(t, "empty", "", &upstream.PartSetHeader{})
	upstreamGolden(t, "basic", "08071202abcd",
		&upstream.PartSetHeader{Total: 7, Hash: []byte{0xab, 0xcd}})
	// Total at int32-max boundary, hash = 32 zero bytes.
	upstreamGolden(t, "max_total",
		"08ffffffff0712200000000000000000000000000000000000000000000000000000000000000000",
		&upstream.PartSetHeader{Total: 0x7fffffff, Hash: make([]byte, 32)})
}

func TestGolden_Upstream_BlockID(t *testing.T) {
	t.Parallel()

	upstreamGolden(t, "empty", "", &upstream.BlockID{})
	upstreamGolden(t, "basic", "0a09626c6f636b68617368120508051201ff",
		&upstream.BlockID{
			Hash:          []byte("blockhash"),
			PartSetHeader: upstream.PartSetHeader{Total: 5, Hash: []byte{0xff}},
		})
}

func TestGolden_Upstream_Vote_Precommit(t *testing.T) {
	t.Parallel()

	v := &upstream.Vote{
		Type:             types.PrecommitType,
		Height:           42,
		Round:            3,
		BlockID:          upstream.BlockID{Hash: []byte{0xaa}},
		Timestamp:        upstreamRefTimestamp,
		ValidatorAddress: make([]byte, 20),
		ValidatorIndex:   5,
		Signature:        []byte{0xde, 0xad, 0xbe, 0xef},
	}
	upstreamGolden(t, "precommit_full",
		"0802102a180322030a01aa2a060880e2cfaa063214000000000000000000000000000000000000000038054204deadbeef",
		v)
}

// Plain int32 -1 sign-extends to a 10-byte varint. Catches any regression
// to zigzag (sint32) encoding, which would shrink -1 to a single 0x01 byte.
func TestGolden_Upstream_Vote_NegativeRound(t *testing.T) {
	t.Parallel()

	upstreamGolden(t, "negative_round",
		"0802100118ffffffffffffffffff0132140000000000000000000000000000000000000000",
		&upstream.Vote{
			Type:             types.PrecommitType,
			Height:           1,
			Round:            -1,
			Timestamp:        time.Unix(0, 0).UTC(),
			ValidatorAddress: make([]byte, 20),
		})
}

func TestGolden_Upstream_Proposal_POLMinusOne(t *testing.T) {
	t.Parallel()

	upstreamGolden(t, "pol_minus_one",
		"08201064180220ffffffffffffffffff012a030a011132060880e2cfaa063a02cafe",
		&upstream.Proposal{
			Type:      types.ProposalType,
			Height:    100,
			Round:     2,
			POLRound:  -1,
			BlockID:   upstream.BlockID{Hash: []byte{0x11}},
			Timestamp: upstreamRefTimestamp,
			Signature: []byte{0xca, 0xfe},
		})
}

// Proposal with the minimum non-default field set: type + height + Unix-epoch
// timestamp. POLRound=0 is the proto3 default and must be omitted (no field-4
// tag in the bytes). Timestamp set to Unix epoch — both upstreampb's nil
// *Timestamp and tm2 amino's time.Unix(0,0) produce an empty inner buffer,
// so the enclosing field-6 tag is omitted in both encoders.
func TestGolden_Upstream_Proposal_Minimal(t *testing.T) {
	t.Parallel()

	upstreamGolden(t, "minimal", "08201001",
		&upstream.Proposal{
			Type:      types.ProposalType,
			Height:    1,
			Round:     0,
			POLRound:  0,
			Timestamp: time.Unix(0, 0).UTC(),
		})
}

// ---- canonical/* (tm2-amino-emitted golden) -------------------------------
//
// These are what tmkms (and any v0.34-compat external signer) actually
// signs over. Height/Round are sfixed64 here — wire-type 1, always 8 bytes,
// not varint — which is the key distinction from the upstream/* Vote/Proposal
// types above.

func TestGolden_Canonical_Proposal_Basic(t *testing.T) {
	t.Parallel()

	canonicalGolden(t, "basic",
		"082011010000000000000019010000000000000020ffffffffffffffffff01320b088092b8c398feffffff013a0d746573742d636861696e2d6964",
		&types.CanonicalProposal{
			Type:     types.ProposalType,
			Height:   1,
			Round:    1,
			POLRound: -1,
			ChainID:  "test-chain-id",
		})
}

func TestGolden_Canonical_Proposal_POLSet(t *testing.T) {
	t.Parallel()

	canonicalGolden(t, "pol_set",
		"082011640000000000000019020000000000000020052a0c0a020102120608071202ffee32060880e2cfaa063a08676e6f2e6c616e64",
		&types.CanonicalProposal{
			Type:     types.ProposalType,
			Height:   100,
			Round:    2,
			POLRound: 5,
			ChainID:  "gno.land",
			BlockID: types.CanonicalBlockID{
				Hash:        []byte{0x01, 0x02},
				PartsHeader: types.CanonicalPartSetHeader{Total: 7, Hash: []byte{0xff, 0xee}},
			},
			Timestamp: upstreamRefTimestamp,
		})
}

func TestGolden_Canonical_Vote_Prevote(t *testing.T) {
	t.Parallel()

	canonicalGolden(t, "prevote",
		"08011164000000000000002a0b088092b8c398feffffff013208676e6f2e6c616e64",
		&types.CanonicalVote{
			Type:    types.PrevoteType,
			Height:  100,
			Round:   0,
			ChainID: "gno.land",
		})
}

func TestGolden_Canonical_Vote_Precommit(t *testing.T) {
	t.Parallel()

	canonicalGolden(t, "precommit",
		"0802112a00000000000000190300000000000000220c0a020102120608071202ffee2a060880e2cfaa063208676e6f2e6c616e64",
		&types.CanonicalVote{
			Type:   types.PrecommitType,
			Height: 42,
			Round:  3,
			BlockID: types.CanonicalBlockID{
				Hash:        []byte{0x01, 0x02},
				PartsHeader: types.CanonicalPartSetHeader{Total: 7, Hash: []byte{0xff, 0xee}},
			},
			Timestamp: upstreamRefTimestamp,
			ChainID:   "gno.land",
		})
}
