package govdao

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
	"gno.land/p/demo/simpledao"
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
