package govdao

import (
	"strconv"

	"github.com/gnolang/gno/examples/gno.land/p/demo/simpledao"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gov/dao"
	pproposal "gno.land/p/gov/proposal"
)

var (
	d         dao.DAO       // the current active DAO implementation
	proposals dao.PropStore // the proposal store
	members   dao.MembStore // the member store
)

func init() {
	var (
		propStore = simpledao.NewPropStore()
		membStore = simpledao.NewMembStore()
	)

	// Set the stores
	proposals = propStore
	members = membStore

	// Set the DAO implementation
	d = simpledao.NewSimpleDAO(membStore, propStore)
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func Propose(comment string, executor pproposal.Executor) (uint64, error) {
	return d.Propose(comment, executor)
}

func VoteOnProposal(id uint64, option VoteOption) error {
	return d.VoteOnProposal(id, option)
}

func ExecuteProposal(id uint64) error {
	return d.ExecuteProposal(id)
}

func GetPropStore() dao.PropStore {
	return proposals
}

func GetMembStore() dao.MembStore {
	return members
}

func Render(path string) string {
	if path == "" {
		if len(proposals) == 0 {
			return "No proposals found :(" // corner case
		}

		output := ""
		for idx, prop := range proposals {
			output += ufmt.Sprintf("- [%d](/r/gov/dao:%d) - %s (**%s**)(by %s)\n", idx, idx, prop.comment, string(prop.Status()), prop.author)
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

	vs := members
	if prop.executed {
		vs = prop.voters
	}

	output := ""
	output += ufmt.Sprintf("# Prop #%d", idx)
	output += "\n\n"
	output += prop.comment
	output += "\n\n"
	output += ufmt.Sprintf("Status: %s", string(prop.Status()))
	output += "\n\n"
	output += ufmt.Sprintf("Voting status: %s", prop.voter.Status(vs))
	output += "\n\n"
	output += ufmt.Sprintf("Author: %s", string(prop.author))
	output += "\n\n"

	return output
}
