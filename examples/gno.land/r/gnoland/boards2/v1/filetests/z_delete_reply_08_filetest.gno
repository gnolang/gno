// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_delete_reply_08_filetest

// Open board: Test deleting a reply of another non member user
package z_delete_reply_08_filetest

// SEND: 1000000ugnot

import (
	"chain"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	user2 address = "g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt"
)

var (
	bid      boards.ID // Operate on board DAO
	tid, rid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, true)
	tid = boards2.CreateThread(cross(cur), bid, "Title", "Body")

	// Make sure user account has the required amount of GNOT for open board actions
	testing.IssueCoins(user, chain.Coins{{"ugnot", 3_000_000_000}})

	// Create a new reply as a non member user
	testing.SetRealm(testing.NewUserRealm(user))
	rid = boards2.CreateReply(cross(cur), bid, tid, 0, "Comment")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user2))

	// Try to delete the reply of another non member user
	boards2.DeleteReply(cross(cur), bid, tid, rid)
}

// Error:
// unauthorized, user g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt doesn't have the required permission
