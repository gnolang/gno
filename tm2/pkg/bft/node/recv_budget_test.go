package node

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cs "github.com/gnolang/gno/tm2/pkg/bft/consensus"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/p2p"
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
// the largest real message each channel can carry. The consensus and mempool
// channels' RecvMessageCapacity is that message plus a small allowance, so
// their terms are read from the reactors; the blockchain and discovery
// channels' capacities are looser, so theirs are the largest messages.
func TestDefaultBudgetCoversWorstLegalConfig(t *testing.T) {
	t.Parallel()

	const (
		// The bcBlockResponseMessage envelope around a length-prefixed block
		// measures 25 bytes; round up generously.
		envelope = 1 << 10

		// discovery messages carry at most maxPeersShared (30) addresses.
		discovery = 64 << 10
	)

	capacities := func(descs []*p2p.ChannelDescriptor) (sum int) {
		for _, desc := range descs {
			sum += desc.FillDefaults().RecvMessageCapacity
		}
		return sum
	}

	worstCase := int(types.MaxBlockDataBytesLimit) + envelope + // blockchain: the largest legal block
		capacities((&cs.ConsensusReactor{}).GetChannels()) + // consensus: its four channels
		capacities((&mempl.Reactor{}).GetChannels()) + // mempool: the largest tx
		discovery

	budget := conn.DefaultMConnConfig().MaxRecvBufferBytes

	t.Logf("worst legal concurrent assembly %d bytes vs budget %d (%d spare)",
		worstCase, budget, budget-worstCase)

	assert.LessOrEqual(t, worstCase, budget,
		"a legal chain configuration can exceed the per-connection recv budget, "+
			"so healthy peers would be disconnected with %q",
		"total recving buffer budget exceeded")
}
