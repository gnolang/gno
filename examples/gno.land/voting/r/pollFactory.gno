package poll

import (
	"bytes"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/poll"
	"gno.land/p/demo/ufmt"
	"std"

	"strconv"
)

// state variables
var (
	polls         *avl.Tree // id -> Poll
	pollIDCounter int
)

func init() {
	polls = avl.NewTree()
	pollIDCounter = 0
}

func NewPoll(title, description string, deadline int64) string {
	// Get block height
	if deadline <= std.GetHeight() {
		return "Error: Deadline has to be in the future."
	}

	id := ufmt.Sprintf("%d", pollIDCounter)
	p := poll.NewPoll(title, description, deadline)

	// set new poll in avl tree
	polls.Set(id, p)
	pollIDCounter = pollIDCounter + 1

	return ufmt.Sprintf("Successfully created poll #%s!", id)
}

func Vote(pollID int, vote bool) string {
	txSender := std.GetOrigCaller()

	pollRaw, exists := polls.Get(strconv.Itoa(pollID))

	if !exists {
		return "Error: Poll with specified doesn't exist."
	}

	poll, _ := pollRaw.(*poll.Poll)

	voted, _ := poll.HasVoted(txSender)
	if voted {
		return "Error: You've already voted!"
	}

	if poll.Deadline() <= std.GetHeight() {
		return "Error: Voting for this poll is closed."
	}

	poll.Vote(txSender, vote)
	polls.Set(strconv.Itoa(pollID), poll)

	if vote == true {
		return ufmt.Sprintf("Successfully voted YAY for poll #%d!", pollID)
	}
	return ufmt.Sprintf("Successfully voted NAY for poll #%d!", pollID)

}

func Render(path string) string {
	if path == "" {
		return renderHome()
	}
	return "tf"
}

func renderHome() string {
	var b bytes.Buffer

	b.WriteString("# Polls!\n\n")

	if polls.Size() == 0 {
		b.WriteString("### No active polls currently!")
		return b.String()
	}
	polls.Iterate("", "", func(key string, value interface{}) bool {

		// cast raw data from tree into Whitelist struct
		p := value.(*poll.Poll)
		ddl := p.Deadline()

		yay, nay := p.VoteCount()

		b.WriteString(
			ufmt.Sprintf(
				"## Poll #%s: %s\n",
				key, // poll ID
				p.Title(),
			),
		)

		b.WriteString(
			ufmt.Sprintf("Voting until block: %d, current vote count: %d\n\n", p.Deadline(), p.Voters().Size()),
		)

		b.WriteString(
			ufmt.Sprintf("YAY votes: %d\n\n", yay),
		)
		b.WriteString(
			ufmt.Sprintf("NAY votes: %d\n\n", nay),
		)

		dropdown := "<details>\n<summary>Voter details</summary><br>"
		b.WriteString(dropdown)

		p.Voters().Iterate("", "", func(key string, value interface{}) bool {

			voter := key
			vote := value.(bool)

			if vote == true {
				b.WriteString(
					ufmt.Sprintf("<br>1. Voter %s voted YAY!\n", voter),
				)
			} else {
				b.WriteString(
					ufmt.Sprintf("<br>1. Voter %s voted NAY!\n", voter),
				)
			}

			b.WriteString("</details>\n")
			return false
		})

		return false
	})

	return b.String()
}
