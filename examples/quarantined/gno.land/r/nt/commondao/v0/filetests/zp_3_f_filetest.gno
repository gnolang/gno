// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_3_f_filetest

package zp_3_f_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

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

	// Create a DAO owned by user, with user as the only council member
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "")
	commondao.UpdateOptions(cross(cur), daoID, commondao.AllowDissolutionProposals(true))

	// Dissolve the DAO through a proposal decided by supermajority
	pID := commondao.CreateDissolutionProposal(cross(cur), daoID)
	commondao.Vote(cross(cur), daoID, pID, pcommondao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// A dissolved DAO rejects new proposals
	commondao.CreateDissolutionProposal(cross(cur), daoID)
}

// Error:
// DAO is deleted
