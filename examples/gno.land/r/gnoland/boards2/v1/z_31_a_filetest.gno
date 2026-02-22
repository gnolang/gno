// Open board: Test creating a new thread as a non member user
package main

// SEND: 1000000ugnot

import (
	"chain"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	users "gno.land/r/gnoland/users/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	title         = "Test Thread"
	body          = "Test body"
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, true)

	// Register a username which is required to be able to add a reply
	testing.SetRealm(testing.NewUserRealm(user))
	users.Register(cross, "user123")

	// Make sure user account has the required amount of GNOT for open board actions
	testing.IssueCoins(user, chain.Coins{{"ugnot", 3_000_000_000}})
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	// Non members should be able to create threads
	tid := boards2.CreateThread(cross, bid, title, body)

	// Ensure that returned ID is right
	println(tid == 1)

	// Render content must contains thread's title and body
	println(boards2.Render("test123/1"))
}

// Output:
// true
// # [Boards](/r/gnoland/boards2/v1) › [test123](/r/gnoland/boards2/v1:test123)
//
//
// > [!NOTE] 2 THINGS YOU NEED TO GET STARTED
// > To post and comment, **retrieve your Gno key** with [Adena](https://adena.app) or [gnokey](https://docs.gno.land/users/interact-with-gnokey). You need **at least 3,000 GNOT** in your account (balance requirement, not a transfer).
// ## Test Thread
//
// **[@user123](/u/user123)** · 2009-02-13 11:31pm UTC
// Test body
//
// ↳ [Flag](/r/gnoland/boards2/v1:test123/1/flag) • [Repost](/r/gnoland/boards2/v1:test123/1/repost) • [Comment](/r/gnoland/boards2/v1:test123/1/reply) • [Edit](/r/gnoland/boards2/v1:test123/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=1&threadID=1)
