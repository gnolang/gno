// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_propose_create_board_00_filetest

// Test the normal flow for ProposeCreateBoard and GovDAO approval.

package z_propose_create_board_00_filetest

import (
	"chain/runtime/unsafe"
	"testing"

	"gno.land/p/nt/testutils/v0"
	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
)

const boardName = "test123"

// member submits the proposal: only GovDAO members can. It holds no boards2
// role, so it could not call CreateBoard.
var (
	member = unsafe.OriginCaller()
	owner  = testutils.TestAddress("owner")
)

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), member)

	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	pr := boards2.ProposeCreateBoard(cross(cur), boardName, true, false, owner)
	dao.MustCreateProposal(cross(cur), pr)
}

func main(cur realm) {
	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	println("boards before =", boards2.BoardCount())

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(0)))
	println("executed =", dao.ExecuteProposal(cross(cur), dao.ProposalID(0)))

	bid, found := boards2.GetBoardIDFromName(cross(cur), boardName)
	println("found =", found)
	println("ID =", bid)
	println("owner is owner =", boards2.HasMemberRole(bid, owner, "owner"))
	println("proposer is owner =", boards2.HasMemberRole(bid, member, "owner"))
}

// Output:
// boards before = 0
// executed = true
// found = true
// ID = 1
// owner is owner = true
// proposer is owner = false
