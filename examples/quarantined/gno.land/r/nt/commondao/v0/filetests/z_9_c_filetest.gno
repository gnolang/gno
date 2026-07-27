// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_9_c_filetest

package z_9_c_filetest

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

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", string(owner))
	commondao.UpdateOptions(cross(cur), daoID, commondao.AllowDissolutionProposals(true))

	// Dissolve the DAO through a proposal
	pID := commondao.CreateDissolutionProposal(cross(cur), daoID, owner)
	commondao.Vote(cross(cur), daoID, pID, pcommondao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Vote(cross(cur), daoID, pID, pcommondao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Resign(cross(cur), daoID)
}

// Error:
// DAO is deleted
