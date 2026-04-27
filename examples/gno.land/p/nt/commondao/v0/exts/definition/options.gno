package definition

import (
	"strconv"
	"time"

	"gno.land/p/nt/commondao/v0"
)

// Option configures CommonDAO proposal definitions.
type Option func(*Definition)

// WithVoteChoices configures a proposal to use custom voting choices.
// It panics when there are less than two choices or when a voting choice is duplicated.
func WithVoteChoices(choices []commondao.VoteChoice) Option {
	count := len(choices)
	if count < 2 {
		panic("proposal requires at least two voting choices: got " + strconv.Itoa(count))
	}

	return func(d *Definition) {
		seen := make(map[commondao.VoteChoice]struct{}, count)
		d.voteChoices = make([]commondao.VoteChoice, count)
		for i, choice := range choices {
			if _, found := seen[choice]; found {
				panic("proposal voting choice is duplicated: " + string(choice))
			}

			seen[choice] = struct{}{}
			d.voteChoices[i] = choice
		}
	}
}

// WithVotingPeriod configures the voting period of a proposal.
// It panics when voting period is negative.
func WithVotingPeriod(period time.Duration) Option {
	if period < 0 {
		// Zero allows unit testing of proposals
		panic("voting period of proposal must be greater or equal to zero")
	}

	return func(d *Definition) {
		d.votingPeriod = period
	}
}

// WithTally configures a proposal definition to use a custom vote tallying.
// New definitions tally using simple majority (51%) by default.
// It panics when tally callback is nil.
func WithTally(cb TallyFunc) Option {
	if cb == nil {
		panic("proposal tally callback is nil")
	}

	return func(d *Definition) {
		d.tallyCb = cb
	}
}

// WithValidation configures a proposal definition to support proposal validation.
// Validation is done before execution and normally also during proposal rendering.
// It panics when validation callback is nil.
func WithValidation(cb func() error) Option {
	if cb == nil {
		panic("proposal validation callback is nil")
	}

	return func(d *Definition) {
		d.validateCb = cb
	}
}

// WithExecutor configures a proposal definition to support proposal execution.
// It panics when executor callback is nil.
func WithExecutor(cb commondao.ExecFunc) Option {
	if cb == nil {
		panic("proposal executor callback is nil")
	}

	return func(d *Definition) {
		d.executeCb = cb
	}
}
