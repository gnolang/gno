package simpledao

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/gov/dao"
	pproposal "gno.land/p/gov/proposal"
)

var errMissingProposal = errors.New("proposal is missing")

// proposal is the internal simpledao proposal implementation
type proposal struct {
	author      std.Address // initiator of the proposal
	description string      // description of the proposal

	executor pproposal.Executor // executor for the proposal
	status   dao.Status         // status of the proposal

	votes *votes // voting mechanism
}

func (p *proposal) GetAuthor() std.Address {
	return p.author
}

func (p *proposal) GetDescription() string {
	return p.description
}

func (p *proposal) GetStatus() dao.Status {
	return p.status
}

func (p *proposal) GetVotes() []dao.Vote {
	return p.votes.getVotes()
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

func (p *PropStore) GetProposals(givenOffset, count uint64) []dao.Proposal {
	// Calculate the left and right bounds
	offset := givenOffset
	if offset < 1 {
		offset = 1
	}

	var (
		startIndex = (offset - 1) * count
		endIndex   = startIndex + count

		numProposals = uint64(p.proposals.Size())
	)

	// Check if the current offset has any proposals
	if startIndex >= numProposals {
		return []dao.Proposal{}
	}

	// Check if the right bound is good
	if endIndex >= numProposals {
		endIndex = numProposals - 1
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

func (p *PropStore) GetProposalsByAddress(address std.Address) []dao.Proposal {
	props := make([]dao.Proposal, 0, p.proposals.Size())

	p.proposals.Iterate("", "", func(_ string, val interface{}) bool {
		prop := val.(*proposal)

		// Check the author
		if prop.GetAuthor() != address {
			props = append(props, prop)
		}

		return false
	})

	return props
}

// getProposalID generates a sequential proposal ID
// from the given ID number
func getProposalID(id uint64) string {
	return seqid.ID(id).String()
}
