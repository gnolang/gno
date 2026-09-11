package node

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
)

// TestDefaultBudgetCoversWorstLegalConfig pins the derivation behind
// conn.defaultMaxRecvBufferBytes, which that constant's comment spells out but
// cannot check: p2p/conn sits below bft, so it cannot import the reactor and
// consensus-param constants its budget is sized against. This test lives here,
// in the package that wires every reactor onto the switch, and fails if the two
// sides ever drift apart.
//
// The budget bounds the bytes a connection can hold across all channels while
// messages are still being assembled. A connection assembles at most one
// incomplete message per channel, so the legitimate worst case is the sum of
// the largest real message each channel can carry -- not the sum of the
// channels' RecvMessageCapacity, which is deliberately looser.
func TestDefaultBudgetCoversWorstLegalConfig(t *testing.T) {
	t.Parallel()

	const (
		// bft/consensus maxMsgSize, over the state, data, vote and
		// vote-set-bits channels. Keep in sync with consensus/reactor.go.
		consensusMaxMsgSize = 1 << 20
		consensusChannels   = 4

		// The bcBlockResponseMessage envelope around a length-prefixed block
		// measures 25 bytes; round up generously.
		envelope = 1 << 10

		// discovery messages carry at most maxPeersShared (30) addresses.
		discovery = 64 << 10
	)

	// The largest values ValidateConsensusParams will accept.
	maxDataBytes := types.MaxBlockDataBytesLimit
	maxTxBytes := types.MaxBlockDataBytesLimit - types.MaxBlockOverheadBytes

	worstCase := int(maxDataBytes) + envelope + // blockchain: a committed block
		consensusChannels*consensusMaxMsgSize + // consensus
		int(maxTxBytes) + envelope + // mempool: a single tx
		discovery

	budget := conn.DefaultMConnConfig().MaxRecvBufferBytes

	t.Logf("worst legal concurrent assembly %d bytes vs budget %d (%d spare)",
		worstCase, budget, budget-worstCase)

	assert.LessOrEqual(t, worstCase, budget,
		"a legal chain configuration can exceed the per-connection recv budget, "+
			"so healthy peers would be disconnected with %q",
		"total recving buffer budget exceeded")
}
