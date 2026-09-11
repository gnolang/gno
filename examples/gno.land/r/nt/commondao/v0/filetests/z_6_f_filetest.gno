// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_6_f_filetest

package z_6_f_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoID uint64
	pID   uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Council is {user, owner}
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", string(owner))

	// One day voting period; a 1Y/1N tie leaves the proposal undecided
	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 1)
	commondao.Vote(cross(cur), daoID, pID, pcommondao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Vote(cross(cur), daoID, pID, pcommondao.ChoiceNo, "")
}

func main(cur realm) {
	// Advance past the voting deadline (5 seconds per height)
	testing.SkipHeights(17281)

	// Finalization after the deadline is permissionless: the caller is
	// not a council member, and undecided proposals are dismissed
	commondao.Execute(cross(cur), daoID, pID)

	p, _ := commondao.GetView(daoID).GetProposal(pID)
	println(string(p.Status()))
}

// Output:
// dismissed
