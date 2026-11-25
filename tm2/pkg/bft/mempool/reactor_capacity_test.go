package mempool

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/stretchr/testify/require"
)

// TestChannelCapacityFitsLargestTx sends a TxMessage carrying the largest
// transaction valid consensus params admit, and a message of exactly the
// channel's capacity, through a connection built with the reactor's channel,
// and one byte more, which the connection refuses.
func TestChannelCapacityFitsLargestTx(t *testing.T) {
	t.Parallel()

	const capacity = 8_257_600

	// The largest Block.MaxTxBytes consensus params admit.
	maxTxBytes := types.MaxBlockDataBytesLimit - types.MaxBlockOverheadBytes
	params := types.DefaultConsensusParams()
	params.Block.MaxDataBytes = types.MaxBlockDataBytesLimit
	params.Block.MaxTxBytes = maxTxBytes
	require.NoError(t, types.ValidateConsensusParams(params))
	params.Block.MaxTxBytes = maxTxBytes + 1
	require.Error(t, types.ValidateConsensusParams(params))

	msg := amino.MustMarshalAny(&TxMessage{Tx: make(types.Tx, maxTxBytes)})
	t.Logf("TxMessage %d bytes, %d of them past the transaction, %d under capacity",
		len(msg), int64(len(msg))-maxTxBytes, capacity-len(msg))

	descs := (&Reactor{}).GetChannels()
	require.Len(t, descs, 1)
	desc := *descs[0]

	require.LessOrEqual(t, len(msg), capacity, "largest transaction")
	require.NoError(t, p2pTesting.Deliver(t, desc, capacity), "at capacity")
	require.ErrorContains(t, p2pTesting.Deliver(t, desc, capacity+1), "exceeds available capacity",
		"one byte past capacity")
}
