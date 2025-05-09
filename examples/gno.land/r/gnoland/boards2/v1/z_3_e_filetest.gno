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
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.RenameBoard)(name, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}

// Error:
// addresses are not allowed as board name
