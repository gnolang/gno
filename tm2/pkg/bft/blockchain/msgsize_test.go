package blockchain

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaxMsgSizeFitsLargestBlockResponse pins the relationship maxMsgSize
// depends on: a bcBlockResponseMessage carrying the largest block the chain
// could ever have committed must fit within maxMsgSize.
//
// A block is only committable if peers could decode its proposal, and
// ConsensusState.addProposalBlockPart decodes with maxSize = MaxDataBytes, so
// the whole serialized block (header and LastCommit included) is bounded by
// MaxDataBytes — itself capped at MaxBlockDataBytesLimit by
// ValidateConsensusParams. maxBlockMsgOverhead therefore only has to cover the
// message envelope amino wraps around that block.
func TestMaxMsgSizeFitsLargestBlockResponse(t *testing.T) {
	t.Parallel()

	// Build a block whose serialized size sits just under the ceiling, which is
	// the largest one that could have passed a proposal decode. The tx count is
	// derived from a single measurement rather than probed, so this stays to a
	// couple of multi-MB marshals -- it runs alongside timing-sensitive suites.
	const txSize = 1 << 16

	blockSize := func(txs []types.Tx) int64 {
		sized, err := amino.MarshalSized(types.MakeBlock(1, txs, &types.Commit{}))
		require.NoError(t, err)

		return int64(len(sized))
	}

	// Allow 8 bytes per tx for amino's field key and length prefix, which
	// over-estimates slightly so the estimate lands under the ceiling.
	count := (types.MaxBlockDataBytesLimit - blockSize(nil)) / (txSize + 8)
	txs := make([]types.Tx, count)

	for i := range txs {
		txs[i] = make(types.Tx, txSize)
	}

	// Safety net in case the estimate ever overshoots.
	for len(txs) > 0 && blockSize(txs) > types.MaxBlockDataBytesLimit {
		txs = txs[:len(txs)-1]
	}

	block := types.MakeBlock(1, txs, &types.Commit{})
	sized, err := amino.MarshalSized(block)
	require.NoError(t, err)
	require.LessOrEqual(t, int64(len(sized)), types.MaxBlockDataBytesLimit,
		"test setup: block must be decodable as a proposal")

	msgBytes := amino.MustMarshalAny(&bcBlockResponseMessage{Block: block})
	t.Logf("block %d bytes, block-response message %d bytes, maxMsgSize %d bytes (%d spare)",
		len(sized), len(msgBytes), maxMsgSize, maxMsgSize-len(msgBytes))

	assert.LessOrEqual(t, len(msgBytes), maxMsgSize,
		"the largest committable block must fit in the fast-sync envelope; "+
			"maxBlockMsgOverhead is too small")

	// maxMsgSize doubles as the blockchain channel's RecvMessageCapacity, so it
	// is also the per-connection recv exposure of this reactor. A lower bound
	// alone would be satisfied by the old MaxBlockSizeBytes (100MB), so assert
	// it stays tight: no more than the envelope allowance above the largest
	// block the chain can commit.
	assert.LessOrEqual(t, int64(maxMsgSize), types.MaxBlockDataBytesLimit+maxBlockMsgOverhead,
		"maxMsgSize doubles as RecvMessageCapacity; keeping it tight is what "+
			"bounds the blockchain channel's share of the recv buffer budget")

	// And the same message must survive the receive-side decode limit.
	_, err = decodeMsg(msgBytes)
	assert.NoError(t, err)
}
