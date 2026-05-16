// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_reply_15_filetest

// Open board: Test creating a new reply as a non member user that has no GNOT
package z_create_reply_15_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	user    address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	comment         = "Test Comment"
)

var (
	bid boards.ID // Operate on board DAO
	tid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test123", false, true)
	tid = boards2.CreateThread(cross2(cur), bid, "Title", "Body")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Non members should be able to add replies only if they have enough GNOT
	boards2.CreateReply(cross2(cur), bid, tid, 0, comment)
}

// Error:
// caller is not allowed to comment: account amount is lower than 3000 GNOT
