// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"strconv"

	"gno.land/r/boards"
)

var bid boards.BoardID
var pid boards.PostID

func init() {
	bid = boards.CreateBoard("test_board")
	boards.CreatePost(bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreatePost(bid, "Second Post (title)", "Body of the second post. (body)")
}

func main() {
	rid := boards.CreateReply(bid, pid, "Reply of the second post")
	println(rid)
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// 3
// # Second Post (title)
//
// Body of the second post. (body)
// - by g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2)
//
// > Reply of the second post
// > - by g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2#3)
