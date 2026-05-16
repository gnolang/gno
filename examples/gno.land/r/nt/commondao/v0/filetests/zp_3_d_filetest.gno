// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_3_d_filetest

package zp_3_d_filetest

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
	commondao.Invite(cross2(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	daoID = commondao.New(cross2(cur), "Foo").ID()

	options := commondao.GetOptions(daoID)
	options.SetAllowDissolutionProposals(false)
}

func main(cur realm) {
	commondao.CreateDissolutionProposal(cross2(cur), daoID)
}

// Error:
// forbidden
