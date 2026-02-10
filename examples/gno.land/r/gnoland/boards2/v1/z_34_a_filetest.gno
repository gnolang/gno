// Open board: Test deleting a reply as the non member user that created it
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
)

var (
	bid      boards.ID // Operate on board DAO
	tid, rid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, true)
	tid = boards2.CreateThread(cross, bid, "Title", "Body")

	// Make sure user account has the required amount of GNOT for open board actions
	testing.IssueCoins(user, chain.Coins{{"ugnot", 3_000_000_000}})

	// Register a username which is required to be able to add
	// a reply and then create a new reply as a non member user.
	testing.SetRealm(testing.NewUserRealm(user))
	users.Register(cross, "user123")
	rid = boards2.CreateReply(cross, bid, tid, 0, "Comment")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	// Delete the reply as the non member user that created it
	boards2.DeleteReply(cross, bid, tid, rid)

	// Ensure reply doesn't exist
	println(rid == 2)
	println(boards2.Render("test123/1/2"))
}

// Output:
// true
// Reply does not exist with ID: 2
