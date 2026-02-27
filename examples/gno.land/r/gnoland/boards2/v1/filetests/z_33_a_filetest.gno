// Open board: Test creating a new reply as a non member user
package main

// SEND: 1000000ugnot

import (
	"chain"
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	uinit "gno.land/r/sys/users/init"
)

const (
	owner   address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user    address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	comment         = "Test Comment"
)

var (
	bid boards.ID // Operate on board DAO
	tid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, true)
	tid = boards2.CreateThread(cross, bid, "Title", "Body")

	// Register a username which is required to be able to add a reply
	uinit.RegisterUser(cross, "user123", user)

	// Make sure user account has the required amount of GNOT for open board actions
	testing.IssueCoins(user, chain.Coins{{"ugnot", 3_000_000_000}})
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	// Non members should be able to add replies
	boards2.CreateReply(cross, bid, tid, 0, comment)

	// Render content must contain the reply
	content := boards2.Render("test123/1")
	println(strings.Contains(content, "\n> "+comment+"\n"))
}

// Output:
// true
