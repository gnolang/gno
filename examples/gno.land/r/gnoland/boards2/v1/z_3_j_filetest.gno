package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	name  = "foo123"
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	cross(boards2.CreateBoard)(name)
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	cross(boards2.RenameBoard)(name, "barbaz")
}

// Error:
// unauthorized
