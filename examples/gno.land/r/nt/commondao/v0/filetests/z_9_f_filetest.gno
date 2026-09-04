// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_9_f_filetest

package z_9_f_filetest

import (
	"strings"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

// The Execute-deadline re-tally uses the proposal's electorate SNAPSHOT, not
// the live council. A 3-member proposal with a single YES is undecided
// (supermajority needs 3*1 >= 2*3). If two members resign after the proposal
// is created, the live council shrinks to 1 (where 1 YES would pass), but the
// snapshot electorate stays 3, so at the deadline the proposal is DISMISSED.
// A mutant tallying the live council at Execute would instead pass it.
const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	moul  = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
)

var (
	daoID uint64
	pID   uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	members := strings.Join([]string{owner.String(), moul.String()}, "\n")
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", members)

	// Snapshot electorate = 3 members; a single YES leaves it undecided.
	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")

	// Two members resign: the live council shrinks to {user}, but the
	// proposal's electorate snapshot is unchanged.
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Resign(cross(cur), daoID)
	testing.SetRealm(testing.NewUserRealm(moul))
	commondao.Resign(cross(cur), daoID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SkipHeights(121000) // pass the 7 day deadline

	commondao.Execute(cross(cur), daoID, pID)

	p, _ := commondao.GetView(daoID).GetProposal(pID)
	println("status:", string(p.Status()))
	println("live council:", commondao.GetView(daoID).Council().Size())
	println("electorate:", p.Electorate().Size())
}

// Output:
// status: dismissed
// live council: 1
// electorate: 3
