package boards

import (
	"std"
	"strconv"

	"gno.land/p/avl"
)

//----------------------------------------
// Board

type BoardID uint64

type Board struct {
	id        BoardID // only set for public boards.
	url       string
	name      string
	creator   std.Address
	posts     *avl.Tree // Post.id -> *Post
	postsCtr  uint64    // increments Post.id
	createdAt std.Time
}

func newBoard(id BoardID, url string, name string, creator std.Address) *Board {
	if !reName.MatchString(name) {
		panic("invalid name: " + name)
	}
	exists := gBoardsByName.Has(name)
	if exists {
		panic("board already exists")
	}
	return &Board{
		id:        id,
		url:       url,
		name:      name,
		creator:   creator,
		createdAt: std.GetTimestamp(),
	}
}

/* TODO support this once we figure out how to ensure URL correctness.
// A private board is not tracked by gBoards*,
// but must be persisted by the caller's realm.
// Private boards have 0 id and does not ping
// back the remote board on reposts.
func NewPrivateBoard(url string, name string, creator std.Address) *Board {
	return newBoard(0, url, name, creator)
}
*/

func (board *Board) IsPrivate() bool {
	return board.id == 0
}

func (board *Board) GetThread(pid PostID) *Post {
	pidkey := postIDKey(pid)
	_, postI, exists := board.posts.Get(pidkey)
	if !exists {
		return nil
	}
	return postI.(*Post)
}

func (board *Board) AddPost(creator std.Address, title string, body string) *Post {
	pid := board.incGetPostID()
	pidkey := postIDKey(pid)
	post := &Post{
		board:     board,
		id:        pid,
		creator:   creator,
		title:     title,
		body:      body,
		threadID:  pid,
		createdAt: std.GetTimestamp(),
	}
	board.posts, _ = board.posts.Set(pidkey, post)
	return post
}

// Renders the board for display suitable as plaintext in
// console.  This is suitable for demonstration or tests,
// but not for prod.
func (board *Board) RenderBoard() string {
	str := ""
	str += "\\[[post](" + board.GetPostFormURL() + ")]\n\n"
	if board.posts.Size() > 0 {
		board.posts.Iterate("", "", func(n *avl.Tree) bool {
			if str != "" {
				str += "----------------------------------------\n"
			}
			str += n.Value().(*Post).RenderSummary() + "\n"
			return false
		})
	}
	return str
}

func (board *Board) incGetPostID() PostID {
	board.postsCtr++
	return PostID(board.postsCtr)
}

func (board *Board) GetURLFromThreadAndReplyID(threadID, replyID PostID) string {
	if replyID == 0 {
		return board.url + "/" + strconv.Itoa(int(threadID))
	} else {
		return board.url + "/" + strconv.Itoa(int(threadID)) + "/" + strconv.Itoa(int(replyID))
	}
}

func (board *Board) GetPostFormURL() string {
	return "/r/boards?help&__func=CreatePost" +
		"&bid=" + strconv.Itoa(int(board.id)) +
		"&body.type=textarea"
}
