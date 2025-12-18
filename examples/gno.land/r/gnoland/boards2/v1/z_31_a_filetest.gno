package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	title1         = "Foo1"
	body1          = "bar1"
	title2         = "Foo2"
	body2          = "bar2"
)

var (
	bid  boards2.BoardID
	tid1 boards2.PostID
	tid2 boards2.PostID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	tid1 = boards2.CreateThread(cross, bid, title1, body1)
	tid2 = boards2.CreateThread(cross, bid, title2, body2)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Just get the total count of threads
	total, _ := boards2.GetPosts(bid, 0, 0, 0, 0)
	println(total == 2)

	// Get all threads
	total, threads := boards2.GetPosts(bid, 0, 0, 0, 10)
	println(len(threads) == 2)

	// Check contents
	println(threads[0].ID == tid1)
	println(threads[0].Creator == owner)
	println(threads[0].Title == title1)
	println(threads[0].Body == body1)

	println(threads[1].ID == tid2)
	println(threads[1].Creator == owner)
	println(threads[1].Title == title2)
	println(threads[1].Body == body2)

	// Get starting from the second thread
	total, threads = boards2.GetPosts(bid, 0, 0, 1, 10)
	println(len(threads) == 1)
	println(threads[0].ID == tid2)
}

// Output:
// true
// true
// true
// true
// true
// true
// true
// true
// true
// true
// true
// true
