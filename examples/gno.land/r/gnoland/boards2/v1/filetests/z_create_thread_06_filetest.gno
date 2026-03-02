// Open board: Test creating a new thread as a non member user that has no username registered
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, true)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	// Non members should be able to create threads only if they have a registered user
	boards2.CreateThread(cross, bid, "Title", "Body")
}

// Error:
// a registered username is required to create threads on open boards
