// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_22_d_filetest

// Bylaws rendering is escape-safe and line-preserving: hostile markdown
// in document text and in patch insert literals is neutralized on both
// the proposal page (the change summary) and the bylaws page (the stored
// document), and multi-line content keeps its line structure. A literal
// line starting with "- " cannot masquerade as a deletion marker: every
// literal line is marker-prefixed by the summary.
package z_22_d_filetest

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

	hostile := "# Heading injection\n\n<div>html block</div>\n\n[link](https://attacker.example)\n\nplain closing line"

	// Create the document; render the ACTIVE proposal page first: the
	// change summary shows the hostile content escaped, line by line.
	pID := commondao.CreateAmendBylawsProposal(cross(cur), daoID, commondao.AmendBylawsPayload(daoID, "bylaws/policy.md", hostile))
	println(commondao.Render(strconv.FormatUint(daoID, 10) + "/proposals/" + strconv.FormatUint(pID, 10)))

	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)

	// Amend with an insert whose lines mimic summary markers.
	pID = commondao.CreateAmendBylawsProposal(cross(cur), daoID, commondao.AmendBylawsPayload(daoID, "bylaws/policy.md", hostile+"\n- spoofed deletion marker\n= 9 unchanged"))
	println(commondao.Render(strconv.FormatUint(daoID, 10) + "/proposals/" + strconv.FormatUint(pID, 10)))

	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)

	// The stored multi-line document renders escaped with its line
	// structure preserved.
	println(commondao.Render(strconv.FormatUint(daoID, 10) + "/bylaws"))
}

// Output:
//
// > Voting ends on **Fri, 20 Feb 2009 11:31pm UTC**
//
// # #1 Add Bylaws Document: bylaws/policy\.md
// [Go to DAO](/r/nt/commondao/v0:2) • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=1&reason=&vote=)
//
// ---
// ## Details
// - Proposer: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// - Submit Time: Fri, 13 Feb 2009 23:31:30 UTC
// - Expected Outcome: **pending** ⏳
// - Status: **active**
// ## Description
// **Document:** bylaws/policy\.md
//
// ```
// + # Heading injection
// +
// + <div>html block</div>
// +
// + [link](https://attacker.example)
// +
// + plain closing line
//
// ```
//
//
//
//
// > Voting ends on **Fri, 20 Feb 2009 11:31pm UTC**
//
// # #2 Amend Bylaws Document: bylaws/policy\.md
// [Go to DAO](/r/nt/commondao/v0:2) • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=2&reason=&vote=)
//
// ---
// ## Details
// - Proposer: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// - Submit Time: Fri, 13 Feb 2009 23:31:30 UTC
// - Expected Outcome: **pending** ⏳
// - Status: **active**
// ## Description
// **Document:** bylaws/policy\.md
//
// ```
// = 96 unchanged
// +
// + - spoofed deletion marker
// + = 9 unchanged
//
// ```
//
//
//
// # Foo: Bylaws & Mandates
// [Go to DAO](/r/nt/commondao/v0:2)
//
// ---
// ## bylaws/policy\.md
// sha256: 9ce0a63e0a9255e7a01911bcd9ae19db2880c7a3af153fe7a3363a332bdf1ea6
//
//
//
// \# Heading injection
//
// <div>html block</div>
//
// [link](https://attacker.example)
//
// plain closing line
// \- spoofed deletion marker
// = 9 unchanged
