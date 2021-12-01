package groups

import "std"

//----------------------------------------
// VoteSet

type VoteSet interface {
	// number of present votes in set.
	Size() int
	// add or update vote for voter.
	SetVote(voter std.Address, value string) error
	// count the number of votes for value.
	CountVotes(value string) int
}

//----------------------------------------
// VoteList

type Vote struct {
	Voter std.Address
	Value string
}

type VoteList []Vote

func NewVoteList() *VoteList {
	return &VoteList{}
}

func (vlist *VoteList) Size() int {
	return len(*vlist)
}

func (vlist *VoteList) SetVote(voter std.Address, value string) error {
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
	MinSigners int
	Addresses  std.AddressSet
}

//----------------------------------------
// CommitteeSession

type CommitteeSession struct {
	Name      string
	Committee Committee
}
