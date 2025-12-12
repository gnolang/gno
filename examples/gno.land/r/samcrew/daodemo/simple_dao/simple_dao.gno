package simple_dao

// This is the most basic example of a DAO using DAOKIT.
// It is a simple DAO that has a single admin role, a public-relationships role and a finance-officer.
// It is used to demonstrate the basic functionality of DAOKIT.

import (
	"gno.land/p/samcrew/basedao"
	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
	"gno.land/r/demo/profile"
)

var (
	DAO        daokit.DAO          // Crossing interface for DAO interaction
	localDAO   daokit.DAO          // Local interface for maketx call
	daoPrivate *basedao.DAOPrivate // Full access to internal DAO state
)

// Initializes the DAO with predefined roles, members, and governance rules.
func init() {
	// Define initial roles in the DAO
	initialRoles := []basedao.RoleInfo{
		{Name: "public-relationships", Description: "Responsible of communication with the public", Color: "#21577A"},
		{Name: "finance-officer", Description: "Responsible of funds management", Color: "#F3D3BC"},
	}

	// Define initial members and their roles
	initialMembers := []basedao.Member{
		{Address: "g1demo1234567890abcdefghijklmnopqrstuvwxyz", Roles: []string{"finance-officer"}},
	}

	// Create the member store now to be able to use it in the condition
	memberStore := basedao.NewMembersStore(initialRoles, initialMembers)

	// Define governance conditions using daocond
	membersMajority := daocond.MembersThreshold(0.4, memberStore.IsMember, memberStore.MembersCount)
	financeOfficer := daocond.RoleCount(1, "finance-officer", memberStore.HasRole)

	// `and` and `or` use va_args so you can pass as many conditions as needed
	adminCond := daocond.And(membersMajority, financeOfficer)

	// Initialize DAO with configuration
	localDAO, daoPrivate = basedao.New(&basedao.Config{
		Name:             "Demo DAOKIT DAO",
		Description:      "This is a demo DAO built with DAOKIT",
		Members:          memberStore,
		InitialCondition: adminCond,
		GetProfileString: profile.GetStringField,
		SetProfileString: profile.SetStringField,
		PrivateVarName:   "daoPrivate",
		SetImplemFn:      setImplem,
		RenderFn: func(path string, dao *basedao.DAOPrivate) string {
			s := ""
			// Demo path to showcase the possibility of this realm
			if path == "" {
				s += renderDemo()
			}
			s += " \n\n--- \n\n "
			s += dao.RenderRouter.Render(path)
			return s
		},
		CrossFn: crossFn,
	})
	DAO = daokit.NewCrossing(localDAO, crossFn)

	initProposals(25)
}

// Creates a new proposal
// To execute this function, you must use a MsgRun `gnokey maketx run“
// See why it is necessary in Gno Documentation: https://docs.gno.land/users/interact-with-gnokey#run
func Propose(cur realm, req daokit.ProposalRequest) {
	localDAO.Propose(req)
}

// Allows DAO members to cast their vote on a specific proposal
func Vote(cur realm, proposalID uint64, vote daocond.Vote) {
	localDAO.Vote(proposalID, vote)
}

// Triggers the implementation of a proposal's actions
func Execute(cur realm, proposalID uint64) {
	localDAO.Execute(proposalID)
}

// Generates a UI representation of the DAO's state
func Render(path string) string {
	return localDAO.Render(path)
}

// Enables cross-realm calls for multi-realm DAO operations.
// It receives a callback function and executes it in the current realm context.
func crossFn(_ realm, cb func()) {
	cb()
}

// Updates the DAO implementation when governance proposals change it.
// This allows the DAO to upgrade itself through proposals that modify its core logic.
func setImplem(newLocalDAO daokit.DAO, newDAO daokit.DAO) {
	localDAO, DAO = newLocalDAO, newDAO
}
