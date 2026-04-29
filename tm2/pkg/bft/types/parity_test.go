package types

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
)

// TestCodecParity_BFTTypes asserts that every hand-crafted consensus value
// round-trips byte-identically through both the reflect codec and the
// genproto2 fast path, that SizeBinary2 matches the encoded length, and
// that both codec paths agree on the decoded value.
//
// The array below hand-picks edge cases representative of the codec
// surfaces the recent fixes touched: nil_elements on Precommits,
// AminoMarshaler zero-repr-Address emission, fixed64 signing fields, and
// interface-carrying fields. Add new cases by appending entries.
func TestCodecParity_BFTTypes(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(abci.Package)
	cdc.RegisterPackage(ed25519.Package)
	cdc.RegisterPackage(Package)
	cdc.Seal()

	for i, c := range parityCasesBFT(t) {
		c := c
		name := fmt.Sprintf("%d/%s", i, c.name)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}

func parityCasesBFT(t *testing.T) []struct {
	name string
	v    any
} {
	t.Helper()

	stamp, err := time.Parse(TimeFormat, "2026-04-24T12:34:56.789Z")
	if err != nil {
		t.Fatalf("parse stamp: %v", err)
	}
	addr1 := crypto.AddressFromPreimage([]byte("validator-1"))
	addr2 := crypto.AddressFromPreimage([]byte("validator-2"))
	proposerAddr := crypto.AddressFromPreimage([]byte("proposer"))

	// Vote signed by validator 1. Non-zero Signature + Address to surface
	// any Signature-byte mangling at decode.
	sigA := &CommitSig{
		Type:             PrecommitType,
		Height:           100,
		Round:            2,
		BlockID:          BlockID{Hash: []byte{0x01, 0x02, 0x03}, PartsHeader: PartSetHeader{Total: 4, Hash: []byte{0xaa}}},
		Timestamp:        stamp,
		ValidatorAddress: addr1,
		ValidatorIndex:   0,
		Signature:        []byte{0xde, 0xad, 0xbe, 0xef},
	}
	sigB := &CommitSig{
		Type:             PrecommitType,
		Height:           100,
		Round:            2,
		BlockID:          sigA.BlockID,
		Timestamp:        stamp,
		ValidatorAddress: addr2,
		ValidatorIndex:   2,
		Signature:        []byte{0xca, 0xfe, 0xba, 0xbe},
	}
	// nil_elements: [sigA, nil, sigB]. The nil entry exercises the
	// consensus-wedging bug surface. NOTE: construct directly rather than
	// via helpers that call commit.Hash() — that would populate the
	// unexported memoization fields, breaking strict reflect.DeepEqual.
	commit := &Commit{
		BlockID:    sigA.BlockID,
		Precommits: []*CommitSig{sigA, nil, sigB},
	}
	// Separate commit for the Block test case so the Commit test's commit
	// stays untouched by Block construction.
	commitForBlock := &Commit{
		BlockID:    sigA.BlockID,
		Precommits: []*CommitSig{sigA, nil, sigB},
	}

	// Construct Block directly (bypassing MakeBlock / fillHeader, which
	// would call LastCommit.Hash() and populate memoized cache fields).
	block := &Block{
		Header: Header{
			Height:          101,
			NumTxs:          0,
			TotalTxs:        0,
			ProposerAddress: proposerAddr,
		},
		LastCommit: commitForBlock,
	}

	return []struct {
		name string
		v    any
	}{
		// Vote/Proposal scalar edge values. fixed64 on Height/Round means
		// math.MinInt64 and math.MaxInt64 specifically stress the encoding.
		{"Vote/zero", &Vote{}},
		{"Vote/minmax", &Vote{
			Type:      PrecommitType,
			Height:    math.MaxInt64,
			Round:     math.MaxInt32, // Round is int, capped by varint
			Timestamp: stamp,
		}},
		{"Vote/negative-height", &Vote{
			Type:      PrecommitType,
			Height:    -1, // ValidateBasic would reject, but wire-format must still round-trip
			Round:     0,
			Timestamp: stamp,
		}},
		// Varint-boundary cases for Round (plain int, not fixed64).
		// 127 vs 128 crosses the 1-byte to 2-byte varint boundary;
		// 16383 vs 16384 crosses 2-byte to 3-byte. Height is fixed64
		// (always 8 bytes) so these cases target Round specifically.
		{"Vote/round-127", &Vote{
			Type: PrecommitType, Height: 1, Round: 127, Timestamp: stamp,
		}},
		{"Vote/round-128", &Vote{
			Type: PrecommitType, Height: 1, Round: 128, Timestamp: stamp,
		}},
		{"Vote/round-16384", &Vote{
			Type: PrecommitType, Height: 1, Round: 16384, Timestamp: stamp,
		}},
		{"Proposal/polround-neg", &Proposal{
			Type:      ProposalType,
			Height:    42,
			Round:     1,
			POLRound:  -1,
			BlockID:   sigA.BlockID,
			Timestamp: stamp,
		}},

		// CommitSig with zero ValidatorAddress — the AminoMarshaler
		// repr-zeroness fix's primary regression surface. Zero [20]byte
		// produces bech32 "g1qqq...luuxe" which must be emitted, not
		// omitted.
		{"CommitSig/zero-address", &CommitSig{
			Type:      PrecommitType,
			Height:    1,
			Round:     0,
			Timestamp: stamp,
			// ValidatorAddress left as zero
			ValidatorIndex: 7,
			Signature:      []byte{0x01, 0x02},
		}},

		// CommitSig with everything populated.
		{"CommitSig/full", sigA},

		// Commit with nil entry — the nil_elements round-trip.
		{"Commit/with-nil-precommit", commit},

		// Block containing the nil-precommit commit plus non-zero
		// ProposerAddress.
		{"Block/with-nil-precommit", block},

		// BlockID variants.
		{"BlockID/empty", &BlockID{}},
		{"BlockID/full", &sigA.BlockID},

		// Header alone with non-trivial fields.
		{"Header/populated", &Header{
			Height:          101,
			NumTxs:          3,
			TotalTxs:        42,
			ProposerAddress: proposerAddr,
			Time:            stamp,
			LastCommitHash:  []byte{0x0a, 0x0b},
		}},

		// Validator and ValidatorSet.
		{"Validator/simple", &Validator{
			Address:     addr1,
			PubKey:      nil, // Interface; nil round-trips as nil
			VotingPower: 100,
		}},

		// Data — transactions in a block. The unexported Data.hash cache
		// is only populated by Data.Hash(); constructing directly keeps
		// strict DeepEqual working.
		{"Data/empty", &Data{}},
		// Txs elements are non-empty — empty []byte roundtrips to nil under
		// amino (documented; see parity helper docstring).
		{"Data/populated", &Data{
			Txs: []Tx{
				Tx([]byte{0xde, 0xad, 0xbe, 0xef}),
				Tx([]byte("hello")),
				Tx([]byte{0x01}),
			},
		}},

		// Part: block part with merkle proof. Proof.Aunts exercises
		// [][]byte nested-byte-slice encoding under a non-empty context.
		{"Part", &Part{
			Index: 2,
			Bytes: []byte{0x01, 0x02, 0x03, 0x04},
			Proof: merkle.SimpleProof{
				Total:    4,
				Index:    2,
				LeafHash: []byte{0xaa, 0xbb, 0xcc},
				Aunts: [][]byte{
					{0x10, 0x20},
					{0x30},
				},
			},
		}},

		// PartSetHeader: small fixed struct. Include empty and populated.
		{"PartSetHeader/empty", &PartSetHeader{}},
		{"PartSetHeader/populated", &PartSetHeader{
			Total: 7,
			Hash:  []byte{0x12, 0x34, 0x56, 0x78},
		}},

		// ValidatorSet: amino-serialized form (the exported Validators +
		// Proposer fields; the unexported totalVotingPower cache is only
		// set via computations we don't invoke).
		{"ValidatorSet", func() *ValidatorSet {
			v1 := &Validator{Address: addr1, VotingPower: 100}
			v2 := &Validator{Address: addr2, VotingPower: 200}
			return &ValidatorSet{
				Validators: []*Validator{v1, v2},
				Proposer:   v1,
			}
		}()},

		// Event types fired by the consensus state machine.
		{"EventNewBlockHeader", &EventNewBlockHeader{
			Header: Header{Height: 100, NumTxs: 3, ProposerAddress: proposerAddr, Time: stamp},
		}},
		{"EventVote", &EventVote{
			Vote: &Vote{
				Type:             PrecommitType,
				Height:           100,
				Round:            0,
				Timestamp:        stamp,
				ValidatorAddress: addr1,
				Signature:        []byte{0x01, 0x02},
			},
		}},
		{"EventString", func() *EventString { e := EventString("new_block"); return &e }()},
		{"EventValidatorSetUpdates", &EventValidatorSetUpdates{
			ValidatorUpdates: []abci.ValidatorUpdate{
				{Address: addr1, Power: 150},
				{Address: addr2, Power: 0}, // zero voting power → removal
			},
		}},
		{"EventTx", &EventTx{
			Result: TxResult{
				Height: 50,
				Index:  1,
				Tx:     Tx([]byte{0xde, 0xad}),
			},
		}},

		// DuplicateVoteEvidence — slashable evidence for equivocation.
		{"DuplicateVoteEvidence", &DuplicateVoteEvidence{
			PubKey: ed25519.PubKeyEd25519{0x01, 0x02, 0x03},
			VoteA: &Vote{
				Type:             PrecommitType,
				Height:           100,
				Round:            0,
				BlockID:          BlockID{Hash: []byte{0xaa}},
				Timestamp:        stamp,
				ValidatorAddress: addr1,
				Signature:        []byte{0xde, 0xad},
			},
			VoteB: &Vote{
				Type:             PrecommitType,
				Height:           100,
				Round:            0,
				BlockID:          BlockID{Hash: []byte{0xbb}},
				Timestamp:        stamp,
				ValidatorAddress: addr1,
				Signature:        []byte{0xbe, 0xef},
			},
		}},
	}
}
