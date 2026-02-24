// Open board: Test deleting a reply of another non member user
package main

// SEND: 1000000ugnot

import (
	"chain"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	susers "gno.land/r/sys/users"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	user2 address = "g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt"
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
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "user123", user)
	testing.SetRealm(testing.NewUserRealm(user))
	tid = boards2.CreateThread(cross, bid, "Title", "Body")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user2))

	// Try to delete the thread of another non member user
	boards2.DeleteThread(cross, bid, tid)
}

// Error:
// unauthorized
