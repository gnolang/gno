package consensus

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/stretchr/testify/require"
)

// fullBitArray is a bit array of n bits whose every element is all ones, the
// longest encoding of n bits since amino writes each element as a varint.
func fullBitArray(n int) *bitarray.BitArray {
	return bitarray.NewBitArray(n).Not()
}

// largestMessages returns, for each channel, the largest instance of every
// message type a peer sends on it: each integer at its widest valid value, each
// hash and signature at its longest, each bit array at the count its
// validation admits, all bits set.
func largestMessages(t *testing.T) map[byte][]ConsensusMessage {
	t.Helper()

	hash := bytes.Repeat([]byte{0xff}, tmhash.Size)
	blockID := types.BlockID{
		Hash:        hash,
		PartsHeader: types.PartSetHeader{Total: types.MaxBlockPartsCount, Hash: hash},
	}
	timestamp := time.Date(9999, 12, 31, 23, 59, 59, 999_999_999, time.UTC)
	signature := bytes.Repeat([]byte{0xff}, types.MaxSignatureSize)
	var address types.Address
	for i := range address {
		address[i] = 0xff
	}

	// The deepest proof a part of the largest block carries.
	_, proofs := merkle.SimpleProofsFromByteSlices(make([][]byte, types.MaxBlockPartsCount))
	proof := proofs[0]
	for _, p := range proofs {
		if len(p.Aunts) > len(proof.Aunts) {
			proof = p
		}
	}
	for i := range proof.Aunts {
		proof.Aunts[i] = hash
	}
	proof.LeafHash = hash

	return map[byte][]ConsensusMessage{
		StateChannel: {
			&NewRoundStepMessage{
				Height:                math.MaxInt64,
				Round:                 math.MaxInt,
				Step:                  cstypes.RoundStepCommit,
				SecondsSinceStartTime: math.MinInt,
				LastCommitRound:       math.MaxInt,
			},
			&NewValidBlockMessage{
				Height:           math.MaxInt64,
				Round:            math.MaxInt,
				BlockPartsHeader: types.PartSetHeader{Total: types.MaxBlockPartsCount, Hash: hash},
				BlockParts:       fullBitArray(types.MaxBlockPartsCount),
				IsCommit:         true,
			},
			&HasVoteMessage{
				Height: math.MaxInt64,
				Round:  math.MaxInt,
				Type:   types.PrecommitType,
				Index:  math.MaxInt,
			},
			&VoteSetMaj23Message{
				Height:  math.MaxInt64,
				Round:   math.MaxInt,
				Type:    types.PrecommitType,
				BlockID: blockID,
			},
		},
		DataChannel: {
			&ProposalMessage{Proposal: &types.Proposal{
				Type:      types.ProposalType,
				Height:    math.MaxInt64,
				Round:     math.MaxInt,
				POLRound:  math.MaxInt - 1,
				BlockID:   blockID,
				Timestamp: timestamp,
				Signature: signature,
			}},
			&ProposalPOLMessage{
				Height:           math.MaxInt64,
				ProposalPOLRound: math.MaxInt,
				ProposalPOL:      fullBitArray(types.MaxVotesCount),
			},
			&BlockPartMessage{
				Height: math.MaxInt64,
				Round:  math.MaxInt,
				Part: &types.Part{
					Index: math.MaxInt,
					Bytes: bytes.Repeat([]byte{0xff}, types.BlockPartSizeBytes),
					Proof: *proof,
				},
			},
		},
		VoteChannel: {
			&VoteMessage{Vote: &types.Vote{
				Type:             types.PrecommitType,
				Height:           math.MaxInt64,
				Round:            math.MaxInt,
				BlockID:          blockID,
				Timestamp:        timestamp,
				ValidatorAddress: address,
				ValidatorIndex:   math.MaxInt,
				Signature:        signature,
			}},
		},
		VoteSetBitsChannel: {
			&VoteSetBitsMessage{
				Height:  math.MaxInt64,
				Round:   math.MaxInt,
				Type:    types.PrecommitType,
				BlockID: blockID,
				Votes:   fullBitArray(types.MaxVotesCount),
			},
		},
	}
}

// TestChannelCapacityFitsLargestMessage sends the largest valid message of
// each channel, and a message of exactly the channel's capacity, through a
// connection built with the reactor's channels, and one byte more, which the
// connection refuses. No channel preallocates more than it accepts.
func TestChannelCapacityFitsLargestMessage(t *testing.T) {
	t.Parallel()

	capacity := map[byte]int{
		StateChannel:       1284,
		DataChannel:        66560,
		VoteChannel:        1024,
		VoteSetBitsChannel: 2594,
	}
	largest := largestMessages(t)
	descs := (&ConsensusReactor{}).GetChannels()
	require.Len(t, descs, len(capacity))

	for _, desc := range descs {
		want, ok := capacity[desc.ID]
		require.True(t, ok, "channel %#x", desc.ID)
		require.LessOrEqual(t, desc.FillDefaults().RecvBufferCapacity, want,
			"channel %#x preallocates more than it accepts", desc.ID)

		size := 0
		for _, msg := range largest[desc.ID] {
			bz := amino.MustMarshalAny(msg)
			decoded, err := decodeMsg(bz)
			require.NoError(t, err)
			require.NoError(t, decoded.ValidateBasic(), "%T", msg)
			t.Logf("channel %#x: %T %d bytes, %d under capacity", desc.ID, msg, len(bz), want-len(bz))
			size = max(size, len(bz))
		}

		require.NoError(t, p2pTesting.Deliver(t, *desc, size), "channel %#x, largest message", desc.ID)
		require.NoError(t, p2pTesting.Deliver(t, *desc, want), "channel %#x, at capacity", desc.ID)
		require.ErrorContains(t, p2pTesting.Deliver(t, *desc, want+1), "exceeds available capacity",
			"channel %#x, one byte past capacity", desc.ID)
	}
}
