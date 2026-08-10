package daodemo

// This is the most basic example of a DAO using DAOKIT.
// It is a simple DAO that has a single admin role and a single public-relationships role.
// It is used to demonstrate the basic functionality of DAOKIT.

import (
	"gno.land/p/samcrew/basedao"
	"gno.land/p/samcrew/daocond"
	"gno.land/r/demo/profile"
)

// DAO is the canonical public surface — a concrete *basedao.DAOWrapper,
// NOT the daokit.DAO interface. Holding the concrete wrapper prevents
// any /r/ from down-casting to the address-typed interface (which would
// re-introduce a Class-2 designation-forgery surface). All mutating
// methods on *DAOWrapper take (_ int, rlm realm, ...) and validate
// IsCurrent before deriving the caller.
var DAO *basedao.DAOWrapper

// TODO: add a little bit more complexity to show daocond initialization
func init(cur realm) {
	initialRoles := []basedao.RoleInfo{
		{Name: "admin", Description: "Admin is the superuser"},
		{Name: "public-relationships", Description: "Responsible of communication with the public"},
		{Name: "finance-officer", Description: "Responsible of funds management"},
	}

	initialMembers := []basedao.Member{
		{Address: "g126gx6p6d3da4ymef35ury6874j6kys044r7zlg", Roles: []string{"admin", "public-relationships"}},
		{Address: "g1ld6uaykyugld4rnm63rcy7vju4zx23lufml3jv", Roles: []string{"public-relationships"}},
		{Address: "g1r69l0vhp7tqle3a0rk8m8fulr8sjvj4h7n0tth", Roles: []string{"finance-officer"}},
		{Address: "g16jv3rpz7mkt0gqulxas56se2js7v5vmc6n6e0r", Roles: []string{}},
	}

	// create the member store now to be able to use it in the condition
	memberStore := basedao.NewMembersStore(initialRoles, initialMembers)

	membersMajority := daocond.MembersThreshold(0.6, memberStore.IsMember, memberStore.MembersCount)
	publicRelationships := daocond.RoleCount(1, "public-relationships", memberStore.HasRole)
	financeOfficer := daocond.RoleCount(1, "finance-officer", memberStore.HasRole)

	// and & or use va_args so you can pass as many conditions as you want
	adminCond := daocond.And(membersMajority, publicRelationships, financeOfficer)

	DAO = basedao.New(&basedao.Config{
		Name:             "Demo DAOKIT DAO",
		Description:      "This is a demo DAO built with [DAOKIT](/p/samcrew/daokit)",
		Members:          memberStore,
		InitialCondition: adminCond,
		GetProfileString: profile.GetStringField,
		SetProfileString: profile.SetStringField,
	}, cur)
}

func Vote(cur realm, proposalID uint64, vote daocond.Vote) {
	DAO.Vote(0, cur, proposalID, vote)
}

func Execute(cur realm, proposalID uint64) {
	DAO.Execute(0, cur, proposalID)
}

func Render(path string) string {
	return DAO.Render(path)
}
