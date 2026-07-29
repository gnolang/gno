// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_19_a_filetest

package z_19_a_filetest

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

	// Governance-disable the text kind
	pID := commondao.CreateSetProposalKindProposal(cross(cur), daoID, "text", false)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The text kind is deregistered while every other kind still works
	println("text enabled:", commondao.IsProposalKindEnabled(daoID, "text"))
	pID := commondao.CreateCouncilUpdateProposal(cross(cur), daoID, string(owner), "")
	println("council update created:", pID)

	// Creating a proposal of the disabled kind is rejected
	commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
}

// Output:
// text enabled: false
// council update created: 2

// Error:
// proposal kind not found
