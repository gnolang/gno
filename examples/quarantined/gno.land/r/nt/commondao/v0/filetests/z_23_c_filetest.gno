// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_23_c_filetest

// The single-vote page's failure branches.
//
// Each is a navigable page (header + links back), not a bare dead-end
// string, and a missing vote distinguishes a council member who has not
// voted yet — an ordinary state — from an account outside the electorate.
package z_23_c_filetest

import (
	"strconv"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner    = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user     = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	silent   = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // council member, never votes
	outsider = address("g1qjnm0dpkcze233rljaqffny3eel3wt6wn4lql2") // not on the council
)

var proposalPath string

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID := commondao.New(cross(cur), "Foo", "Purpose", "", silent.String())

	pID := commondao.CreateTextProposal(cross(cur), daoID, "Decision", "Body", 7)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "yes from the proposer")
	proposalPath = strconv.FormatUint(daoID, 10) + "/proposals/" + strconv.FormatUint(pID, 10)
}

func main() {
	// A recorded vote.
	println(commondao.Render(proposalPath + "/vote/" + string(user)))

	// A council member who has not voted yet.
	println(commondao.Render(proposalPath + "/vote/" + string(silent)))

	// An account that was never in the electorate.
	println(commondao.Render(proposalPath + "/vote/" + string(outsider)))

	// A malformed address.
	println(commondao.Render(proposalPath + "/vote/not-an-address"))
}

// Output:
// # Vote: Proposal #1
// [Go to DAO](/r/nt/commondao/v0:2) • [Go to Proposal](/r/nt/commondao/v0:2/proposals/1)
//
// ---
// ## Details
// - User: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// - Vote: YES
// ## Reason
// yes from the proposer
//
//
// # Vote: Proposal #1
// [Go to DAO](/r/nt/commondao/v0:2) • [Go to Proposal](/r/nt/commondao/v0:2/proposals/1)
//
// ---
// This council member has not voted on this proposal yet.
//
//
// # Vote: Proposal #1
// [Go to DAO](/r/nt/commondao/v0:2) • [Go to Proposal](/r/nt/commondao/v0:2/proposals/1)
//
// ---
// This account is not a member of the proposal's electorate.
//
//
// # Vote: Proposal #1
// [Go to DAO](/r/nt/commondao/v0:2) • [Go to Proposal](/r/nt/commondao/v0:2/proposals/1)
//
// ---
// Invalid address.
