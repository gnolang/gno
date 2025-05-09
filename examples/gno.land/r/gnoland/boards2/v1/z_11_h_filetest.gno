package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid   boards2.BoardID
	pid   boards2.PostID
	title = strings.Repeat("X", boards2.MaxThreadTitleLength+1)
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	pid = cross(boards2.CreateThread)(bid, "Foo", "bar")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.EditThread)(bid, pid, title, "bar")
}

// Error:
// thread title is too long, maximum allowed is 100 characters
