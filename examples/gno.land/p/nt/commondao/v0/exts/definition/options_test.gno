package definition_test

import (
	"testing"
	"time"

	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/commondao/v0/exts/definition"
	"gno.land/p/nt/urequire/v0"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   func() definition.Option
		panicMsg string
	}{
		{
			name: "custom voting choices",
			option: func() definition.Option {
				return definition.WithVoteChoices([]commondao.VoteChoice{"A", "B", "C"})
			},
		},
		{
			name: "no voting choices",
			option: func() definition.Option {
				return definition.WithVoteChoices([]commondao.VoteChoice{})
			},
			panicMsg: "proposal requires at least two voting choices: got 0",
		},
		{
			name: "one voting choice",
			option: func() definition.Option {
				return definition.WithVoteChoices([]commondao.VoteChoice{"A"})
			},
			panicMsg: "proposal requires at least two voting choices: got 1",
		},
		{
			name: "duplicated voting choice",
			option: func() definition.Option {
				return definition.WithVoteChoices([]commondao.VoteChoice{"A", "B", "A"})
			},
			panicMsg: "proposal voting choice is duplicated: A",
		},
		{
			name: "voting period",
			option: func() definition.Option {
				return definition.WithVotingPeriod(time.Hour * 24)
			},
		},
		{
			name: "zero voting period",
			option: func() definition.Option {
				return definition.WithVotingPeriod(0)
			},
		},
		{
			name: "negative voting period",
			option: func() definition.Option {
				return definition.WithVotingPeriod(-time.Hour)
			},
			panicMsg: "voting period of proposal must be greater or equal to zero",
		},
		{
			name: "tally callback",
			option: func() definition.Option {
				return definition.WithTally(func(commondao.VotingContext) (bool, error) { return true, nil })
			},
		},
		{
			name: "nil tally callback",
			option: func() definition.Option {
				return definition.WithTally(nil)
			},
			panicMsg: "proposal tally callback is nil",
		},
		{
			name: "validator callback",
			option: func() definition.Option {
				return definition.WithValidation(func() error { return nil })
			},
		},
		{
			name: "nil validator callback",
			option: func() definition.Option {
				return definition.WithValidation(nil)
			},
			panicMsg: "proposal validation callback is nil",
		},
		{
			name: "nil executor callback",
			option: func() definition.Option {
				return definition.WithExecutor(nil)
			},
			panicMsg: "proposal executor callback is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var option definition.Option

			// Act
			fn := func() {
				option = tt.option()
				_ = definition.MustNew("Title", "Body", option)
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
