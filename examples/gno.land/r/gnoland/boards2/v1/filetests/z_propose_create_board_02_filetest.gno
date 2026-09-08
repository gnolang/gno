// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_propose_create_board_02_filetest

// The failure for ProposeCreateBoard and GovDAO approval where the
// board name is taken between proposal creation and execution.

package z_propose_create_board_02_filetest

import (
	"chain/runtime/unsafe"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	boardName         = "test123"
)

var member address = unsafe.OriginCaller()

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), member)

	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	pr := boards2.ProposeCreateBoard(cross(cur), boardName, true, false, member)
	dao.MustCreateProposal(cross(cur), pr)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross(cur), boardName, false, false)

	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(0)))
	println("executed =", dao.ExecuteOrRejectProposal(cross(cur), dao.ProposalID(0)))
	println("boards =", boards2.BoardCount())
}

// Output:
// executed = false
// boards = 1
