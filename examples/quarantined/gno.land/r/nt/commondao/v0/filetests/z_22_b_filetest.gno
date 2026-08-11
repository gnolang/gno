// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_22_b_filetest

// Two bylaws amendments race from the same base: the first to execute
// wins, the second fails cleanly (StatusFailed) and clobbers nothing.
package z_22_b_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Single-member council: one YES decides any proposal immediately
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Both amendments diff against the same (absent) document.
	p1 := commondao.CreateAmendBylawsProposal(cross(cur), daoID, commondao.AmendBylawsPayload(daoID, "bylaws/a.md", "one"))
	p2 := commondao.CreateAmendBylawsProposal(cross(cur), daoID, commondao.AmendBylawsPayload(daoID, "bylaws/a.md", "two"))

	commondao.Vote(cross(cur), daoID, p1, pdao.ChoiceYes, "")
	commondao.Vote(cross(cur), daoID, p2, pdao.ChoiceYes, "")

	// The first to execute wins and re-pins the document hash.
	commondao.Execute(cross(cur), daoID, p1)
	println("doc after winner:", commondao.GetBylawsDoc(daoID, "bylaws/a.md"))

	// The second is now stale: it fails cleanly instead of clobbering.
	commondao.Execute(cross(cur), daoID, p2)
	prop, _ := commondao.GetView(daoID).GetProposal(p2)
	println("loser status:", string(prop.Status()))
	println("loser reason:", prop.StatusReason())
	println("doc after loser:", commondao.GetBylawsDoc(daoID, "bylaws/a.md"))
}

// Output:
// doc after winner: one
// loser status: failed
// loser reason: bylaws: document changed since the patch base
// doc after loser: one
