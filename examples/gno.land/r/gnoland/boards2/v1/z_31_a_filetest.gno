// Open board: Test creating a new thread as a non member user
package main

// SEND: 1000000ugnot

import (
	"strings"
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
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	// Non members should be able to create threads
	tid := boards2.CreateThread(cross, bid, title, body)

	// Ensure that returned ID is right
	println(tid == 1)

	// Render content must contains thread's title and body
	content := boards2.Render("test123/1")
	println(strings.HasPrefix(content, "# "+title))
	println(strings.Contains(content, body))
}

// Output:
// true
// true
// true
