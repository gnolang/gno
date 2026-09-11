// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_9_e_filetest

package z_9_e_filetest

import (
	"testing"

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

	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Withdraw(cross(cur), daoID, pID)

	p, found := commondao.GetView(daoID).GetProposal(pID)
	println(found)
	println(string(p.Status()))
	println(commondao.GetView(daoID).ActiveProposalsSize())
}

// Output:
// true
// withdrawn
// 0
