package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	name          = "foo123"
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross, name, false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.RenameBoard(cross, name, "")
}

// Error:
// board name is empty
