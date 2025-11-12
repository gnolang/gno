package custom_resource

import (
	"gno.land/p/samcrew/basedao"
	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
	"gno.land/r/demo/profile"
)

var (
	DAO        daokit.DAO
	localDAO   daokit.DAO
	daoPrivate *basedao.DAOPrivate
	myblog     Blog
)

func init() {
	initialRoles := []basedao.RoleInfo{
		{Name: "Role", Description: "Has a role", Color: "#329175"},
	}
	initialMembers := []basedao.Member{
		{Address: "g1demo1234567890abcdefghijklmnopqrstuvw1", Roles: []string{}},
		{Address: "g1demo1234567890abcdefghijklmnopqrstuvw2", Roles: []string{}},
		{Address: "g1demo1234567890abcdefghijklmnopqrstuvw3", Roles: []string{}},
		{Address: "g1demo1234567890abcdefghijklmnopqrstuvw4", Roles: []string{}},
	}
	memberStore := basedao.NewMembersStore(initialRoles, initialMembers)
	membersMajority := daocond.MembersThreshold(0.1, memberStore.IsMember, memberStore.MembersCount)

	localDAO, daoPrivate = basedao.New(&basedao.Config{
		Name:             "Demo DAOKIT DAO",
		Description:      "This is a demo DAO built with DAOKIT",
		Members:          memberStore,
		InitialCondition: membersMajority,
		GetProfileString: profile.GetStringField,
		SetProfileString: profile.SetStringField,
		PrivateVarName:   "daoPrivate",
		RenderFn: func(path string, dao *basedao.DAOPrivate) string {
			s := ""
			if path == "" {
				s += renderDemo()
			}
			s += " \n\n--- \n\n "
			s += dao.RenderRouter.Render(path)
			s += "# My blog\n"
			s += myblog.Render()
			return s
		},
		SetImplemFn: setImplem,
		CrossFn:     crossFn,
	})
	DAO = daokit.NewCrossing(localDAO, crossFn)

	// Create a new custom resource
	resource := daokit.Resource{
		Handler:     NewPostHandler(&myblog),
		Condition:   membersMajority,
		DisplayName: "Create Blog Post",
		Description: "Allows DAO members to create and publish new blog posts",
	}
	// Add it to our DAO
	daoPrivate.Core.Resources.Set(&resource)

	// Add some demo blog post proposals to showcase custom resource
	initBlogProposals()
}

// Propose creates a new proposal
// To execute this function, you must use a MsgRun (maketx run)
// See why it is necessary in Gno Documentation: https://docs.gno.land/users/interact-with-gnokey#run
func Propose(cur realm, req daokit.ProposalRequest) {
	localDAO.Propose(req)
}

// Vote allows DAO members to cast their vote on a specific proposal
func Vote(cur realm, proposalID uint64, vote daocond.Vote) {
	localDAO.Vote(proposalID, vote)
}

// Execute triggers the implementation of a proposal's actions
func Execute(cur realm, proposalID uint64) {
	localDAO.Execute(proposalID)
}

// Render generates a UI representation of the DAO's state
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
