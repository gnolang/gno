// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_7_f_filetest

package z_7_f_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var parentID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	parentID = commondao.New(
		cross(cur),
		"Parent",
		"",
		commondao.AllowChildren(true),
		commondao.AllowDissolutionProposals(true),
	)

	// Dissolve the parent (sole council member passes the supermajority)
	pID := commondao.CreateDissolutionProposal(cross(cur), parentID)
	commondao.Vote(cross(cur), parentID, pID, pcommondao.ChoiceYes, "")
	commondao.Execute(cross(cur), parentID, pID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// A sub-DAO grafted onto a dissolved parent could never be dissolved
	// itself, so creation must be rejected.
	commondao.NewSubDAO(cross(cur), "Child", parentID, "")
}

// Error:
// DAO has already been dissolved
