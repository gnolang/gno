package govdao

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gov/proposal"
)

var proposals = make([]*Proposal, 0)

// XXX var members ...

// Proposal represents a proposal in the governance system.
// XXX: make it an interface with various proposal implementations
type Proposal struct {
	author   std.Address
	idx      int
	Comment  string
	Executor proposal.Executor
	accepted bool
	finished bool
}

func (p Proposal) Status() string {
	if p.Executor.Done() {
		return p.Executor.Status()
	}
	if p.accepted {
		return "accepted"
	}
	// XXX: timeout
	// XXX: not_accepted
	return "active"
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func Propose(proposal *Proposal) int {
	// XXX: require payment?
	// XXX: sanitize proposal or make dedicated constructor which is not the proposal itself
	caller := std.PrevRealm().Addr()
	AssertIsMember(caller)
	proposal.author = caller
	proposal.idx = len(proposals)
	proposals = append(proposals, proposal)
	return proposal.idx
}

func VoteOnProposal(idx int, option string) {
	assertProposalExists(idx)
	caller := std.PrevRealm().Addr()
	AssertIsMember(caller)

	prop := getProposal(idx)
	if prop.finished {
		panic("prop is not active anymore, cannot vote.")
	}
	// XXX: implement the real voting (would be cool to have a generic p/)
	prop.accepted = option == "YES"
	prop.finished = true
}

func ExecuteProposal(idx int) {
	assertProposalExists(idx)
	prop := getProposal(idx)
	if !prop.finished {
		panic("prop is still active, cannot execute.")
	}
	if !prop.accepted {
		panic("prop is not accepted, cannot execute.")
	}
	prop.Executor.Execute()
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

func Render(path string) string {
	if path == "" {
		output := ""
		for idx, prop := range proposals {
			output += ufmt.Sprintf("- [/r/gov/dao:%d](%d) - %s (by %s)", idx, idx, prop.Comment, prop.author)
		}
		return output
	}

	// else display the proposal
	idx, err := strconv.Atoi(path)
	if err != nil {
		return "404"
	}

	if !proposalExists(idx) {
		return "404"
	}
	prop := getProposal(idx)
	output := ""
	output += ufmt.Sprintf("# Prop#%d", idx) + "\n"
	output += "\n"
	output += prop.Comment
	output += "\n"
	output += ufmt.Sprintf("Status: %s", prop.Status())
	return output
}

func getProposal(idx int) *Proposal {
	return proposals[idx-1]
}

func proposalExists(idx int) bool {
	return idx > 0 && idx <= len(proposals)
}

func assertProposalExists(idx int) {
	if !proposalExists(idx) {
		panic("invalid proposal id")
	}
}
