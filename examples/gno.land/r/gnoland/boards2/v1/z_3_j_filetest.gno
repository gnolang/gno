package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	name  = "foo123"
)

func init() {
	testing.SetOriginCaller(owner)
	boards2.CreateBoard(name)

	testing.SetOriginCaller(user)
}

func main() {
	boards2.RenameBoard(name, "barbaz")
}

// Error:
// unauthorized
