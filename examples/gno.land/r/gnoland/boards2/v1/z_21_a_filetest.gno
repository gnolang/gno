package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetRealmNotice(cross, "This is a test realm message")

	// Render content must contain the message
	println(boards2.Render(""))
}

// Output:
// > [!INFO] Notice
// > This is a test realm message
// # Boards
// [Create Board](/r/gnoland/boards2/v1:create-board) • [List Admin Users](/r/gnoland/boards2/v1:admin-users) • [Help](/r/gnoland/boards2/v1:help)
//
// > [!NOTE] 2 THINGS YOU NEED TO GET STARTED
// > To post and comment, **retrieve your Gno key** with [Adena](https://adena.app) or [gnokey](https://docs.gno.land/users/interact-with-gnokey). You need **at least 3,000 GNOT** in your account (balance requirement, not a transfer).
// ---
// ### Currently there are no boards
// Be the first to [create a new board](/r/gnoland/boards2/v1:create-board)!
