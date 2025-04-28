package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	name  = "foo123"
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	cross(boards2.CreateBoard)(name)

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.RenameBoard)(name, "short")
}

// Error:
// the minimum allowed board name length is 6 characters
