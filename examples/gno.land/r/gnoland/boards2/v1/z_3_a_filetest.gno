package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	name    = "foo123"
	newName = "bar123"
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, name, false)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.RenameBoard(cross, name, newName)

	// Ensure board is renamed by the default board owner
	bid2, _ := boards2.GetBoardIDFromName(cross, newName)
	println("IDs match =", bid == bid2)
}

// Output:
// IDs match = true
