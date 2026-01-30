package main

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	bid   boards.ID
	pid   boards.ID
	title = strings.Repeat("X", boards2.MaxThreadTitleLength+1)
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.EditThread(cross, bid, pid, title, "bar")
}

// Error:
// title is too long, maximum allowed is 100 characters
