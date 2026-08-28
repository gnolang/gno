// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_22_a_filetest

// Bylaws lifecycle through governance: create a document with a diff
// patch, amend it, list the set, render the bylaws page, remove it.
package z_22_a_filetest

import (
	"strconv"
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

	// Create a document through governance: the payload is a create patch.
	payload := commondao.AmendBylawsPayload(daoID, "bylaws/quorum.md", "Quorum is half the council.")
	pID := commondao.CreateAmendBylawsProposal(cross(cur), daoID, payload)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	println("created:", commondao.GetBylawsDoc(daoID, "bylaws/quorum.md"))

	// Amend it with a diff patch against the current text.
	payload = commondao.AmendBylawsPayload(daoID, "bylaws/quorum.md", "Quorum is two thirds of the council.")
	pID = commondao.CreateAmendBylawsProposal(cross(cur), daoID, payload)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	println("amended:", commondao.GetBylawsDoc(daoID, "bylaws/quorum.md"))

	// Add a second document, then list the whole set (sorted).
	payload = commondao.AmendBylawsPayload(daoID, "bylaws/spend.md", "Spend only by vote.")
	pID = commondao.CreateAmendBylawsProposal(cross(cur), daoID, payload)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	for _, path := range commondao.ListBylawsDocs(daoID, "") {
		println("doc:", path)
	}

	// The bylaws page renders every document with its base hash.
	println(commondao.Render(strconv.FormatUint(daoID, 10) + "/bylaws"))

	// Remove it: an empty proposed text is a remove patch.
	payload = commondao.AmendBylawsPayload(daoID, "bylaws/spend.md", "")
	pID = commondao.CreateAmendBylawsProposal(cross(cur), daoID, payload)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	println("removed:", commondao.GetBylawsDoc(daoID, "bylaws/spend.md") == "")
}

// Output:
// created: Quorum is half the council.
// amended: Quorum is two thirds of the council.
// doc: bylaws/quorum.md
// doc: bylaws/spend.md
// # Foo: Bylaws & Mandates
// [Go to DAO](/r/nt/commondao/v0:2)
//
// ---
// ## bylaws/quorum\.md
// sha256: 20fd0dbce30ca94d0c6210a877e7b7c50c81c9b66e60265a5a9b6125805db799
//
//
//
// Quorum is two thirds of the council.
//
// ## bylaws/spend\.md
// sha256: c890333668c0f8104cf5a6607139f61a962f51077c5e2c8cb3933930ff5e9dc5
//
//
//
// Spend only by vote.
//
//
// removed: true
