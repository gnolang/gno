package node

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/random"
)

// TestCreateProposalBlockFitsDecodeLimit is a regression test for a proposal
// block that fills Block.MaxDataBytes with tx data being undecodable by peers.
//
// MaxDataBytes is used two ways: CreateProposalBlock reaps up to that many raw
// tx bytes, while ConsensusState.addProposalBlockPart hands it to amino as the
// max size for the *whole* serialized block. The header, the LastCommit and
// amino's framing therefore have to come out of the same budget, or a full
// block is rejected by every peer and the round fails -- and since the next
// proposer reaps the same mempool, the chain stalls while it holds that much
// data.
func TestCreateProposalBlockFitsDecodeLimit(t *testing.T) {
	config, _ := cfg.ResetTestRoot("node_create_proposal_size")
	defer os.RemoveAll(config.RootDir)

	cc := proxy.NewLocalClientCreator(kvstore.NewKVStoreApplication())
	proxyApp := appconn.NewAppConns(cc)
	require.NoError(t, proxyApp.Start())
	defer proxyApp.Stop()

	logger := log.NewTestingLogger(t)

	const height int64 = 1

	// A small MaxDataBytes keeps the test fast; the arithmetic is the same at
	// the 2MB default.
	const maxDataBytes int64 = 64 * 1024

	state, stateDB := state(1, height)
	state.ConsensusParams.Block.MaxDataBytes = maxDataBytes
	proposerAddr, _ := state.Validators.GetByIndex(0)

	mempool := mempl.NewCListMempool(
		config.Mempool,
		proxyApp.Mempool(),
		state.LastBlockHeight,
		state.ConsensusParams.Block.MaxTxBytes,
		mempl.WithPreCheck(sm.TxPreCheck(state)),
	)
	mempool.SetLogger(logger)

	// Fill the mempool well past what one block can carry, so the reap is
	// bounded by MaxDataBytes rather than by what is available.
	const txLength = 512
	for range int(maxDataBytes) / txLength * 2 {
		require.NoError(t, mempool.CheckTx(random.RandBytes(txLength), nil))
	}

	blockExec := sm.NewBlockExecutor(stateDB, logger, proxyApp.Consensus(), mempool)

	commit := types.NewCommit(types.BlockID{}, nil)
	block, parts := blockExec.CreateProposalBlock(height, state, commit, proposerAddr)

	require.NotEmpty(t, block.Txs, "the proposal should carry txs")
	t.Logf("block: %d txs, %d serialized bytes, MaxDataBytes %d",
		len(block.Txs), parts.ByteSize(), maxDataBytes)

	// The block must still be valid, and it must not be empty just because the
	// overhead was reserved.
	require.NoError(t, state.ValidateBlock(block))
	assert.LessOrEqual(t, int64(parts.ByteSize()), maxDataBytes,
		"the serialized block must fit within MaxDataBytes")

	// This is exactly what a peer does once it has every part:
	// ConsensusState.addProposalBlockPart.
	var decoded types.Block
	_, err := amino.UnmarshalSizedReader(parts.GetReader(), &decoded, maxDataBytes)
	assert.NoError(t, err, "peers must be able to decode the proposal block")
}
