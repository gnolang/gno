// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_d_filetest

package z_20_d_filetest

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

	// Single-member council: one YES decides any proposal immediately.
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Deregister-by-name round trip through the permanent manage-kinds kind:
	// the text kind is seeded, so a supermajority deregister removes it and
	// HasProposalKind flips to false. Deregistering blocks only new proposals
	// of the kind; the registry is read at Propose time.
	println("text registered before:", commondao.HasProposalKind(daoID, "text"))

	pID := commondao.CreateDeregisterKindProposal(cross(cur), daoID, "text")
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	println("text registered after:", commondao.HasProposalKind(daoID, "text"))

	// A catalog kind deregistered by name can be re-registered by name: the
	// manage-kinds kind is never bricked, so the kind set stays governable.
	pID = commondao.CreateRegisterKindProposal(cross(cur), daoID, "text")
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	println("text re-registered:", commondao.HasProposalKind(daoID, "text"))
}

// Output:
// text registered before: true
// text registered after: false
// text re-registered: true
