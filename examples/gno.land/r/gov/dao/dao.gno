package govdao

import "std"

var proposals = make([]Proposal, 0)

// XXX var members ...

// Proposal represents a proposal in the governance system.
type Proposal struct {
	author   std.Address
	idx      int
	Comment  string
	Callback func()
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
	AssertIsMember(caller)
	panic("not implemented")
	// XXX: implement the voting (woudl be cool to have a generic p/)
}

func ExecuteProposal(idx int) {
	assertProposalExists(idx)
	// XXX: assert voting is finished
	// XXX: assert voting result is YES
	// XXX: proposal was not already executed
	proposal := proposals[idx]
	proposal.Callback()
}

func assertProposalExists(idx int) {
	if idx < 0 || idx >= len(proposals) {
		panic("invalid proposal id")
	}
}

func IsMember(addr std.Address) bool {
	panic("not implemented")
	// XXX: implement
	return true
}

func AssertIsMember(addr std.Address) {
	if !IsMember(addr) {
		panic("caller is not member of govdao")
	}
}
