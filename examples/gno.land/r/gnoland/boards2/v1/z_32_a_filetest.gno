// Open board: Test deleting a thread as the non member user that created it
package main

// SEND: 1000000ugnot

import (
	"chain"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	uinit "gno.land/r/sys/users/init"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid boards.ID // Operate on board DAO
	tid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, true)

	// Make sure user account has the required amount of GNOT for open board actions
	testing.IssueCoins(user, chain.Coins{{"ugnot", 3_000_000_000}})

	// Register a username which is required to be able to add
	// a reply and then create a new reply as a non member user.
	uinit.RegisterUser(cross, "user123", user)
	testing.SetRealm(testing.NewUserRealm(user))
	tid = boards2.CreateThread(cross, bid, "Title", "Body")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	// Delete the thread as the non member user that created it
	boards2.DeleteThread(cross, bid, tid)

	// Ensure reply doesn't exist
	println(tid == 1)
	println(boards2.Render("test123/1"))
}

// Output:
// true
// Thread not found
