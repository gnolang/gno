// package strawpoll eases creation and management of polls with Gno.
package strawpoll

import (
	"errors"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/txlink"
)

// Poll represents a blueprint for a single poll.
type Poll struct {
	Creator  std.Address          // The address of the person who created the poll.
	Question string               // The poll's question.
	Options  []string             // A list of valid options for this poll.
	Votes    map[string]int64     // A map to store the vote count for each option.
	Voters   map[std.Address]bool // A map to track who has voted *in this specific poll*.
}

func (poll Poll) String() string {
	var result string
	var totalVotes int64

	for _, count := range poll.Votes {
		totalVotes += count
	}

	result += "| Option | Votes | Percentage |\n"
	result += "|:---|:---|:---|\n"

	// Display each option as a row in the new table format.
	for _, option := range poll.Options {
		voteCount := poll.Votes[option]

		var percentage int64 = 0
		if totalVotes > 0 {
			percentage = (voteCount * 100) / totalVotes
		}

		votes := strconv.FormatInt(voteCount, 10)
		distribution := strconv.FormatInt(percentage, 10) + "%"

		result += "| **" + option + "** | " + votes + " | " + distribution + " |\n"
	}

	return result
}

// polls is a list of polls in the blockchain.
var polls []Poll

var (
	ErrQuestionMustBeProvided         = errors.New("poll question must be provided")
	ErrMiniminalOptionsMustBeProvided = errors.New("at least two comma-separated options must be provided")
	ErrPollNotFound                   = errors.New("poll not found")
	ErrAlreadyVoted                   = errors.New("you have already voted in this poll")
)

// CreatePoll creates a new poll. It accepts a question and a comma-separated list of choices.
func CreatePoll(_ realm, question string, choices string) error {
	caller := std.PreviousRealm().Address()
	options := strings.Split(choices, ",")

	isQuestionProvided := question != ""
	if !isQuestionProvided {
		return ErrQuestionMustBeProvided
	}

	areMinimalOptionsProvided := len(options) >= 2
	if !areMinimalOptionsProvided {
		return ErrMiniminalOptionsMustBeProvided
	}

	poll := Poll{
		Creator:  caller,
		Question: question,
		Options:  options,
		Votes:    make(map[string]int64),
		Voters:   make(map[std.Address]bool),
	}

	for _, option := range options {
		poll.Votes[option] = 0
	}

	polls = append(polls, poll)

	return nil
}

// Vote allows a user to vote on a specific poll, identified by its position in the list.
func Vote(_ realm, pollNumber int64, choice string) error {
	pollIdx := pollNumber - 1
	caller := std.PreviousRealm().Address()

	if pollIdx >= int64(len(polls)) {
		return ErrPollNotFound
	}

	poll := &polls[pollIdx]

	_, hasAlreadyVoted := poll.Voters[caller]

	if hasAlreadyVoted {
		return ErrAlreadyVoted
	}

	_, isOptionAvailable := poll.Votes[choice]

	if !isOptionAvailable {
		return errors.New("the option you voted for does not exist in this poll")
	}

	poll.Votes[choice]++
	poll.Voters[caller] = true

	return nil
}

const maxPollsRendered = 3

// Render provides a detailed view of all polls.
func Render(_ string) string {
	result := "# Latest polls created\n"

	// Show the most recent polls
	startIdx := len(polls) - maxPollsRendered
	if startIdx < 0 {
		startIdx = 0
	}

	for id, poll := range polls[startIdx:] {
		pollNumber := 1 + startIdx + id
		result += "## " + strconv.Itoa(pollNumber) + ") " + poll.Question + "\n\n"
		result += poll.String()
		result += "\n---\n\n"
	}

	if len(polls) == 0 {
		result += "No polls have been created yet. Be the first!\n"
	}

	createPollLink := txlink.NewLink("CreatePoll").
		AddArgs(
			"question", "What is your favourite colour?",
			"choices", "Orange,Yellow",
		).URL()

	voteLink := txlink.NewLink("Vote").
		AddArgs(
			"pollNumber", "1",
			"choice", "Orange",
		).URL()

	result += ufmt.Sprintf(`# Actions
- [Create a poll](%s)
- [Vote](%s)
`, createPollLink, voteLink)

	return result
}
