package definition

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/nt/commondao/v0"
)

// DefaultVotingPeriod defines the default voting period for proposals.
const DefaultVotingPeriod = time.Hour * 24 * 7

var (
	ErrBodyIsRequired  = errors.New("proposal body is required")
	ErrTitleIsRequired = errors.New("proposal title is required")
)

var defaultVoteChoices = []commondao.VoteChoice{
	commondao.ChoiceYes,
	commondao.ChoiceNo,
	commondao.ChoiceAbstain,
}

// TallyFunc defines a function to handle proposal votes tallying.
type TallyFunc func(commondao.VotingContext) (passes bool, _ error)

// New creates a new proposal definition.
//
// By default, new definitions have a voting period of 7 days, allowing YES, NO and
// ABSTAIN votes, tallying those votes using an absolute majority of more than 50% of
// member votes, considering that a proposal passes when the majority of the votes are YES.
//
// Default definition behavior can be configured by setting custom options.
func New(title, body string, options ...Option) (Definition, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Definition{}, ErrTitleIsRequired
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return Definition{}, ErrBodyIsRequired
	}

	def := Definition{
		title:        title,
		body:         body,
		votingPeriod: DefaultVotingPeriod,
		voteChoices:  defaultVoteChoices,
		tallyCb:      TallyByAbsoluteMajority,
	}
	for _, apply := range options {
		apply(&def)
	}
	return def, nil
}

// MustNew creates a new proposal definition or panics on error.
func MustNew(title, body string, options ...Option) Definition {
	def, err := New(title, body, options...)
	if err != nil {
		panic(err)
	}
	return def
}

// Definition defines CommonDAO proposal types.
type Definition struct {
	title        string
	body         string
	votingPeriod time.Duration
	voteChoices  []commondao.VoteChoice
	tallyCb      TallyFunc
	validateCb   func() error
	executeCb    commondao.ExecFunc
}

// Title returns the proposal title.
func (d Definition) Title() string {
	return d.title
}

// Body returns proposal's body.
func (d Definition) Body() string {
	return d.body
}

// VotingPeriod returns the period where votes are allowed after proposal creation.
func (d Definition) VotingPeriod() time.Duration {
	return d.votingPeriod
}

// CustomVoteChoices returns a list of valid voting choices.
func (d Definition) CustomVoteChoices() []commondao.VoteChoice {
	return d.voteChoices
}

// Tally counts the number of votes and verifies if proposal passes.
func (d Definition) Tally(ctx commondao.VotingContext) (passes bool, _ error) {
	return d.tallyCb(ctx)
}

// Validate validates that the proposal is valid for the current state.
func (d Definition) Validate() error {
	if d.validateCb != nil {
		return d.validateCb()
	}
	return nil
}

// Executor returns a function to execute the proposal.
func (d Definition) Executor() commondao.ExecFunc {
	return d.executeCb
}

// TallyByAbsoluteMajority tallies votes by absolute majority.
// A quorum of 51% of member votes is required to tally votes.
// Tally considers that proposals passes when more than half of members votes YES.
func TallyByAbsoluteMajority(ctx commondao.VotingContext) (bool, error) {
	// Check if a quorum of 51% has been met
	if !commondao.IsQuorumReached(commondao.QuorumMoreThanHalf, ctx.VotingRecord, ctx.Members) {
		return false, commondao.ErrNoQuorum
	}

	// Tally votes by absolute majority, which requires 51% votes
	c, success := commondao.SelectChoiceByAbsoluteMajority(ctx.VotingRecord, ctx.Members.Size())
	if success {
		return c == commondao.ChoiceYes, nil
	}
	return false, nil
}
