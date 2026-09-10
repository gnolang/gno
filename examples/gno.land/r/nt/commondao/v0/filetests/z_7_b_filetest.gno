// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_7_b_filetest

package z_7_b_filetest

// Entry points must authenticate the *caller*, not the transaction origin.
// The council here is a realm, reached by seating it through New's members
// argument and resigning; the origin driving it is an address that is not on
// the council. Every assertion below fails if an entry point reads
// unsafe.OriginCaller() instead of cur.Previous().Address().
//
// z_7_a used to carry this coverage implicitly (its DAO was created by a code
// realm while the origin stayed a user). New now rejects realm callers, so the
// caller != origin case needs a DAO that a realm reached the supported way.

import (
	"chain"
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner  = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user   = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	origin = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
	policy = "gno.land/r/demo/policy"
)

var (
	daoID   uint64
	council address
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// The invited user seats the policy realm next to itself, then resigns:
	// the realm is left as the DAO's sole council member.
	council = chain.PackageAddress(policy)
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", string(council))
	commondao.Resign(cross(cur), daoID)
}

func main(cur realm) {
	// origin is NOT on the council; the policy realm is.
	testing.SetRealm(testing.NewUserRealm(origin))
	testing.SetRealm(testing.NewCodeRealm(policy))

	println(commondao.GetView(daoID).Council().Has(council))
	println(commondao.GetView(daoID).Council().Has(origin))

	// Create, vote and execute as the realm. Each of these authorizes the
	// caller against the council, and the realm is a one-member council, so
	// its single YES decides.
	pID := commondao.CreateSubDAOProposal(cross(cur), daoID, "Bar", "Purpose", "", string(council))
	commondao.Vote(cross(cur), daoID, pID, pcommondao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)

	p, found := commondao.GetView(daoID).GetProposal(pID)
	println(found && p.Creator() == council)
	println(string(p.Status()))

	subID := uint64(3)
	println(commondao.GetView(subID).Name())

	// A text proposal on the same DAO, also as the realm.
	tID := commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 1)
	t, found := commondao.GetView(daoID).GetProposal(tID)
	println(found && t.Creator() == council)
}

// Output:
// true
// false
// true
// executed
// Bar
// true
