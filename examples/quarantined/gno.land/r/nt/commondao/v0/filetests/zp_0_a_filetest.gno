// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_0_a_filetest

package zp_0_a_filetest

import (
	"testing"


	"gno.land/r/nt/commondao/v0"
)

const (
	owner         = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user          = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newMembers    = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq\ng147ah9520z0r6jh9mjr6c75rv6l8aypzvcd3f7d"
	removeMembers = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", string(user))

}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateCouncilUpdateProposal(cross(cur), daoID, newMembers, removeMembers)

	p, found := commondao.GetView(daoID).GetProposal(pID)
	if !found {
		panic("expected proposal to be created")
	}

	println(string(p.Status()))
	println(p.Creator() == user)
	println(p.Title() == "Council Update")
	println("")
	println(p.Body())
}

// Output:
// active
// true
// true
//
// **Council Members to Add:**
// - g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
// - g147ah9520z0r6jh9mjr6c75rv6l8aypzvcd3f7d
//
//
// **Council Members to Remove:**
// - g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
