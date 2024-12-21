package foodao

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	pproposal "gno.land/p/gov/proposal"
)

var proposals = make([]*proposal, 0)

// XXX var members ...

type proposal struct {
	idx      int
	author   std.Address
	comment  string
	executor pproposal.Executor
	// XXX: make "accepted" and "finished" managed by an interface that can have various voting implementations
	accepted bool
	finished bool
}

func (p proposal) Status() Status {
	if p.executor.Done() {
		return Status(p.executor.Status())
	}
	if p.accepted {
		return Accepted
	}
	// XXX: timeout
	// XXX: not_accepted
	return Active
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
		idx:      len(proposals),
	}

	proposals = append(proposals, prop)
	return prop.idx
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
	prop.executor.Execute()
}

func IsMember(addr std.Address) bool {
	// XXX: implement
	return true // in the meantime, everyone is a DAO member
}

func AssertIsMember(addr std.Address) {
	if !IsMember(addr) {
		panic("caller is not member of foodao")
	}
}

func Render(path string) string {
	if path == "" {
		output := ""
		for idx, prop := range proposals {
			output += ufmt.Sprintf("- [/r/foo/dao:%d](%d) - %s (by %s)", idx, idx, prop.comment, prop.author)
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
	output += ufmt.Sprintf("# Prop#%d", prop.idx) + "\n"
	output += "\n"
	output += prop.comment
	output += "\n"
	output += ufmt.Sprintf("Status: %s", string(prop.Status()))
	output += "\n"
	output += ufmt.Sprintf("Author: %s", string(prop.author))
	return output
}

func getProposal(idx int) *proposal {
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
