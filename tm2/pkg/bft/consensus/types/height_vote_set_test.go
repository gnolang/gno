package cstypes

import (
	"errors"
	"fmt"
	"os"
	"testing"

	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
)

var config *cfg.Config // NOTE: must be reset for each _test.go file

func TestMain(m *testing.M) {
	config, _ = cfg.ResetTestRoot("consensus_height_vote_set_test")
	code := m.Run()
	os.RemoveAll(config.RootDir)
	os.Exit(code)
}

func TestPeerCatchupRounds(t *testing.T) {
	t.Parallel()

	valSet, privVals := types.RandValidatorSet(10, 1)

	hvs := NewHeightVoteSet(config.ChainID(), 1, valSet)

	vote999_0 := makeVoteHR(t, 1, 999, privVals, 0)
	added, err := hvs.AddVote(vote999_0, "peer1")
	if !added || err != nil {
		t.Error("Expected to successfully add vote from peer", added, err)
	}

	vote1000_0 := makeVoteHR(t, 1, 1000, privVals, 0)
	added, err = hvs.AddVote(vote1000_0, "peer1")
	if !added || err != nil {
		t.Error("Expected to successfully add vote from peer", added, err)
	}

	vote1001_0 := makeVoteHR(t, 1, 1001, privVals, 0)
	added, err = hvs.AddVote(vote1001_0, "peer1")
	if !errors.Is(err, ErrGotVoteFromUnwantedRoundError) {
		t.Errorf("Expected ErrGotVoteFromUnwantedRoundError, but got %v", err)
	}
	if added {
		t.Error("Expected to *not* add vote from peer, too many catchup rounds.")
	}

	added, err = hvs.AddVote(vote1001_0, "peer2")
	if !added || err != nil {
		t.Error("Expected to successfully add vote from another peer")
	}
}

func makeVoteHR(t *testing.T, height int64, round int, privVals []types.PrivValidator, valIndex int) *types.Vote {
	t.Helper()

	privVal := privVals[valIndex]
	vote := &types.Vote{
		ValidatorAddress: privVal.PubKey().Address(),
		ValidatorIndex:   valIndex,
		Height:           height,
		Round:            round,
		Timestamp:        tmtime.Now(),
		Type:             types.PrecommitType,
		BlockID:          types.BlockID{Hash: []byte("fakehash"), PartsHeader: types.PartSetHeader{}},
	}
	chainID := config.ChainID()
	if err := privVal.SignVote(chainID, vote); err != nil {
		panic(fmt.Sprintf("Error signing vote: %v", err))
	}
	return vote
}
