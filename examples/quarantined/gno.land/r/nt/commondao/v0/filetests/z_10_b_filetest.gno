// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_10_b_filetest

package z_10_b_filetest

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
	daoID        uint64
	proposalPath string
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "", "")

	pID := commondao.CreateTextProposal(cross(cur), daoID, "*Title* [t](https://evil.example)", "**Body** with <img src=x> and [b](https://evil.example)", 0)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "*reason* [r](https://evil.example)")
	proposalPath = strconv.FormatUint(daoID, 10) + "/proposals/" + strconv.FormatUint(pID, 10)
}

func main() {
	// Proposal titles, bodies and vote reasons are user controlled:
	// markdown metacharacters must render escaped, never as markup
	println(commondao.Render(proposalPath))
	println(commondao.Render(proposalPath + "/vote/" + string(user)))
}

// Output:
// # #1 \*Title\* \[t\]\(https://evil\.example\)
// [Go to DAO](/r/nt/commondao/v0:2) • [Execute](/r/nt/commondao/v0$help&func=Execute&daoID=2&proposalID=1)
//
// ---
// ## Details
// - Proposer: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// - Submit Time: Fri, 13 Feb 2009 23:31:30 UTC
// - Status: **passed**
// ## Description
// \*\*Body\*\* with \<img src=x\> and \[b\]\(https://evil\.example\)
//
// ## Stats
// | Vote Choices | Percentage of Votes |
// | --- | --- |
// | YES | 100.00% |
//
//
// ## Votes
// Total number of votes: **1**
// | Users | Votes |
// | --- | --- |
// | g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 | [YES](/r/nt/commondao/v0:2/proposals/1/vote/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5) with a reason |
//
//
//
// # Vote: Proposal #1
// [Go to DAO](/r/nt/commondao/v0:2) • [Go to Proposal](/r/nt/commondao/v0:2/proposals/1)
//
// ---
// ## Details
// - User: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// - Vote: YES
// ## Reason
// \*reason\* \[r\]\(https://evil\.example\)
