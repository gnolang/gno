// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_19_b_filetest

package z_19_b_filetest

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
	daoID uint64
	pID   uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

	// A text proposal is created while its kind is still registered
	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)

	// Governance-deregister the text kind after the proposal exists
	gID := commondao.CreateDeregisterKindProposal(cross(cur), daoID, "text")
	commondao.Vote(cross(cur), daoID, gID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, gID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The kind registry is read only at Propose, so the in-flight proposal
	// created before the kind was deregistered still votes and executes
	println("text registered:", commondao.HasProposalKind(daoID, "text"))

	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)

	p, _ := commondao.GetView(daoID).GetProposal(pID)
	println("status:", string(p.Status()))
}

// Output:
// text registered: false
// status: executed
