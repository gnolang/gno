package state_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/mempool/mock"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// skewable: powers 1, 10 and 1000, the strongest voting 20s late, so only a
// real-power weighting returns want.
type skewable struct {
	vals    *types.ValidatorSet
	commit  *types.Commit
	want    time.Time // the median an honest weighting produces
	weakest int       // the position holding the least power
}

func makeSkewable(t *testing.T, height int64) skewable {
	t.Helper()

	// Fixed, so two commits from this helper carry the same timestamps.
	base := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	s := skewable{want: base.Add(20 * time.Second)}
	s.vals, s.commit = signedCommit(t, height, []int64{1, 10, 1000}, func(i int, val *types.Validator) time.Time {
		switch val.VotingPower {
		case 1000:
			return s.want
		case 1:
			s.weakest = i
		}
		return base
	})
	return s
}

// signedCommit returns a validator set of the given powers and a commit signed
// by each of them, validator i voting at timestamp(i, validator).
func signedCommit(
	t *testing.T,
	height int64,
	powers []int64,
	timestamp func(i int, val *types.Validator) time.Time,
) (*types.ValidatorSet, *types.Commit) {
	t.Helper()

	vals := make([]*types.Validator, len(powers))
	privVals := make(map[crypto.Address]types.PrivValidator, len(powers))
	for i, power := range powers {
		pk := ed25519.GenPrivKeyFromSecret(fmt.Appendf(nil, "median%d", i))
		vals[i] = types.NewValidator(pk.PubKey(), power)
		privVals[pk.PubKey().Address()] = types.NewMockPVWithPrivKey(pk)
	}
	// NewValidatorSet orders by address, so a power lands at a position the
	// caller does not choose. Read them back by position.
	valSet := types.NewValidatorSet(vals)

	blockID := types.BlockID{
		Hash: tmhash.Sum([]byte("median time block")),
		PartsHeader: types.PartSetHeader{
			Total: 1,
			Hash:  tmhash.Sum([]byte("median time parts")),
		},
	}

	sigs := make([]*types.CommitSig, valSet.Size())
	for i := range valSet.Size() {
		_, val := valSet.GetByIndex(i)
		vote := &types.Vote{
			ValidatorAddress: val.Address,
			ValidatorIndex:   i,
			Height:           height,
			Round:            0,
			Timestamp:        timestamp(i, val),
			Type:             types.PrecommitType,
			BlockID:          blockID,
		}
		require.NoError(t, privVals[val.Address].SignVote(chainID, vote))
		sigs[i] = vote.CommitSig()
	}

	return valSet, types.NewCommit(blockID, sigs)
}

// VerifyCommit refuses a precommit whose unsigned CommitSig.ValidatorIndex
// differs from its position, and median time weights by position, so the index
// never moves the median even in a commit nobody verified.
func TestMedianTimeIgnoresCommitSigValidatorIndex(t *testing.T) {
	t.Parallel()

	const height = int64(5)
	honest := makeSkewable(t, height)

	require.NoError(t, honest.vals.VerifyCommit(chainID, honest.commit.BlockID, height, honest.commit))
	require.Equal(t, honest.want, sm.MedianTime(honest.commit, honest.vals))

	tests := []struct {
		name     string
		index    int
		lastOnly bool // tamper the last precommit alone, not every one
	}{
		{"past the end of the set", 999, false},
		{"negative", -1, false},
		// In range: every timestamp weighted with the least power, which moves
		// the median 20 seconds earlier.
		{"pointing at the weakest validator", honest.weakest, false},
		{"on the last precommit alone", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := makeSkewable(t, height)
			sigs := s.commit.Precommits
			if tt.lastOnly {
				sigs = sigs[len(sigs)-1:]
			}
			for _, sig := range sigs {
				sig.ValidatorIndex = tt.index
			}

			require.ErrorContains(t, s.vals.VerifyCommit(chainID, s.commit.BlockID, height, s.commit),
				"invalid validator index")
			require.Equal(t, s.want, sm.MedianTime(s.commit, s.vals))
		})
	}
}

// TestMedianTimeSkipsPrecommitsPastTheValidatorSet covers the positional lookup
// running off the end of the set. ValidateBlock rejects a commit of the wrong
// length before MedianTime sees it; MakeBlock takes whatever it is handed.
func TestMedianTimeSkipsPrecommitsPastTheValidatorSet(t *testing.T) {
	t.Parallel()

	s := makeSkewable(t, 5)
	s.commit.Precommits = append(s.commit.Precommits, s.commit.Precommits[0])

	require.Equal(t, s.want, sm.MedianTime(s.commit, s.vals))
}

// A far-future vote sorts after every honest one, whatever its UnixNano wraps
// to, so it never becomes the median.
func TestMedianTimeFarFutureVote(t *testing.T) {
	t.Parallel()

	const height = int64(5)
	base := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	// 2^64ns after base+500ms, in 2608: UnixNano wraps it back to base+500ms.
	farFuture := base.Add(math.MaxInt64).Add(math.MaxInt64).Add(2 + 500*time.Millisecond)
	require.Equal(t, base.Add(500*time.Millisecond).UnixNano(), farFuture.UnixNano())

	// Equal powers, so the far-future vote may sit at any position.
	vals, commit := signedCommit(t, height, []int64{10, 10, 10, 10}, func(i int, _ *types.Validator) time.Time {
		if i == 0 {
			return farFuture
		}
		return base.Add(time.Duration(i-1) * time.Second)
	})
	// The ordering, not the commit check, keeps a far-future vote out of the
	// median.
	require.NoError(t, vals.VerifyCommit(chainID, commit.BlockID, height, commit))

	// Votes at base, base+1s, base+2s and 2608: the second earliest.
	require.Equal(t, base.Add(time.Second), sm.MedianTime(commit, vals))
}

// TestBlockTimeIgnoresValidatorIndex checks that MakeBlock derives one block
// time whatever ValidatorIndex holds, and that ValidateBlock refuses the block
// whose LastCommit carries an index off its position.
func TestBlockTimeIgnoresValidatorIndex(t *testing.T) {
	t.Parallel()

	proxyApp := newTestApp()
	require.NoError(t, proxyApp.Start())
	defer proxyApp.Stop()

	state, stateDB, privVals := makeState(3, 1)
	blockExec := sm.NewBlockExecutor(stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})
	proposer := state.Validators.GetProposer().Address

	state, _, lastCommit, err := makeAndCommitGoodBlock(
		state, 1, types.NewCommit(types.BlockID{}, nil), proposer, blockExec, privVals)
	require.NoError(t, err)

	honest, _ := state.MakeBlock(2, makeTxs(2), lastCommit, proposer)

	for _, sig := range lastCommit.Precommits {
		sig.ValidatorIndex = 999
	}

	tampered := types.NewCommit(lastCommit.BlockID, lastCommit.Precommits)
	block, _ := state.MakeBlock(2, makeTxs(2), tampered, proposer)
	require.NotEqual(t, honest.LastCommitHash, block.LastCommitHash)
	require.Equal(t, honest.Time, block.Time)
	require.ErrorContains(t, state.ValidateBlock(block), "invalid validator index")
}
