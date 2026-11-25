package state_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/mempool/mock"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// ValidateBlock must reject a LastCommit precommit relabelled with another
// validator's unsigned address.
func TestValidateBlockTamperedLastCommitAddress(t *testing.T) {
	t.Parallel()

	proxyApp := newTestApp()
	require.NoError(t, proxyApp.Start())
	defer proxyApp.Stop()

	state, stateDB, privVals := makeState(4, 1)
	blockExec := sm.NewBlockExecutor(stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})
	state, _, commit, err := makeAndCommitGoodBlock(state, 1, types.NewCommit(types.BlockID{}, nil),
		state.Validators.GetProposer().Address, blockExec, privVals)
	require.NoError(t, err)

	// The honest block at height 2 is accepted.
	good, _ := state.MakeBlock(2, makeTxs(2), commit, state.Validators.GetProposer().Address)
	require.NoError(t, state.ValidateBlock(good))

	// Same commit, precommit 0 relabelled with validator 1's address; every
	// signature still verifies, because the address is not signed.
	sigs := make([]*types.CommitSig, len(commit.Precommits))
	for i, s := range commit.Precommits {
		c := *s
		sigs[i] = &c
	}
	_, other := state.LastValidators.GetByIndex(1)
	sigs[0].ValidatorAddress = other.Address
	tampered, _ := state.MakeBlock(2, makeTxs(2), types.NewCommit(commit.BlockID, sigs),
		state.Validators.GetProposer().Address)

	// VerifyCommit checks each address against its position, so the block is
	// invalid. Without that check ValidateBlock accepts it:
	// require.NoError(t, state.ValidateBlock(tampered))
	require.ErrorContains(t, state.ValidateBlock(tampered), "invalid validator address")
}
