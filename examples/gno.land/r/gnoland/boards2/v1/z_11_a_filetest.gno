package main

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	title         = "Test Thread"
	body          = "Test body"
	path          = "test-board/1"
)

var (
	bid boards.ID
	pid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.EditThread(cross, bid, pid, title, body)

	// Render content must contains thread's title and body
	content := boards2.Render(path)
	println(strings.HasPrefix(content, "# "+title))
	println(strings.Contains(content, body))
}

// Output:
// true
// true
