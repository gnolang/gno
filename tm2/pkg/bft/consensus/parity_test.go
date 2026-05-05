package consensus

// Internal test package: most reactor message types are unexported.

import (
	"fmt"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	btypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
)

// mustBitArray builds a non-nil BitArray of the given size with the bits
// from `mask` set (LSB = index 0).
func mustBitArray(size int, mask uint64) *bitarray.BitArray {
	ba := bitarray.NewBitArray(size)
	for i := 0; i < size; i++ {
		if mask&(1<<uint(i)) != 0 {
			ba.SetIndex(i, true)
		}
	}
	return ba
}

func TestCodecParity_Consensus(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(btypes.Package)
	cdc.RegisterPackage(ctypes.Package)
	cdc.RegisterPackage(Package)
	cdc.Seal()

	stamp := time.Date(2026, time.April, 24, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		v    any
	}{
		{"NewRoundStepMessage", &NewRoundStepMessage{
			Height: 100, Round: 0, Step: ctypes.RoundStepPropose,
			SecondsSinceStartTime: 3, LastCommitRound: -1,
		}},
		{"HasVoteMessage", &HasVoteMessage{
			Height: 100, Round: 0, Type: btypes.PrecommitType, Index: 2,
		}},
		{"VoteSetMaj23Message", &VoteSetMaj23Message{
			Height: 100, Round: 0, Type: btypes.PrecommitType,
			BlockID: btypes.BlockID{Hash: []byte{0x01, 0x02}},
		}},
		{"timeoutInfo", &timeoutInfo{
			Duration: 2 * time.Second, Height: 50, Round: 1,
			Step: ctypes.RoundStepPrevote,
		}},
		{"newRoundStepInfo", &newRoundStepInfo{
			HRS: ctypes.HRS{Height: 10, Round: 0, Step: ctypes.RoundStepPrevote},
		}},
		// msgInfo.Msg is a ConsensusMessage interface; leave nil to keep
		// this test self-contained (other cases exercise concrete message
		// types directly).
		{"msgInfo/nil-msg", &msgInfo{PeerID: p2pTypes.ID("peer-1")}},

		// ProposalMessage: consensus-critical — the gossip envelope for
		// block proposals. Wraps a *Proposal (zero POLRound here; varied
		// in bft/types parity fixtures).
		{"ProposalMessage", &ProposalMessage{
			Proposal: &btypes.Proposal{
				Type:      btypes.ProposalType,
				Height:    100,
				Round:     2,
				POLRound:  -1,
				BlockID:   btypes.BlockID{Hash: []byte{0xaa, 0xbb}},
				Timestamp: stamp,
				Signature: []byte{0x01, 0x02, 0x03, 0x04},
			},
		}},

		// VoteMessage: consensus-critical — the gossip envelope for votes.
		{"VoteMessage", &VoteMessage{
			Vote: &btypes.Vote{
				Type:             btypes.PrecommitType,
				Height:           100,
				Round:            0,
				BlockID:          btypes.BlockID{Hash: []byte{0xcc, 0xdd}},
				Timestamp:        stamp,
				ValidatorAddress: crypto.AddressFromPreimage([]byte("v1")),
				ValidatorIndex:   3,
				Signature:        []byte{0xde, 0xad, 0xbe, 0xef},
			},
		}},

		// BlockPartMessage: gossip envelope for a part of a proposed block.
		{"BlockPartMessage", &BlockPartMessage{
			Height: 100,
			Round:  0,
			Part: &btypes.Part{
				Index: 3,
				Bytes: []byte{0x01, 0x02, 0x03},
				Proof: merkle.SimpleProof{
					Total:    8,
					Index:    3,
					LeafHash: []byte{0xaa},
					Aunts:    [][]byte{{0xbb}, {0xcc}},
				},
			},
		}},

		// NewValidBlockMessage: broadcast when a block is seen as valid.
		{"NewValidBlockMessage", &NewValidBlockMessage{
			Height:           100,
			Round:            0,
			BlockPartsHeader: btypes.PartSetHeader{Total: 4, Hash: []byte{0xde, 0xad}},
			BlockParts:       mustBitArray(4, 0b1010),
			IsCommit:         true,
		}},

		// ProposalPOLMessage: re-propose a previous proposal with its POL bits.
		{"ProposalPOLMessage", &ProposalPOLMessage{
			Height:           100,
			ProposalPOLRound: 0,
			ProposalPOL:      mustBitArray(8, 0b10101010),
		}},

		// VoteSetBitsMessage: bit-array of votes seen for a BlockID.
		{"VoteSetBitsMessage", &VoteSetBitsMessage{
			Height:  100,
			Round:   0,
			Type:    btypes.PrecommitType,
			BlockID: btypes.BlockID{Hash: []byte{0xef}},
			Votes:   mustBitArray(5, 0b11011),
		}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
