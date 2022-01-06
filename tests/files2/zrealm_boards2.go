// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"gno.land/p/testutils"
	"gno.land/r/boards"
)

var post *boards.Post

func init() {
	addr0 := testutils.TestAddress("addr0")
	board := boards.NewPrivateBoard("test_board", addr0)
	board.AddPost(addr0, "First Post (title)", "Body of the first post. (body)")
	post = board.AddPost(addr0, "Second Post (title)", "Body of the second post. (body)")
	post.AddReply(addr0, "Reply of the second post")
}

func main() {
	println(post.Render(""))
}

// Output:
// # Second Post (title)
//
// Body of the second post. (body)
// - by g1v9jxgu3sta047h6lta047h6lta047h6l2czdz2, 1970-01-01 12:00am (UTC)
//
// > Reply of the second post
// > - by g1v9jxgu3sta047h6lta047h6lta047h6l2czdz2, 1970-01-01 12:00am (UTC)
