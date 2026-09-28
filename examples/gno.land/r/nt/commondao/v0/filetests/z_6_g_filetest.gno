// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_6_g_filetest

package z_6_g_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx") // govdao t1 multisig
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoID uint64
	pID   uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

	// A passed proposal that is not executed before its deadline
	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
}

func main(cur realm) {
	// Past the deadline, finalization is permissionless: a non-council
	// caller can execute the passed proposal.
	testing.SetRealm(testing.NewUserRealm(owner))
	testing.SkipHeights(121000) // 5s blocks: pass the 7 day voting deadline

	commondao.Execute(cross(cur), daoID, pID)

	p, _ := commondao.GetView(daoID).GetProposal(pID)
	println("status:", string(p.Status()))
}

// Output:
// status: executed
