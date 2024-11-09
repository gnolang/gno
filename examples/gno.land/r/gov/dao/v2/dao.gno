package govdao

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
	"gno.land/p/demo/simpledao"
	"gno.land/p/demo/ufmt"

	"gno.land/r/demo/users"
)

var (
	d       *simpledao.SimpleDAO  // the current active DAO implementation
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
		output := "# GovDAO Proposals\n\n"
		numProposals := d.Size()

		if numProposals == 0 {
			output += "No proposals found :(" // corner case
			return output
		}

		offset := uint64(0)
		if numProposals >= 10 {
			offset = uint64(numProposals) - 10
		}

		// Fetch the last 10 proposals
		proposals := d.Proposals(offset, uint64(10))
		for i := len(proposals) - 1; i >= 0; i-- {
			prop := proposals[i]

			title := prop.Title()
			if len(title) > 40 {
				title = title[:40] + "..."
			}

			propID := offset + uint64(i)
			output += ufmt.Sprintf("## [Prop #%d - %s](/r/gov/dao/v2:%d)\n\n", propID, title, propID)
			output += ufmt.Sprintf("#### Status: %s\n\n", strings.ToUpper(prop.Status().String()))

			user := users.GetUserByAddress(prop.Author())
			authorUsername := prop.Author().String()
			if user != nil {
				authorUsername = ufmt.Sprintf("[%s](/r/demo/users:%s)", user.Name, user.Name)
			}

			output += ufmt.Sprintf("#### Author: %s\n\n", authorUsername)

			if i != 0 {
				output += "---\n\n"
			}
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

	authorUsername := ""
	user := users.GetUserByAddress(prop.Author())
	if user != nil {
		authorUsername = user.Name
	}

	// Render the proposal
	output := prop.Render(idx, authorUsername)
	output += "\n\n"

	return output
}
