package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	// Set an initial message so it can be cleared
	testing.SetRealm(std.NewUserRealm(owner))
	cross(boards2.SetRealmNotice)("This is a test realm message")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.SetRealmNotice)("")

	// Render content must contain the message
	content := cross(boards2.Render)("")
	println(strings.HasPrefix(content, "> This is a test realm message\n\n"))
	println(strings.HasPrefix(content, "# Boards"))
}

// Output:
// false
// true
