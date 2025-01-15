package govdao

import (
	"std"

	"gno.land/r/gov/dao/bridge"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
	"gno.land/p/demo/simpledao"
)

var (
	d       *simpledao.SimpleDAO  // the current active DAO implementation
	members membstore.MemberStore // the member store

	// GovDAO exposes all functions of this contract as methods
	GovDAO = &DAO{}
)

// DAO is an empty struct that allows all
// functions of this realm to be methods instead of functions
// This allows a registry, such as r/gov/dao/bridge
// to take this object and match it to a required interface
type DAO struct{}

const daoPkgPath = "gno.land/r/gov/dao/v2"

func init() {
	// Example initial member set (just test addresses)
	set := []membstore.Member{
		{
			Address:     std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"),
			VotingPower: 10,
		},
	}

	// Set the member store
	members = membstore.NewMembStore(membstore.WithInitialMembers(set), membstore.WithDAOPkgPath(daoPkgPath))

	// Set the DAO implementation
	d = simpledao.New(members)

	bridge.RegisterNewDAOImpl(GovDAO)
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func (_ DAO) Propose(request dao.ProposalRequest) uint64 {
	idx, err := d.Propose(request)
	if err != nil {
		panic(err)
	}

	return idx
}

// VoteOnProposal casts a vote for the given proposal
func (_ DAO) VoteOnProposal(id uint64, option dao.VoteOption) {
	if err := d.VoteOnProposal(id, option); err != nil {
		panic(err)
	}
}

// ExecuteProposal executes the proposal
func (_ DAO) ExecuteProposal(id uint64) {
	if err := d.ExecuteProposal(id); err != nil {
		panic(err)
	}
}

// GetPropStore returns the active proposal store
func (_ DAO) GetPropStore() dao.PropStore {
	return d
}

// GetMembStore returns the active member store
func (_ DAO) GetMembStore() membstore.MemberStore {
	return members
}
