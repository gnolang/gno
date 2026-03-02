// Open board: Test creating a new reply as a non member user that has no username registered
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
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
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	// Non members should be able to add replies only if they have a registered user
	boards2.CreateReply(cross, bid, tid, 0, comment)
}

// Error:
// a registered username is required to comment on open boards
