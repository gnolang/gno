package main

import (
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

func init() {
	// Set an initial message so it can be cleared
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.SetRealmNotice(cross, "This is a test realm message")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetRealmNotice(cross, "")

	// Render content must contain the message
	content := boards2.Render("")
	println(strings.HasPrefix(content, "> This is a test realm message\n\n"))
	println(strings.HasPrefix(content, "# Boards"))
}

// Output:
// false
// true
