package govdao

import (
	"std"
	"strconv"

	"gno.land/p/demo/membstore"
	"gno.land/p/demo/simpledao"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gov/dao"
)

var (
	d       dao.DAO               // the current active DAO implementation
	members membstore.MemberStore // the member store
)

func init() {
	var (
		// Example initial member set (just test addresses)
		set = []membstore.Member{
			{
				Address:     std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"),
				VotingPower: 10,
			},
		}
	)

	// Set the member store
	members = membstore.NewMembStore(membstore.WithInitialMembers(set))

	// Set the DAO implementation
	d = simpledao.New(members)
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func Propose(request dao.ProposalRequest) uint64 {
	idx, err := d.Propose(request)
	if err != nil {
		panic(err)
	}

	return idx
}

// VoteOnProposal casts a vote for the given proposal
func VoteOnProposal(id uint64, option dao.VoteOption) {
	if err := d.VoteOnProposal(id, option); err != nil {
		panic(err)
	}
}

// ExecuteProposal executes the proposal
func ExecuteProposal(id uint64) {
	if err := d.ExecuteProposal(id); err != nil {
		panic(err)
	}
}

// GetPropStore returns the active proposal store
func GetPropStore() dao.PropStore {
	return d
}

// GetMembStore returns the active member store
func GetMembStore() membstore.MemberStore {
	return members
}

func Render(path string) string {
	if path == "" {
		numProposals := d.Size()

		if numProposals == 0 {
			return "No proposals found :(" // corner case
		}

		output := ""

		offset := uint64(0)
		if numProposals >= 10 {
			offset = uint64(numProposals) - 10
		}

		// Fetch the last 10 proposals
		for idx, prop := range d.Proposals(offset, uint64(10)) {
			output += ufmt.Sprintf(
				"- [Proposal #%d](%s:%d) - (**%s**)(by %s)\n",
				idx,
				"/r/gov/dao/v2",
				idx,
				prop.Status().String(),
				prop.Author().String(),
			)
		}

		return output
	}

	// Display the detailed proposal
	idx, err := strconv.Atoi(path)
	if err != nil {
		return "404: Invalid proposal ID"
	}

	// Fetch the proposal
	prop, err := d.ProposalByID(uint64(idx))
	if err != nil {
		return ufmt.Sprintf("unable to fetch proposal, %s", err.Error())
	}

	// Render the proposal
	output := ""
	output += ufmt.Sprintf("# Prop #%d", idx)
	output += "\n\n"
	output += prop.Render()
	output += "\n\n"

	return output
}
