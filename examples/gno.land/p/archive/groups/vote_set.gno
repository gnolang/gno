package groups

import (
	"time"

	"gno.land/p/archive/rat"
)

//----------------------------------------
// VoteSet

type VoteSet interface {
	// number of present votes in set.
	Size() int
	// add or update vote for voter.
	SetVote(voter address, value string) error
	// count the number of votes for value.
	CountVotes(value string) int
}

//----------------------------------------
// VoteList

type Vote struct {
	Voter address
	Value string
}

type VoteList []Vote

func NewVoteList() *VoteList {
	return &VoteList{}
}

func (vlist *VoteList) Size() int {
	return len(*vlist)
}

func (vlist *VoteList) SetVote(voter address, value string) error {
	// TODO optimize with binary algorithm
	for i, vote := range *vlist {
		if vote.Voter == voter {
			// update vote
			(*vlist)[i] = Vote{
				Voter: voter,
				Value: value,
			}
			return nil
		}
	}
	*vlist = append(*vlist, Vote{
		Voter: voter,
		Value: value,
	})
	return nil
}

func (vlist *VoteList) CountVotes(target string) int {
	// TODO optimize with binary algorithm
	var count int
	for _, vote := range *vlist {
		if vote.Value == target {
			count++
		}
	}
	return count
}

//----------------------------------------
// Committee

type Committee struct {
	Quorum    rat.Rat
	Threshold rat.Rat
	Addresses []address
}

//----------------------------------------
// VoteSession
// NOTE: this seems a bit too formal and
// complicated vs what might be possible;
// something simpler, more informal.

type SessionStatus int

const (
	SessionNew SessionStatus = iota
	SessionStarted
	SessionCompleted
	SessionCanceled
)

type VoteSession struct {
	Name      string
	Creator   address
	Body      string
	Start     time.Time
	Deadline  time.Time
	Status    SessionStatus
	Committee *Committee
	Votes     VoteSet
	Choices   []string
	Result    string
}
