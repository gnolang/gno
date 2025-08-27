package strawpoll

import (
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
)

const minimumOptions = 2

var (
	ErrQuestionMustBeProvided       = errors.New("poll question must be provided")
	ErrMinimumOptionsMustBeProvided = errors.New("at least two comma-separated options must be provided")
	ErrOptionNotAvailable           = errors.New("the option you voted for does not exist in this poll")
	ErrAlreadyVoted                 = errors.New("you have already voted in this poll")
)

// Poll represents a blueprint for a single poll.
type Poll struct {
	creator  std.Address          // The address of the person who created the poll.
	question string               // The poll's question.
	options  []string             // A list of valid options for this poll.
	votes    map[string]int64     // A map to store the vote count for each option.
	voters   map[std.Address]bool // A map to track who has voted *in this specific poll*.
}

// OptionCount returns the number of votes for a specific option in the poll.
func (poll *Poll) OptionCount(option string) int64 {
	return poll.votes[option]
}

// New creates a new poll with the given question and options.
func New(question string, options []string) (Poll, error) {
	if question == "" {
		return Poll{}, ErrQuestionMustBeProvided
	}

	if len(options) < minimumOptions {
		return Poll{}, ErrMinimumOptionsMustBeProvided
	}

	votes := make(map[string]int64)
	voters := make(map[std.Address]bool)

	for _, option := range options {
		votes[option] = 0
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

	_, hasAlreadyVoted := poll.voters[caller]
	if hasAlreadyVoted {
		return ErrAlreadyVoted
	}

	_, isOptionAvailable := poll.votes[option]
	if !isOptionAvailable {
		return ErrOptionNotAvailable
	}

	poll.votes[option]++
	poll.voters[caller] = true

	return nil
}

// Markdown generates a Markdown representation of the poll's results.
func (poll *Poll) Markdown() string {
	totalVotes := poll.totalVotes()

	result := poll.question + "\n"
	result += "| Option | Votes | Percentage |\n"
	result += "|:---|:---|:---|\n"

	// Display each option as a row in the new table format.
	for _, option := range poll.options {
		voteCount := poll.votes[option]

		var percentage int64 = 0
		if totalVotes > 0 {
			percentage = (voteCount * 100) / totalVotes
		}

		votes := strconv.FormatInt(voteCount, 10)
		distribution := strconv.FormatInt(percentage, 10) + "%"

		result += ufmt.Sprintf("| **%s** | %s | %s |\n", option, votes, distribution)
	}

	return result
}

func (poll *Poll) totalVotes() (total int64) {
	for _, count := range poll.votes {
		total += count
	}

	return total
}
