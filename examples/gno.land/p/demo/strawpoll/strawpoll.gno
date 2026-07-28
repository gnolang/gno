package strawpoll

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/samcrew/piechart"
)

const (
	minimumOptions = 2
	maximumOptions = 10
)

var (
	ErrQuestionMustBeProvided       = errors.New("poll question must be provided")
	ErrMinimumOptionsMustBeProvided = errors.New("at least two comma-separated options must be provided")
	ErrMaximumOptionsExceeded       = ufmt.Errorf("a maximum of %d options can be provided", maximumOptions)
	ErrOptionNotAvailable           = errors.New("the option you voted for does not exist in this poll")
	ErrAlreadyVoted                 = errors.New("you have already voted in this poll")

	pieColors = []string{
		"#4E79A7", "#F28E2B", "#E15759", "#76B7B2",
		"#59A14F", "#EDC949", "#AF7AA1", "#FF9DA7",
		"#9C755F", "#BAB0AB",
	}
)

// Poll represents a blueprint for a single poll.
type Poll struct {
	creator  std.Address // The address of the person who created the poll.
	question string      // The poll's question.
	options  []string    // A list of valid options for this poll.
	votes    *avl.Tree   // A map to store the vote count for each option.
	voters   *avl.Tree   // A map to track who has voted *in this specific poll*.
}

// VoteCount returns the number of votes for a specific option in the poll.
func (poll *Poll) VoteCount(option string) int64 {
	if count, ok := poll.votes.Get(option); ok {
		return count.(int64)
	}

	return 0
}

// New creates a new poll with the given question and options.
func New(question string, options []string) (Poll, error) {
	if question == "" {
		return Poll{}, ErrQuestionMustBeProvided
	}

	if len(options) < minimumOptions {
		return Poll{}, ErrMinimumOptionsMustBeProvided
	}

	if len(options) > maximumOptions {
		return Poll{}, ErrMaximumOptionsExceeded
	}

	votes := avl.NewTree()
	voters := avl.NewTree()

	for _, option := range options {
		votes.Set(option, int64(0))
	}

	return Poll{
		creator:  std.PreviousRealm().Address(),
		question: question,
		options:  options,
		votes:    votes,
		voters:   voters,
	}, nil
}

// Vote allows a user to cast their vote for a specific option in the poll.
func (poll *Poll) Vote(option string) error {
	caller := std.PreviousRealm().Address()

	if poll.voters.Has(caller.String()) {
		return ErrAlreadyVoted
	}

	count, ok := poll.votes.Get(option)
	if !ok {
		return ErrOptionNotAvailable
	}

	poll.votes.Set(option, count.(int64)+1)
	poll.voters.Set(caller.String(), true)

	return nil
}

// Markdown generates a Markdown representation of the poll's results, using a piechart.
func (poll *Poll) Markdown() string {
	slices := make([]piechart.PieSlice, len(poll.options))

	for i, option := range poll.options {
		slices[i] = piechart.PieSlice{
			Value: float64(poll.VoteCount(option)),
			Color: pieColors[i%len(pieColors)],
			Label: option,
		}
	}

	return piechart.Render(slices, poll.question)
}
