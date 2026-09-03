// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_0_c_filetest

// Test that reads are open to any caller
package z_hub_0_c_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

var boardID boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
}

func main(cur realm) {
	// The read API performs no caller authorization: an EOA and a realm
	// outside the Boards2 namespace must both be able to read.
	testing.SetRealm(testing.NewUserRealm(owner))
	_, found := boards2.GetBoard(uint64(boardID))
	println("user realm:", found)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/unrelated/caller"))
	_, found = boards2.GetBoard(uint64(boardID))
	println("unrelated realm:", found)

	println("threads:", len(boards2.GetThreads(uint64(boardID), 0, 10)))
	println("members:", len(boards2.GetMembers(uint64(boardID), 0, 10)))
}

// Output:
// user realm: true
// unrelated realm: true
// threads: 0
// members: 1
