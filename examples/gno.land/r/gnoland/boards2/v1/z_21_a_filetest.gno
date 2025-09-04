package main

import (
	"chain/runtime"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.SetRealmNotice(cross, "This is a test realm message")

	// Render content must contain the message
	content := boards2.Render("")
	println(strings.HasPrefix(content, "> This is a test realm message\n"))
}

// Output:
// true
