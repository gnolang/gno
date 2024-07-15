package govdao

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/simpledao"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gov/dao"
)

var (
	d         dao.DAO         // the current active DAO implementation
	proposals dao.PropStore   // the proposal store
	members   dao.MemberStore // the member store
)

func init() {
	// Example initial member set (just test addresses)
	set := []dao.Member{
		{
			Address:     std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"),
			VotingPower: 10,
		},
	}

	var (
		propStore = simpledao.NewPropStore()
		membStore = simpledao.NewMembStore(simpledao.WithInitialMembers(set))
	)

	// Set the stores
	proposals = propStore
	members = membStore

	// Set the DAO implementation
	d = simpledao.New(membStore, propStore)
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func Propose(comment string, executor dao.Executor) uint64 {
	idx, err := d.Propose(comment, executor)
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
	return proposals
}

// GetMembStore returns the active member store
func GetMembStore() dao.MemberStore {
	return members
}

func Render(path string) string {
	if path == "" {
		numProposals := proposals.Size()

		if numProposals == 0 {
			return "No proposals found :(" // corner case
		}

		output := ""

		offset := uint64(0)
		if numProposals >= 10 {
			offset = uint64(numProposals) - 10
		}

		// Fetch the last 10 proposals
		for idx, prop := range proposals.GetProposals(offset, uint64(10)) {
			output += ufmt.Sprintf(
				"- [Proposal #%d](%s:%d) - (**%s**)(by %s)\n",
				idx,
				strings.TrimPrefix(getGOVDAOPath(), "gno.land"),
				idx,
				prop.GetStatus().String(),
				prop.GetAuthor().String(),
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
	prop, err := proposals.GetProposalByID(uint64(idx))
	if err != nil {
		return ufmt.Sprintf("unable to fetch proposal, %s", err.Error())
	}

	// Fetch the voting stats
	stats := prop.GetVotingStats()

	output := ""
	output += ufmt.Sprintf("# Prop #%d", idx)
	output += "\n\n"
	output += ufmt.Sprintf("Author: %s", prop.GetAuthor().String())
	output += "\n\n"
	output += prop.GetDescription()
	output += "\n\n"
	output += ufmt.Sprintf("Status: %s", prop.GetStatus().String())
	output += "\n\n"
	output += ufmt.Sprintf(
		"Voting stats: YAY %d (%d%%), NAY %d (%d%%), ABSTAIN %d (%d%%), HAVEN'T VOTED %d (%d%%)",
		stats.YayVotes,
		stats.GetYayPercent(),
		stats.NayVotes,
		stats.GetNayPercent(),
		stats.AbstainVotes,
		stats.GetAbstainPercent(),
		stats.MissingVotes,
		stats.GetMissingPercent(),
	)
	output += "\n\n"
	output += ufmt.Sprintf("Threshold met: %t", stats.ThresholdMet)
	output += "\n\n"

	return output
}

// getGOVDAOPath returns the active govdao Realm path
func getGOVDAOPath() string {
	// TODO use r/sys/vars
	return "gno.land/r/gov/dao"
}
