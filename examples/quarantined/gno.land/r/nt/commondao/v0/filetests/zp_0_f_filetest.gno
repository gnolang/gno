// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_0_f_filetest

package zp_0_f_filetest

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
	dao := commondao.New(cross(cur), "Foo")
	subDAO := commondao.NewSubDAO(cross(cur), "Bar", dao.ID())
	subDAO.Members().Add(user)

	daoID = dao.ID()

	options := commondao.GetOptions(cross(cur), daoID)
	options.SetAllowMembersUpdate(false) // Disallow udpating members
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateMembersUpdateProposal(cross(cur), daoID, "", "")
}

// Error:
// forbidden
