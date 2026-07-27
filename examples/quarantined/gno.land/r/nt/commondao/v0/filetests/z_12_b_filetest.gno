// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_b_filetest

package z_12_b_filetest

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
	testing.SetRealm(testing.NewUserRealm(user))

	// A DAO can never claw back its own treasury
	commondao.CreateTreasuryClawbackProposal(cross(cur), daoID, daoID)
}

// Error:
// a DAO cannot target its own treasury
