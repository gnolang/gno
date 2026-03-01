// package strawpoll eases creation and management of polls with Gno.
package strawpoll

import (
	"errors"
	"strconv"
	"strings"

	"gno.land/p/demo/strawpoll"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/txlink"
)

var (
	// polls is a list of polls in the blockchain.
	polls          = make([]strawpoll.Poll, 0)
	createPollLink = txlink.NewLink("CreatePoll").
			AddArgs("question", "What is your favourite colour?", "choices", "Orange,Yellow").URL()
	voteLink = txlink.NewLink("Vote").
			AddArgs("pollNumber", "1", "choice", "Orange").URL()
	ErrPollNotFound = errors.New("poll not found")
)

// CreatePoll creates a new poll. It accepts a question and a comma-separated list of choices.
func CreatePoll(_ realm, question string, choices string) error {
	poll, err := strawpoll.New(question, strings.Split(choices, ","))
	if err != nil {
		return err
	}

	polls = append(polls, poll)

	return nil
}

// Vote allows a user to vote on a specific poll, identified by its position in the list.
func Vote(_ realm, pollNumber int64, choice string) error {
	pollIdx := pollNumber - 1

	if pollIdx >= int64(len(polls)) {
		return ErrPollNotFound
	}

	poll := &polls[pollIdx]

	return poll.Vote(choice)
}

const maxPollsRendered = 3

// Render provides a detailed view of recent polls.
func Render(_ string) string {
	result := "# Latest polls created\n"

	// Show the most recent polls
	startIdx := len(polls) - maxPollsRendered
	if startIdx < 0 {
		startIdx = 0
	}

	for id, poll := range polls[startIdx:] {
		pollNumber := 1 + startIdx + id
		result += "# " + strconv.Itoa(pollNumber) + ")\n" + poll.Markdown()
		result += "\n---\n\n"
	}

	if len(polls) == 0 {
		result += "No polls have been created yet. Be the first!\n"
	}

	result += ufmt.Sprintf(`# Actions
- [Create a poll](%s)
- [Vote](%s)
`, createPollLink, voteLink)

	return result
}
