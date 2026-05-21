// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_delete_thread_06_filetest

// Open board: Test deleting a reply of another non member user
package z_delete_thread_06_filetest

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
	bid boards.ID // Operate on board DAO
	tid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, true)

	// Make sure user account has the required amount of GNOT for open board actions
	testing.IssueCoins(user, chain.Coins{{"ugnot", 3_000_000_000}})

	// Create a new reply as a non member user
	testing.SetRealm(testing.NewUserRealm(user))
	tid = boards2.CreateThread(cross(cur), bid, "Title", "Body")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user2))

	// Try to delete the thread of another non member user
	boards2.DeleteThread(cross(cur), bid, tid)
}

// Error:
// unauthorized, user g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt doesn't have the required permission
