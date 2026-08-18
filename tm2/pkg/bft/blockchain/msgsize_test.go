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
	// the largest one that could have passed a proposal decode.
	const txSize = 1 << 16
	txs := make([]types.Tx, 0, types.MaxBlockDataBytesLimit/txSize)
	for {
		candidate := append(txs, make(types.Tx, txSize)) //nolint:gocritic // intentional: probing size
		sized, err := amino.MarshalSized(types.MakeBlock(1, candidate, &types.Commit{}))
		require.NoError(t, err)
		if int64(len(sized)) > types.MaxBlockDataBytesLimit {
			break
		}
		txs = candidate
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

	// And the same message must survive the receive-side decode limit.
	_, err = decodeMsg(msgBytes)
	assert.NoError(t, err)
}
