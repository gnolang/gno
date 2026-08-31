// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_10_e_filetest

package z_10_e_filetest

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

var (
	daoID uint64
	pID   uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	// A 3-member council keeps the proposal active after one vote, so the
	// voter can change their choice.
	members := "g1cyh9qm4y43w38gfk77qxczrhqj2asdr8l7uk54\ng1x95kxqpwvqjf98qpkvsay7tppkh6zs9fr7wwda"
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", members)
	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 30)

	// Vote YES, then change to NO: the vote stats must show only NO,
	// never a stale "YES | 0.00%" row from the zeroed counter.
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceNo, "")
}

func main() {
	println(commondao.Render(strconv.FormatUint(daoID, 10) + "/proposals/" + strconv.FormatUint(pID, 10)))
}

// Output:
//
// > Voting ends on **Sun, 15 Mar 2009 11:31pm UTC**
//
// # #1 Title
// [Go to DAO](/r/nt/commondao/v0:2) • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=1&reason=&vote=)
//
// ---
// ## Details
// - Proposer: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// - Submit Time: Fri, 13 Feb 2009 23:31:30 UTC
// - Expected Outcome: **pending** ⏳
// - Status: **active**
// ## Description
// Body
//
// ## Stats
// | Vote Choices | Percentage of Votes |
// | --- | --- |
// | NO | 100.00% |
//
//
// ## Votes
// Total number of votes: **1**
// | Users | Votes |
// | --- | --- |
// | g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 | [NO](/r/nt/commondao/v0:2/proposals/1/vote/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5) |
