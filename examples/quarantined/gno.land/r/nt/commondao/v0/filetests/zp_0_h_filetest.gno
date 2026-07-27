// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_0_h_filetest

package zp_0_h_filetest

import (
	"testing"

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

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	daoID = commondao.New(cross(cur), "Foo", "", string(user))


}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateCouncilUpdateProposal(cross(cur), daoID, user.String(), "")

	// Adding an existing member is an idempotent no-op, not an error
	p, found := commondao.GetView(daoID).GetProposal(pID)
	println(found)
	println(string(p.Status()))
}

// Output:
// true
// active
