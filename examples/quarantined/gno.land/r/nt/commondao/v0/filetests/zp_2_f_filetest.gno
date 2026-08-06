// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_2_f_filetest

package zp_2_f_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "Foo"
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", string(user))

	{
		pID := commondao.CreateSubDAOProposal(cross(cur), daoID, name, "Purpose", "", string(user))
		commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
		testing.SetRealm(testing.NewUserRealm(user))
		commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
		commondao.Execute(cross(cur), daoID, pID)
	}
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateSubDAOProposal(cross(cur), daoID, name, "Purpose", "", string(user))
}

// Error:
// a SubDAO with the same name already exists
