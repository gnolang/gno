package simpledao

import (
	"bytes"
	"errors"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/txlink"
)

var ErrMissingProposal = errors.New("proposal is missing")

// maxRequestProposals is the maximum number of
// paginated proposals that can be requested
const maxRequestProposals = 10

// proposal is the internal simpledao proposal implementation
type proposal struct {
	author      std.Address // initiator of the proposal
	title       string      // title of the proposal
	description string      // description of the proposal

	executor dao.Executor       // executor for the proposal
	status   dao.ProposalStatus // status of the proposal

	tally                 *tally        // voting tally
	getTotalVotingPowerFn func() uint64 // callback for the total voting power
}

func (p *proposal) Author() std.Address {
	return p.author
}

func (p *proposal) Title() string {
	return p.title
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

func (p *proposal) IsExpired() bool {
	return false // this proposal never expires
}

func (p *proposal) Render(idx int, authorUsername string) string {
	// Fetch the voting stats
	stats := p.Stats()

	var out bytes.Buffer

	out.WriteString(ufmt.Sprintf("# Proposal #%d - %s\n\n", idx, p.title))

	out.WriteString("## Description\n\n")
	if strings.TrimSpace(p.description) != "" {
		out.WriteString(ufmt.Sprintf("%s\n\n", p.description))
	} else {
		out.WriteString("No description provided.\n\n")
	}

	out.WriteString("## Proposal information\n\n")
	out.WriteString(ufmt.Sprintf("#### Status: %s\n\n", strings.ToUpper(p.Status().String())))

	out.WriteString(ufmt.Sprintf(
		"#### Voting stats:\n- YES %d (%d%%)\n- NO %d (%d%%)\n- ABSTAIN %d (%d%%)\n- MISSING VOTE %d (%d%%)\n",
		stats.YayVotes,
		stats.YayPercent(),
		stats.NayVotes,
		stats.NayPercent(),
		stats.AbstainVotes,
		stats.AbstainPercent(),
		stats.MissingVotes(),
		stats.MissingVotesPercent(),
	))
	out.WriteString("\n\n")
	thresholdOut := strings.ToUpper(ufmt.Sprintf("%t", stats.YayVotes > (2*stats.TotalVotingPower)/3))

	out.WriteString(ufmt.Sprintf("#### Threshold met: %s\n\n", thresholdOut))

	if authorUsername != "" {
		out.WriteString(ufmt.Sprintf("#### Author: [%s](/r/demo/users:%s)\n\n", authorUsername, authorUsername))
	} else {
		out.WriteString(ufmt.Sprintf("#### Author: %s\n\n", p.author.String()))
	}

	out.WriteString("## Actions\n\n")
	if p.status == dao.Active {
		out.WriteString(ufmt.Sprintf("[Vote YES](%s) - [Vote NO](%s) - [Vote ABSTAIN](%s)",
			txlink.URL("VoteOnProposal", "id", strconv.Itoa(idx), "option", "YES"),
			txlink.URL("VoteOnProposal", "id", strconv.Itoa(idx), "option", "NO"),
			txlink.URL("VoteOnProposal", "id", strconv.Itoa(idx), "option", "ABSTAIN"),
		))
	} else {
		out.WriteString("The voting period for this proposal is over.\n\n")
	}

	return out.String()
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
