package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommitToVoteSetTamperedValidatorAddress rewrites one precommit's
// unsigned ValidatorAddress: VerifyCommit must reject the commit, and
// CommitToVoteSet must rebuild the genuine vote set from the positions.
func TestCommitToVoteSetTamperedValidatorAddress(t *testing.T) {
	t.Parallel()

	lastID := makeBlockIDRandom()
	h := int64(3)

	voteSet, valSet, vals := randVoteSet(h-1, 1, PrecommitType, 10, 1)
	commit, err := MakeCommit(lastID, h-1, 1, voteSet, vals)
	require.NoError(t, err)
	chainID := voteSet.ChainID()

	require.NoError(t, valSet.VerifyCommit(chainID, lastID, h-1, commit))

	tamperAddress := func(addr Address) *Commit {
		tamperedSigs := make([]*CommitSig, len(commit.Precommits))
		for i, sig := range commit.Precommits {
			if sig == nil {
				continue
			}
			copied := *sig
			if i == 0 {
				// The signature stays valid: the address is not signed.
				copied.ValidatorAddress = addr
			}
			tamperedSigs[i] = &copied
		}
		return NewCommit(commit.BlockID, tamperedSigs)
	}

	// A peer can rewrite the address to another validator's or blank it.
	for _, addr := range []Address{vals[9].PubKey().Address(), {}} {
		tampered := tamperAddress(addr)

		require.Error(t, valSet.VerifyCommit(chainID, lastID, h-1, tampered),
			"VerifyCommit accepted a commit with a tampered ValidatorAddress")

		assert.NotPanics(t, func() {
			voteSet2 := CommitToVoteSet(chainID, tampered, valSet)
			assert.True(t, voteSet2.HasTwoThirdsMajority())
			assert.Equal(t, vals[0].PubKey().Address(), voteSet2.GetByIndex(0).ValidatorAddress)
			for i := range vals {
				assert.Equal(t, voteSet.GetByIndex(i), voteSet2.GetByIndex(i),
					"reconstructed vote %d differs from the original", i)
			}
		})
	}
}
