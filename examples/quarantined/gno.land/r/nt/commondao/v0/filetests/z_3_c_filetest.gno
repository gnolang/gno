// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_3_c_filetest

package z_3_c_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var subID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	daoID := commondao.New(cross(cur), "Foo", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), daoID, "Bar", "", string(user))
		commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), daoID, pID)
	}
	subID = 3
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))

	commondao.TransferOwnership(cross(cur), subID, user)
}

// Error:
// ownership belongs to the root DAO, transfer using the root DAO ID
