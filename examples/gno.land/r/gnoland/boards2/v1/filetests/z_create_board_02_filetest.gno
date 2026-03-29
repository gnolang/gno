package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
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
