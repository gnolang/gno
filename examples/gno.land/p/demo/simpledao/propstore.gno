package simpledao

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/gov/dao"
	"gno.land/p/gov/executor"
)

var (
	errMissingProposal = errors.New("proposal is missing")
	errMissingVote     = errors.New("member has not voted")
)

const (
	// maxRequestProposals is the maximum number of
	// paginated proposals that can be requested
	maxRequestProposals = 10

	// maxRequestMembers is the maximum number of
	// paginated members that can be requested
	maxRequestMembers = 50
)

// proposal is the internal simpledao proposal implementation
type proposal struct {
	author      std.Address // initiator of the proposal
	description string      // description of the proposal

	executor executor.Executor  // executor for the proposal
	status   dao.ProposalStatus // status of the proposal

	votes *votes // voting mechanism
}

func (p *proposal) GetAuthor() std.Address {
	return p.author
}

func (p *proposal) GetDescription() string {
	return p.description
}

func (p *proposal) GetStatus() dao.ProposalStatus {
	return p.status
}

func (p *proposal) GetVotes() []dao.Vote {
	var (
		voters = p.votes.getVoters()
		votes  = make([]dao.Vote, 0, voters.Size())
	)

	voters.Iterate("", "", func(key string, val interface{}) bool {
		option := val.(dao.VoteOption)

		vote := dao.Vote{
			Address: std.Address(key),
			Option:  option,
		}

		votes = append(votes, vote)

		return false
	})

	return votes
}

func (p *proposal) GetVoteByMember(address std.Address) (dao.Vote, error) {
	voters := p.votes.getVoters()

	optionRaw, exists := voters.Get(address.String())
	if !exists {
		return dao.Vote{}, errMissingVote
	}

	vote := dao.Vote{
		Address: address,
		Option:  optionRaw.(dao.VoteOption),
	}

	return vote, nil
}

// PropStore implements the dao.PropStore abstraction
type PropStore struct {
	proposals *avl.Tree // seqid.ID -> proposal
}

// NewPropStore returns a new simpledao PropStore
func NewPropStore() *PropStore {
	return &PropStore{
		proposals: avl.NewTree(),
	}
}

// addProposal adds a new simpledao proposal to the store
func (p *PropStore) addProposal(proposal *proposal) (uint64, error) {
	// See what the next proposal number should be
	nextID := uint64(p.proposals.Size())

	// Save the proposal
	p.proposals.Set(getProposalID(nextID), proposal)

	return nextID, nil
}

func (p *PropStore) GetProposals(offset, count uint64) []dao.Proposal {
	// Calculate the left and right bounds
	if count < 1 {
		return []dao.Proposal{}
	}

	// Limit the maximum number of returned proposals
	if count > maxRequestProposals {
		count = maxRequestProposals
	}

	var (
		startIndex = offset
		endIndex   = startIndex + count

		numProposals = uint64(p.proposals.Size())
	)

	// Check if the current offset has any proposals
	if startIndex >= numProposals {
		return []dao.Proposal{}
	}

	// Check if the right bound is good
	if endIndex > numProposals {
		endIndex = numProposals
	}

	props := make([]dao.Proposal, 0)
	p.proposals.Iterate(
		getProposalID(startIndex),
		getProposalID(endIndex),
		func(_ string, val interface{}) bool {
			prop := val.(*proposal)

			// Save the proposal
			props = append(props, prop)

			return false
		},
	)

	return props
}

func (p *PropStore) GetProposalByID(id uint64) (dao.Proposal, error) {
	prop, exists := p.proposals.Get(getProposalID(id))
	if !exists {
		return nil, errMissingProposal
	}

	return prop.(*proposal), nil
}

func (p *PropStore) Size() int {
	return p.proposals.Size()
}

// getProposalID generates a sequential proposal ID
// from the given ID number
func getProposalID(id uint64) string {
	return seqid.ID(id).String()
}
