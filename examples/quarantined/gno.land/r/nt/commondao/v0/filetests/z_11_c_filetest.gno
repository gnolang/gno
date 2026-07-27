// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_11_c_filetest

package z_11_c_filetest

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
	daoID = commondao.New(cross(cur), "Foo", "", "")
}

func main(cur realm) {
	// Call as a user which is not a council member
	testing.SetRealm(testing.NewUserRealm(owner))

	commondao.CreateTreasurySpendProposal(cross(cur), daoID, owner, "ugnot", 1)
}

// Error:
// caller is not a council member
