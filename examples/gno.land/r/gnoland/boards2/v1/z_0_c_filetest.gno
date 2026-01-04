package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	boardName         = "test123"
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross, boardName, false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross, boardName, false, false)
}

// Error:
// board already exists
