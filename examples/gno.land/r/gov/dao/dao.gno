package govdao

import (
	"std"

	"gno.land/p/gov/proposal"
)

var proposals = make([]Proposal, 0)

// XXX var members ...

// Proposal represents a proposal in the governance system.
type Proposal struct {
	author   std.Address
	idx      int
	Comment  string
	Executor proposal.Executor
	approved bool
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func Propose(proposal Proposal) int {
	// XXX: require payment?
	// XXX: sanitize proposal
	caller := std.PrevRealm().Addr()
	AssertIsMember(caller)
	proposal.author = caller
	proposal.idx = len(proposals)
	proposals = append(proposals, proposal)
	return proposal.idx
}

func VoteOnProposal(idx int, option string) {
	caller := std.PrevRealm().Addr()
	AssertIsMember(caller)
	assertProposalExists(idx)

	proposal := proposals[idx]
	// XXX: implement the voting (would be cool to have a generic p/)
	proposal.approved = option == "approved"
}

func ExecuteProposal(idx int) {
	assertProposalExists(idx)
	// XXX: assert voting is finished
	// XXX: assert voting result is YES
	// XXX: proposal was not already executed
	proposal := proposals[idx]
	proposal.Executor.Execute()
}

func assertProposalExists(idx int) {
	if idx < 0 || idx >= len(proposals) {
		panic("invalid proposal id")
	}
}

func IsMember(addr std.Address) bool {
	// XXX: implement
	return true // in the meantime, everyone is a DAO member
}

func AssertIsMember(addr std.Address) {
	if !IsMember(addr) {
		panic("caller is not member of govdao")
	}
}
