// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_10_c_filetest

package z_10_c_filetest

import (
	"strconv"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var proposalPath string

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID := commondao.New(cross(cur), "Foo", "Purpose", "", "")

	// The SubDAO name is user controlled; the proposal body is pre-built
	// markdown chrome that must render formatted, with the name escaped.
	pID := commondao.CreateSubDAOProposal(cross(cur), daoID, "*Evil* [x](https://evil.example)", "Purpose", "", string(owner))
	proposalPath = strconv.FormatUint(daoID, 10) + "/proposals/" + strconv.FormatUint(pID, 10)
}

func main() {
	println(commondao.Render(proposalPath))
}

// Output:
//
// > Voting ends on **Fri, 20 Feb 2009 11:31pm UTC**
//
// # #1 New SubDAO: \*Evil\* \[x\]\(https://evil\.example\)
// [Go to DAO](/r/nt/commondao/v0:2) • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=1&reason=&vote=)
//
// ---
// ## Details
// - Proposer: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// - Submit Time: Fri, 13 Feb 2009 23:31:30 UTC
// - Expected Outcome: **pending** ⏳
// - Status: **active**
// ## Description
// **Parent DAO:**
// [Foo](/r/nt/commondao/v0:2)
//
// **SubDAO Name:**
// \*Evil\* \[x\]\(https://evil\.example\)
//
// **SubDAO Purpose:**
// Purpose
//
// **Council Members:**
// - g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
