// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_1_a_filetest

package zp_1_a_filetest

import (
	"testing"


	"gno.land/r/nt/commondao/v0"
)

const (
	owner      = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user       = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	title      = "General Proposal"
	body       = "Foo Bar"
	votingDays = uint8(1)
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", string(user) + "\n" + string(owner))

}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateTextProposal(cross(cur), daoID, title, body, votingDays)

	p, found := commondao.GetView(daoID).GetProposal(pID)
	if !found {
		panic("expected proposal to be created")
	}

	println(string(p.Status()))
	println(p.Creator() == user)
	println(p.Title() == title)
	println(p.VotingDeadline())
	println(p.Body())
}

// Output:
// active
// true
// true
// 2009-02-14 23:31:30 +0000 UTC
// Foo Bar
