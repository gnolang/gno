package proposal_test

import (
	"testing"
	"time"

	"gno.land/p/nt/commondao"
	"gno.land/p/nt/commondao/exts/proposal"
	"gno.land/p/nt/urequire"
)

func TestDefinitionOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   func() proposal.DefinitionOption
		panicMsg string
	}{
		{
			name: "custom voting choices",
			option: func() proposal.DefinitionOption {
				return proposal.WithVoteChoices([]commondao.VoteChoice{"A", "B", "C"})
			},
		},
		{
			name: "no voting choices",
			option: func() proposal.DefinitionOption {
				return proposal.WithVoteChoices([]commondao.VoteChoice{})
			},
			panicMsg: "proposal requires at least two voting choices: got 0",
		},
		{
			name: "one voting choice",
			option: func() proposal.DefinitionOption {
				return proposal.WithVoteChoices([]commondao.VoteChoice{"A"})
			},
			panicMsg: "proposal requires at least two voting choices: got 1",
		},
		{
			name: "duplicated voting choice",
			option: func() proposal.DefinitionOption {
				return proposal.WithVoteChoices([]commondao.VoteChoice{"A", "B", "A"})
			},
			panicMsg: "proposal voting choice is duplicated: A",
		},
		{
			name: "voting period",
			option: func() proposal.DefinitionOption {
				return proposal.WithVotingPeriod(time.Hour * 24)
			},
		},
		{
			name: "zero voting period",
			option: func() proposal.DefinitionOption {
				return proposal.WithVotingPeriod(0)
			},
		},
		{
			name: "negative voting period",
			option: func() proposal.DefinitionOption {
				return proposal.WithVotingPeriod(-time.Hour)
			},
			panicMsg: "voting period of proposal must be greater or equal to zero",
		},
		{
			name: "tally callback",
			option: func() proposal.DefinitionOption {
				return proposal.WithTally(func(commondao.VotingContext) (bool, error) { return true, nil })
			},
		},
		{
			name: "nil tally callback",
			option: func() proposal.DefinitionOption {
				return proposal.WithTally(nil)
			},
			panicMsg: "proposal tally callback is nil",
		},
		{
			name: "validator callback",
			option: func() proposal.DefinitionOption {
				return proposal.WithValidation(func() error { return nil })
			},
		},
		{
			name: "nil validator callback",
			option: func() proposal.DefinitionOption {
				return proposal.WithValidation(nil)
			},
			panicMsg: "proposal validation callback is nil",
		},
		{
			name: "nil executor callback",
			option: func() proposal.DefinitionOption {
				return proposal.WithExecutor(nil)
			},
			panicMsg: "proposal executor callback is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var option proposal.DefinitionOption

			// Act
			fn := func() {
				option = tt.option()
				_ = proposal.MustNewDefinition("Title", "Body", option)
			}

			// Assert
			if tt.panicMsg != "" {
				urequire.PanicsWithMessage(t, tt.panicMsg, fn, "expect panic with message")
			} else {
				urequire.NotPanics(t, fn, "expect no panic")
				urequire.True(t, option != nil, "expect a non nil option")
			}
		})
	}
}
