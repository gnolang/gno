package main

import (
	"chain"
	"chain/runtime"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	body  = "Test reply"
	path  = "test-board/1/2"
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "body")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.EditReply(cross, bid, tid, rid, body)

	// Render content must contain the modified reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> "+body+"\n"))
}

// Output:
// true
