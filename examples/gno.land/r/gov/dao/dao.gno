package govdao

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	pproposal "gno.land/p/gov/proposal"
)

var proposals = make([]*proposal, 0)
var members = make([]std.Address, 0)

type proposal struct {
	author   std.Address
	comment  string
	executor pproposal.Executor
	voter    Voter
}

func (p proposal) Status() Status {
	if p.executor.IsExpired() {
		return Expired
	}

	if p.executor.IsDone() {
		return Succeeded
	}

	if !p.voter.Finished(members) {
		return Active
	}

	if p.voter.Accepted(members) {
		return Accepted
	}

	return NotAccepted
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func Propose(comment string, executor pproposal.Executor) int {
	// XXX: require payment?
	if executor == nil {
		panic("missing proposal executor")
	}
	caller := std.PrevRealm().Addr()
	AssertIsMember(caller)

	prop := &proposal{
		comment:  comment,
		executor: executor,
		author:   caller,
		voter:    NewPercentageVoter(66), // at least 2/3 must say yes
	}

	proposals = append(proposals, prop)

	return len(proposals) - 1
}

func VoteOnProposal(idx int, option string) {
	assertProposalExists(idx)
	caller := std.PrevRealm().Addr()
	AssertIsMember(caller)

	prop := getProposal(idx)

	if prop.executor.IsExpired() {
		panic("prop is expired, cannot vote.")
	}

	if prop.voter.Finished(members) {
		panic("prop is not active anymore, cannot vote.")
	}

	prop.voter.Vote(members, option)
}

func ExecuteProposal(idx int) {
	assertProposalExists(idx)
	prop := getProposal(idx)

	if prop.executor.IsExpired() {
		panic("prop is expired, cannot execute.")
	}

	if !prop.voter.Finished(members) {
		panic("prop is still active, cannot execute.")
	}
	if !prop.voter.Accepted(members) {
		panic("prop is not accepted, cannot execute.")
	}

	prop.executor.Execute()
}

func IsMember(addr std.Address) bool {
	if len(members) == 0 { // special case for initial execution
		return true
	}

	for _, v := range members {
		if v == addr {
			return true
		}
	}

	return false
}

func AssertIsMember(addr std.Address) {
	if !IsMember(addr) {
		panic("caller is not member of govdao")
	}
}

func Render(path string) string {
	if path == "" {
		if len(proposals) == 0 {
			return "No proposals found :(." // corner case
		}

		output := ""
		for idx, prop := range proposals {
			output += ufmt.Sprintf("- [/r/gov/dao:%d](%d) - %s (**%s**)(by %s)\n", idx, idx, prop.comment, string(prop.Status()), prop.author)
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
	output += prop.comment
	output += "\n"
	output += ufmt.Sprintf("Status: %s", string(prop.Status()))
	output += "\n"
	output += ufmt.Sprintf("Author: %s", string(prop.author))
	return output
}

func getProposal(idx int) *proposal {
	if idx > len(proposals)-1 {
		panic("proposal not found")
	}

	return proposals[idx]
}

func proposalExists(idx int) bool {
	return idx >= 0 && idx <= len(proposals)
}

func assertProposalExists(idx int) {
	if !proposalExists(idx) {
		panic("invalid proposal id")
	}
}
