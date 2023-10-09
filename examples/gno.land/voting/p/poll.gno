package poll

import (
	"gno.land/p/demo/avl"
	"std"
)

type Poll struct {
	title       string
	description string
	deadline    int64     // block height
	voters      *avl.Tree // addr -> yes / no (bool)
}

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

func NewPoll(title, description string, deadline int64) *Poll {
	return &Poll{
		title:       title,
		description: description,
		deadline:    deadline,
		voters:      avl.NewTree(),
	}
}

func (p *Poll) Vote(voter std.Address, vote bool) {
	p.Voters().Set(string(voter), vote)
}

// HasVoted vote: yes - true, no - false
func (p *Poll) HasVoted(address std.Address) (bool, bool) {
	vote, exists := p.Voters().Get(string(address))
	if exists {
		return true, vote.(bool)
	}
	return false, false
}

func (p Poll) VoteCount() (int, int) {
	var (
		yay int
		nay int
	)
	p.Voters().Iterate("", "", func(key string, value interface{}) bool {
		vote := value.(bool)

		if vote == true {
			yay = yay + 1
		} else {
			nay = nay + 1
		}
	})

	return yay, nay
}
