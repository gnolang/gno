// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_22_c_filetest

// A bylaws amendment payload built against superseded document text is
// rejected at proposal creation (stale base), so a proposal can never be
// created against text that no longer exists.
package z_22_c_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"

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

	// Single-member council: one YES decides any proposal immediately
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// A payload built against the current (absent) document...
	stale := commondao.AmendBylawsPayload(daoID, "bylaws/a.md", "one")

	// ...is superseded when another amendment lands first.
	pID := commondao.CreateAmendBylawsProposal(cross(cur), daoID, commondao.AmendBylawsPayload(daoID, "bylaws/a.md", "two"))
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)

	// Proposing the superseded payload is rejected at creation.
	commondao.CreateAmendBylawsProposal(cross(cur), daoID, stale)
}

// Error:
// bylaws: document changed since the patch base
