// This package is used to manage the validator set.
package validators

type Change struct {
	ValAddr     std.Address
	VotingPower uint
}

type Proposal struct {
	author    std.Address
	Comment   string
	ChangesFN func() []Change // requires a callback defined in another code realm (cannot be set from maketx call or maketx run).
}

var proposals = make([]Proposal, 0)

// ProposeChanges is designed to be called by another contract, not by a maketx call.
func ProposeChanges(proposal Proposal) int {
	// XXX: require payment?
	// XXX: assert caller is auhtorized
	// XXX: sanitize proposal
	proposal.author = std.PrevRealm().Addr()
	idx := len(proposals)
	proposals = append(proposals, proposal)
	return idx
}

func VoteOnProposal(idx int, option string) {
	// XXX: assert caller
	// XXX: implement the voting (woudl be nice to have a dedicated p/)
}

func ExecuteProposal(idx int) {
	// no need for sanitization.
	assertProposalExists(idx)
	proposal := proposals[idx]
	changes := proposal.ChangesFN()
	std.Emit(changes)
}

func assertProposalExists(idx int) {
	if idx < 0 || idx >= len(proposals) {
		panic("invalid proposal id")
	}
}
