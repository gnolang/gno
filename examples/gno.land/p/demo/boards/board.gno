package boards

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
)

//----------------------------------------
// Board

type BoardID uint64

func BoardIDKey(bid BoardID) string {
	return padZero(uint64(bid), 10)
}

func (bid BoardID) String() string {
	return strconv.Itoa(int(bid))
}

type Board struct {
	id        BoardID // only set for public boards.
	url       string
	name      string
	creator   std.Address
	threads   avl.Tree // Post.id -> *Post
	postsCtr  uint64   // increments Post.id
	createdAt time.Time
	deleted   avl.Tree // TODO reserved for fast-delete.
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

func NewBoard(id BoardID, url string, name string, creator std.Address) *Board {
	return &Board{
		id:        id,
		url:       url,
		name:      name,
		creator:   creator,
		threads:   avl.Tree{},
		createdAt: time.Now(),
		deleted:   avl.Tree{},
	}
}

func (board *Board) Id() BoardID {
	return board.id
}

func (board *Board) Url() string {
	return board.url
}

func (board *Board) Name() string {
	return board.name
}

func (board *Board) Creator() std.Address {
	return board.creator
}

func (board *Board) Threads() avl.Tree {
	return board.threads
}

func (board *Board) CreatedAt() time.Time {
	return board.createdAt
}

func (board *Board) IsPrivate() bool {
	return board.id == 0
}

func (board *Board) GetThread(pid PostID) *Post {
	pidkey := postIDKey(pid)
	postI, exists := board.threads.Get(pidkey)
	if !exists {
		return nil
	}
	return postI.(*Post)
}

func (board *Board) AddThread(creator std.Address, title string, body string) *Post {
	pid := board.incGetPostID()
	pidkey := postIDKey(pid)
	thread := NewPost(board, pid, creator, title, body, pid, 0, 0)
	board.threads.Set(pidkey, thread)
	return thread
}

// NOTE: this can be potentially very expensive for threads with many replies.
// TODO: implement optional fast-delete where thread is simply moved.
func (board *Board) DeleteThread(pid PostID) {
	pidkey := postIDKey(pid)
	_, removed := board.threads.Remove(pidkey)
	if !removed {
		panic("thread does not exist with id " + pid.String())
	}
}

func (board *Board) HasPermission(addr std.Address, perm Permission) bool {
	if board.creator == addr {
		switch perm {
		case EditPermission:
			return true
		case DeletePermission:
			return true
		default:
			return false
		}
	}
	return false
}

func (board *Board) incGetPostID() PostID {
	board.postsCtr++
	return PostID(board.postsCtr)
}

func (board *Board) GetURLFromThreadAndReplyID(threadID, replyID PostID) string {
	if replyID == 0 {
		return board.url + "/" + threadID.String()
	} else {
		return board.url + "/" + threadID.String() + "/" + replyID.String()
	}
}
