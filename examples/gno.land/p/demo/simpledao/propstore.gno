package simpledao

import (
	"errors"
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

var ErrMissingProposal = errors.New("proposal is missing")

const (
	// maxRequestProposals is the maximum number of
	// paginated proposals that can be requested
	maxRequestProposals = 10

	// maxRequestVotes is the maximum number of
	// paginated votes that can be requested
	maxRequestVotes = 50
)

// proposal is the internal simpledao proposal implementation
type proposal struct {
	author      std.Address // initiator of the proposal
	description string      // description of the proposal

	executor dao.Executor       // executor for the proposal
	status   dao.ProposalStatus // status of the proposal

	tally                 *tally        // voting tally
	getTotalVotingPowerFn func() uint64 // callback for the total voting power
}

func (p *proposal) Author() std.Address {
	return p.author
}

func (p *proposal) Description() string {
	return p.description
}

func (p *proposal) Status() dao.ProposalStatus {
	return p.status
}

func (p *proposal) Executor() dao.Executor {
	return p.executor
}

func (p *proposal) Stats() dao.Stats {
	// Get the total voting power of the body
	totalPower := p.getTotalVotingPowerFn()

	return dao.Stats{
		YayVotes:         p.tally.yays,
		NayVotes:         p.tally.nays,
		AbstainVotes:     p.tally.abstains,
		TotalVotingPower: totalPower,
	}
}

func (p *proposal) Render() string {
	// Fetch the voting stats
	stats := p.Stats()

	output := ""
	output += ufmt.Sprintf("Author: %s", p.Author().String())
	output += "\n\n"
	output += p.Description()
	output += "\n\n"
	output += ufmt.Sprintf("Status: %s", p.Status().String())
	output += "\n\n"
	output += ufmt.Sprintf(
		"Voting stats: YAY %d (%d%%), NAY %d (%d%%), ABSTAIN %d (%d%%), HAVEN'T VOTED %d (%d%%)",
		stats.YayVotes,
		stats.YayPercent(),
		stats.NayVotes,
		stats.NayPercent(),
		stats.AbstainVotes,
		stats.AbstainPercent(),
		stats.MissingVotes(),
		stats.MissingVotesPercent(),
	)
	output += "\n\n"
	output += ufmt.Sprintf("Threshold met: %t", stats.YayVotes > (2*stats.TotalVotingPower)/3)

	return output
}

// addProposal adds a new simpledao proposal to the store
func (s *SimpleDAO) addProposal(proposal *proposal) (uint64, error) {
	// See what the next proposal number should be
	nextID := uint64(s.proposals.Size())

	// Save the proposal
	s.proposals.Set(getProposalID(nextID), proposal)

	return nextID, nil
}

func (s *SimpleDAO) Proposals(offset, count uint64) []dao.Proposal {
	// Check the requested count
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

		numProposals = uint64(s.proposals.Size())
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
	s.proposals.Iterate(
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

func (s *SimpleDAO) ProposalByID(id uint64) (dao.Proposal, error) {
	prop, exists := s.proposals.Get(getProposalID(id))
	if !exists {
		return nil, ErrMissingProposal
	}

	return prop.(*proposal), nil
}

func (s *SimpleDAO) Size() int {
	return s.proposals.Size()
}

// getProposalID generates a sequential proposal ID
// from the given ID number
func getProposalID(id uint64) string {
	return seqid.ID(id).String()
}
