package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	boardName = "test123"
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	cross(boards2.CreateBoard)(boardName)

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.CreateBoard)(boardName)
}

// Error:
// board already exists
