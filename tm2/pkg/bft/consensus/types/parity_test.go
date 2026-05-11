package cstypes_test

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
)

func bitArrayWith(size int, mask uint64) *bitarray.BitArray {
	ba := bitarray.NewBitArray(size)
	for i := 0; i < size; i++ {
		if mask&(1<<uint(i)) != 0 {
			ba.SetIndex(i, true)
		}
	}
	return ba
}

func TestCodecParity_ConsensusTypes(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(btypes.Package)
	cdc.RegisterPackage(ctypes.Package)
	cdc.Seal()

	stamp := time.Date(2026, time.April, 24, 12, 0, 0, 0, time.UTC)
	addr := crypto.AddressFromPreimage([]byte("proposer"))

	cases := []struct {
		name string
		v    any
	}{
		{"HRS", &ctypes.HRS{Height: 100, Round: 2, Step: ctypes.RoundStepPrecommit}},
		{"EventNewRound", &ctypes.EventNewRound{
			HRS:           ctypes.HRS{Height: 7, Round: 0, Step: ctypes.RoundStepPropose},
			Proposer:      btypes.Validator{Address: addr, VotingPower: 1},
			ProposerIndex: 0,
		}},
		{"EventNewRoundStep", &ctypes.EventNewRoundStep{
			HRS: ctypes.HRS{Height: 7, Round: 1, Step: ctypes.RoundStepPrevote},
		}},
		{"EventCompleteProposal", &ctypes.EventCompleteProposal{
			HRS:     ctypes.HRS{Height: 9, Round: 0, Step: ctypes.RoundStepPropose},
			BlockID: btypes.BlockID{Hash: []byte{0x01}},
		}},
		{"EventTimeoutPropose", &ctypes.EventTimeoutPropose{
			HRS: ctypes.HRS{Height: 1, Round: 0, Step: ctypes.RoundStepPropose},
		}},
		{"EventTimeoutWait", &ctypes.EventTimeoutWait{
			HRS: ctypes.HRS{Height: 1, Round: 0, Step: ctypes.RoundStepPrevoteWait},
		}},

		// EventNewValidBlock: fired when a valid block is seen (part of the
		// gossip that precedes a commit).
		{"EventNewValidBlock", &ctypes.EventNewValidBlock{
			HRS:              ctypes.HRS{Height: 50, Round: 0, Step: ctypes.RoundStepCommit},
			BlockPartsHeader: btypes.PartSetHeader{Total: 4, Hash: []byte{0xaa, 0xbb}},
			BlockParts:       bitArrayWith(4, 0b1111),
			IsCommit:         true,
		}},

		// RoundStateSimple: compressed round state for RPC.
		{"RoundStateSimple", &ctypes.RoundStateSimple{
			HeightRoundStep:   "100/0/propose",
			StartTime:         stamp,
			ProposalBlockHash: []byte{0x01, 0x02, 0x03},
			// Leave LockedBlockHash / ValidBlockHash nil; exercises the
			// nil vs non-nil byte-slice distinction in one struct.
		}},

		// PeerRoundState: the full peer round-state snapshot.
		{"PeerRoundState", &ctypes.PeerRoundState{
			Height:                   100,
			Round:                    2,
			Step:                     ctypes.RoundStepPrevote,
			StartTime:                stamp,
			Proposal:                 true,
			ProposalBlockPartsHeader: btypes.PartSetHeader{Total: 8, Hash: []byte{0xde, 0xad}},
			ProposalBlockParts:       bitArrayWith(8, 0b11110000),
			ProposalPOLRound:         -1,
			ProposalPOL:              nil, // -1 round → nil POL
			Prevotes:                 bitArrayWith(4, 0b1010),
			Precommits:               bitArrayWith(4, 0),
			LastCommitRound:          1,
			LastCommit:               bitArrayWith(4, 0b1111),
			CatchupCommitRound:       -1,
			CatchupCommit:            nil,
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
