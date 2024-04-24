package poll

import (
	"std"

	"gno.land/p/demo/avl"
)

// Main struct
type Poll struct {
	title       string
	description string
	deadline    int64     // block height
	voters      *avl.Tree // addr -> yes / no (bool)
}

// Getters
func (p Poll) Title() string {
	return p.title
}

func (p Poll) Description() string {
	return p.description
}

func (p Poll) Deadline() int64 {
	return p.deadline
}

func (p Poll) Voters() *avl.Tree {
	return p.voters
}

// Poll instance constructor
func NewPoll(title, description string, deadline int64) *Poll {
	return &Poll{
		title:       title,
		description: description,
		deadline:    deadline,
		voters:      avl.NewTree(),
	}
}

// Vote Votes for a user
func (p *Poll) Vote(voter std.Address, vote bool) {
	p.Voters().Set(voter.String(), vote)
}

// HasVoted vote: yes - true, no - false
func (p *Poll) HasVoted(address std.Address) (bool, bool) {
	vote, exists := p.Voters().Get(address.String())
	if exists {
		return true, vote.(bool)
	}
	return false, false
}

// VoteCount Returns the number of yay & nay votes
func (p Poll) VoteCount() (int, int) {
	var yay int

	p.Voters().Iterate("", "", func(key string, value interface{}) bool {
		vote := value.(bool)
		if vote == true {
			yay = yay + 1
		}
	})
	return yay, p.Voters().Size() - yay
}
