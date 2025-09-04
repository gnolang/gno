package main

import (
	"chain"
	"chain/runtime"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid   boards2.BoardID
	pid   boards2.PostID
	title = strings.Repeat("X", boards2.MaxThreadTitleLength+1)
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.EditThread(cross, bid, pid, title, "bar")
}

// Error:
// thread title is too long, maximum allowed is 100 characters
