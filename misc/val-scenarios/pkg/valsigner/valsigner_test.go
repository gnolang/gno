package valsigner

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

func TestClassifySignBytesProposal(t *testing.T) {
	t.Parallel()

	proposal := types.NewProposal(7, 2, -1, types.BlockID{
		Hash: []byte("blockhash"),
		PartsHeader: types.PartSetHeader{
			Total: 1,
			Hash:  []byte("partshash"),
		},
	})

	signBytes := proposal.SignBytes("dev")
	target, err := ClassifySignBytes(signBytes)
	if err != nil {
		t.Fatalf("ClassifySignBytes() error = %v", err)
	}

	if target.Phase != PhaseProposal || target.Height != 7 || target.Round != 2 {
		t.Fatalf("unexpected target: %+v", target)
	}
}

func TestClassifySignBytesVote(t *testing.T) {
	t.Parallel()

	vote := &types.Vote{
		Type:      types.PrecommitType,
		Height:    11,
		Round:     3,
		Timestamp: time.Now().UTC(),
		BlockID: types.BlockID{
			Hash: []byte("blockhash"),
			PartsHeader: types.PartSetHeader{
				Total: 1,
				Hash:  []byte("partshash"),
			},
		},
	}

	signBytes := vote.SignBytes("dev")
	target, err := ClassifySignBytes(signBytes)
	if err != nil {
		t.Fatalf("ClassifySignBytes() error = %v", err)
	}

	if target.Phase != PhasePrecommit || target.Height != 11 || target.Round != 3 {
		t.Fatalf("unexpected target: %+v", target)
	}
}

func TestRuleMatches(t *testing.T) {
	t.Parallel()

	height := int64(12)
	round := 1
	rule := Rule{
		Action: ActionDrop,
		Height: &height,
		Round:  &round,
	}

	if !rule.Matches(SignedTarget{Phase: PhasePrevote, Height: 12, Round: 1}) {
		t.Fatal("expected matching rule")
	}
	if rule.Matches(SignedTarget{Phase: PhasePrevote, Height: 12, Round: 2}) {
		t.Fatal("expected round mismatch")
	}
}

func TestParseRuleRequest(t *testing.T) {
	t.Parallel()

	rule, err := ParseRuleRequest(ruleRequest{
		Action: ActionDelay,
		Delay:  "1500ms",
	})
	if err != nil {
		t.Fatalf("ParseRuleRequest() error = %v", err)
	}
	if rule.Delay != 1500*time.Millisecond {
		t.Fatalf("unexpected delay: %v", rule.Delay)
	}
}
