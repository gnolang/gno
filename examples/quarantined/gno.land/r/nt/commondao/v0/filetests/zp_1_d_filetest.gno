// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_1_d_filetest

package zp_1_d_filetest

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
	daoID = commondao.New(cross(cur), "Foo", string(user) + "\n" + string(owner))

	commondao.UpdateOptions(cross(cur), daoID, commondao.AllowTextProposals(true))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateTextProposal(cross(cur), daoID, "", "", 31)
}

// Error:
// maximum proposal voting period is 30 days
