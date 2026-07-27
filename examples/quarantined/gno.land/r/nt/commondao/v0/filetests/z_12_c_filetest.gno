// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_c_filetest

package z_12_c_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
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
	rootID := commondao.New(cross(cur), "Root", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "A", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	subA = 3
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "B", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	subB = 4
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Siblings have no authority over each other
	commondao.CreateTreasuryClawbackProposal(cross(cur), subA, subB)
}

// Error:
// DAO is not an ancestor of the target DAO
