package poll

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// Poll holds the data for a multi-option poll.
type Poll struct {
	Title       string
	Description string
	Deadline    time.Time // now a standard time.Time
	Creator     std.Address
	Options     []string  // e.g. ["Red", "Green", "Blue"]
	MaxChoices  int       // how many options a voter can select
	VoteTotal   int       // total number of votes cast
	Votes       *avl.Tree // key: voter address, value: []int (indices of Options)
}

// NewPoll constructs a Poll with the desired configuration.
func NewPoll(
	title string,
	description string,
	opts []string, // often an empty slice in this refactor
	maxChoices int,
	deadline time.Time, // now a time.Time
	creator std.Address,
) *Poll {
	if maxChoices < 1 {
		panic("maxChoices must be at least 1")
	}
	if deadline.Before(time.Now()) {
		panic("deadline must be in the future")
	}
	p := &Poll{
		Title:       title,
		Description: description,
		Deadline:    deadline,
		Creator:     creator,
		Options:     opts,
		MaxChoices:  maxChoices,
		VoteTotal:   0,
		Votes:       avl.NewTree(),
	}
	return p
}

// HasVoted returns true if the given address has already cast a vote.
func (p *Poll) HasVoted(addr std.Address) bool {
	_, exists := p.Votes.Get(addr.String())
	return exists
}

// Vote registers the given list of option indices for the address `voter`.
func (p *Poll) Vote(voter std.Address, choices []int) {
	if time.Now().After(p.Deadline) {
		panic("voting for this poll is closed")
	}
	if p.HasVoted(voter) {
		panic("you have already voted")
	}
	if len(choices) == 0 {
		panic("must select at least one option")
	}
	if len(choices) > p.MaxChoices {
		panic(ufmt.Sprintf("too many selections; the maximum is %d", p.MaxChoices))
	}
	for _, cIndex := range choices {
		if cIndex < 0 || cIndex >= len(p.Options) {
			panic("invalid option index")
		}
	}
	p.VoteTotal = p.VoteTotal + len(choices)
	p.Votes.Set(voter.String(), choices)
}

// Tally returns a slice counting how many votes each option has.
func (p *Poll) Tally() []int {
	tally := make([]int, len(p.Options))
	p.Votes.Iterate("", "", func(_ string, value interface{}) bool {
		choiceIndices := value.([]int)
		for _, idx := range choiceIndices {
			tally[idx]++
		}
		return false
	})
	return tally
}

// VoterCount returns how many unique addresses have cast votes.
func (p *Poll) VoterCount() int {
	return p.Votes.Size()
}

// Edit poll metadata (title, description, deadline, maxChoices),
// *not* the options. This is separate from adding/deleting/editing
// options one at a time.
func (p *Poll) Edit(
	editor std.Address,
	newTitle string,
	newDescription string,
	newDeadline time.Time, // now a time.Time
	newMaxChoices int,
) {
	if editor != p.Creator {
		panic("only the poll creator can edit this poll")
	}
	if time.Now().After(p.Deadline) {
		panic("poll is already closed (cannot edit closed poll)")
	}
	if newDeadline.Before(time.Now()) {
		panic("deadline must be in the future")
	}
	p.Title = newTitle
	p.Description = newDescription
	p.Deadline = newDeadline
	p.MaxChoices = newMaxChoices
}

//----------------------------------------
// Option Management

// AddOption appends a new option to the poll.
func (p *Poll) AddOption(optionText string) {
	if optionText == "" {
		panic("option text must not be empty")
	}
	p.Options = append(p.Options, optionText)
}

// EditOption modifies the text of an existing option by index.
func (p *Poll) EditOption(index int, newText string) {
	if index < 0 || index >= len(p.Options) {
		panic("invalid option index")
	}
	if newText == "" {
		panic("new option text must not be empty")
	}
	p.Options[index] = newText
}

// DeleteOption removes one option from the poll by index.
func (p *Poll) DeleteOption(index int) {
	if index < 0 || index >= len(p.Options) {
		panic("invalid option index")
	}
	p.Options = append(p.Options[:index], p.Options[index+1:]...)
}
