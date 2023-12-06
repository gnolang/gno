package poll

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/poll"
	"gno.land/p/demo/ufmt"
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

// NewPoll - Creates a new Poll instance
func NewPoll(title, description string, deadline int64) string {
	// get block height
	if deadline <= std.GetHeight() {
		return "Error: Deadline has to be in the future."
	}

	// convert int ID to string used in AVL tree
	id := ufmt.Sprintf("%d", pollIDCounter)
	p := poll.NewPoll(title, description, deadline)

	// add new poll in avl tree
	polls.Set(id, p)

	// increment ID counter
	pollIDCounter = pollIDCounter + 1

	return ufmt.Sprintf("Successfully created poll #%s!", id)
}

// Vote - vote for a specific Poll
// yes - true, no - false
func Vote(pollID int, vote bool) string {
	// get txSender
	txSender := std.GetOrigCaller()

	id := ufmt.Sprintf("%d", pollID)
	// get specific Poll from AVL tree
	pollRaw, exists := polls.Get(id)

	if !exists {
		return "Error: Poll with specified doesn't exist."
	}

	// cast Poll into proper format
	poll, _ := pollRaw.(*poll.Poll)

	voted, _ := poll.HasVoted(txSender)
	if voted {
		return "Error: You've already voted!"
	}

	if poll.Deadline() <= std.GetHeight() {
		return "Error: Voting for this poll is closed."
	}

	// record vote
	poll.Vote(txSender, vote)

	// update Poll in tree
	polls.Set(id, poll)

	if vote == true {
		return ufmt.Sprintf("Successfully voted YAY for poll #%s!", id)
	}
	return ufmt.Sprintf("Successfully voted NAY for poll #%s!", id)
}
