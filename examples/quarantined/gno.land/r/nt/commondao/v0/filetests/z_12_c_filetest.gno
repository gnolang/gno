// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_c_filetest

package z_12_c_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	subA uint64
	subB uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	rootID := commondao.New(cross(cur), "Root", "", commondao.AllowChildren(true))
	subA = commondao.NewSubDAO(cross(cur), "A", rootID, "", commondao.AllowTreasuryProposals(true))
	subB = commondao.NewSubDAO(cross(cur), "B", rootID, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Siblings have no authority over each other
	commondao.CreateTreasuryClawbackProposal(cross(cur), subA, subB)
}

// Error:
// DAO is not an ancestor of the target DAO
