package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	// Set an initial message so it can be cleared
	testing.SetRealm(std.NewUserRealm(owner))
	cross(boards2.SetRealmNotice)("This is a test realm message")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.SetRealmNotice)("")

	// Render content must contain the message
	content := boards2.Render("")
	println(strings.HasPrefix(content, "> This is a test realm message\n\n"))
	println(strings.HasPrefix(content, "# Boards"))
}

// Output:
// false
// true
