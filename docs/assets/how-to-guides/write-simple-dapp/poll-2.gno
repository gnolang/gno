package poll

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/poll"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

// state variables
var (
	polls         *avl.Tree // id -> Poll
	pollIDCounter seqid.ID
)

func init() {
	polls = avl.NewTree()
}

// NewPoll - Creates a new Poll instance
func NewPoll(title, description string, deadline int64) string {
	// get block height
	if deadline <= std.GetHeight() {
		panic("deadline has to be in the future")
	}

	// Generate int
	id := pollIDCounter.Next().String()
	p := poll.NewPoll(title, description, deadline)

	// add new poll in avl tree
	polls.Set(id, p)

	return ufmt.Sprintf("Successfully created poll #%s!", id)
}

// Vote - vote for a specific Poll
// yes - true, no - false
func Vote(id string, vote bool) string {
	// get txSender
	txSender := std.GetOrigCaller()

	// get specific Poll from AVL tree
	pollRaw, exists := polls.Get(id)

	if !exists {
		panic("poll with specified doesn't exist")
	}

	// cast Poll into proper format
	poll, _ := pollRaw.(*poll.Poll)

	voted, _ := poll.HasVoted(txSender)
	if voted {
		panic("you've already voted!")
	}

	if poll.Deadline() <= std.GetHeight() {
		panic("voting for this poll is closed")
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
