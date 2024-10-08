package govdao

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	pproposal "gno.land/p/gov/proposal"
)

var (
	proposals = make([]*proposal, 0)
	members   = make([]std.Address, 0) // XXX: these should be pointers to avoid data duplication. Not possible due to VM bugs
)

const (
	msgMissingExecutor = "missing proposal executor"
	msgPropExecuted    = "prop already executed"
	msgPropExpired     = "prop is expired"
	msgPropInactive    = "prop is not active anymore"
	msgPropActive      = "prop is still active"
	msgPropNotAccepted = "prop is not accepted"

	msgCallerNotAMember = "caller is not member of govdao"
	msgProposalNotFound = "proposal not found"
)

type proposal struct {
	author      std.Address
	title       string
	description string
	executor    pproposal.Executor
	voter       Voter
	executed    bool
	voters      []std.Address // XXX: these should be pointers to avoid data duplication. Not possible due to VM bugs.
}

func (p proposal) Status() Status {
	if p.executor.IsExpired() {
		return Expired
	}

	if p.executor.IsDone() {
		return Succeeded
	}

	if !p.voter.IsFinished(members) {
		return Active
	}

	if p.voter.IsAccepted(members) {
		return Accepted
	}

	return Rejected
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func Propose(title, comment string, executor pproposal.Executor) int {
	// XXX: require payment?
	if executor == nil {
		panic(msgMissingExecutor)
	}
	caller := std.GetOrigCaller() // XXX: CHANGE THIS WHEN MSGRUN PERSIST CODE ESCAPING THE main() SCOPE! IT IS UNSAFE!
	AssertIsMember(caller)

	prop := &proposal{
		title:       title,
		description: comment,
		executor:    executor,
		author:      caller,
		voter:       NewPercentageVoter(66), // at least 2/3 must say yes
	}

	proposals = append(proposals, prop)

	return len(proposals) - 1
}

func VoteOnProposal(idx int, option string) {
	assertProposalExists(idx)
	caller := std.GetOrigCaller() // XXX: CHANGE THIS WHEN MSGRUN PERSIST CODE ESCAPING THE main() SCOPE! IT IS UNSAFE!
	AssertIsMember(caller)

	prop := getProposal(idx)

	if prop.executed {
		panic(msgPropExecuted)
	}

	if prop.executor.IsExpired() {
		panic(msgPropExpired)
	}

	if prop.voter.IsFinished(members) {
		panic(msgPropInactive)
	}

	prop.voter.Vote(members, caller, option)
}

func ExecuteProposal(idx int) {
	assertProposalExists(idx)
	prop := getProposal(idx)

	if prop.executed {
		panic(msgPropExecuted)
	}

	if prop.executor.IsExpired() {
		panic(msgPropExpired)
	}

	if !prop.voter.IsFinished(members) {
		panic(msgPropActive)
	}

	if !prop.voter.IsAccepted(members) {
		panic(msgPropNotAccepted)
	}

	prop.executor.Execute()
	prop.voters = members
	prop.executed = true
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
		panic(msgCallerNotAMember)
	}
}

func Render(path string) string {
	output := ""

	if path == "" { // render home
		output = "# GovDAO Proposals\n\n"
		if len(proposals) == 0 {
			output += "No proposals."
			return output
		}

		for idx, prop := range proposals {
			output += prop.RenderlListItem(idx)

			if len(proposals) > 1 {
				output += "---\n\n"
			}
		}

		return output
	}

	// render individual proposals
	idx, err := strconv.Atoi(path)
	if err != nil {
		return "404"
	}

	if !proposalExists(idx) {
		return "404"
	}

	prop := getProposal(idx)

	vs := members
	if prop.executed {
		vs = prop.voters
	}

	output += ufmt.Sprintf("# Prop #%d - %s\n\n", idx, prop.title)
	output += ufmt.Sprintf("## Description\n\n%s\n\n", prop.description)

	output += "## Proposal information\n\n"
	output += ufmt.Sprintf("#### Status: %s\n\n", string(prop.Status()))
	output += ufmt.Sprintf("#### Votes: \n\n%s\n\n", prop.voter.Status(vs))
	output += ufmt.Sprintf("#### Author: %s", string(prop.author))

	return output
}

func getProposal(idx int) *proposal {
	if idx > len(proposals)-1 {
		panic(msgProposalNotFound)
	}

	return proposals[idx]
}

func (p proposal) RenderlListItem(idx int) string {
	output := ufmt.Sprintf("## [Prop #%d - %s](/r/gov/dao:%d)\n\n", idx, p.title, idx)
	output += ufmt.Sprintf("#### Status: %s\n\n", string(p.Status()))
	return output
}

func proposalExists(idx int) bool {
	return idx >= 0 && idx <= len(proposals)
}

func assertProposalExists(idx int) {
	if !proposalExists(idx) {
		panic("invalid proposal id")
	}
}
