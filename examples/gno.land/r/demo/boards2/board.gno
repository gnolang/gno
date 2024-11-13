package boards

import (
	"regexp"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/txlink"
)

var reBoardName = regexp.MustCompile(`^[a-z]{3}[_a-z0-9]{0,23}[0-9]{3}$`)

type BoardID uint64

func (id BoardID) String() string {
	return strconv.Itoa(int(id))
}

func (id BoardID) Key() string {
	return padZero(uint64(id), 10)
}

type Board struct {
	id        BoardID // only set for public boards.
	name      string
	creator   std.Address
	threads   avl.Tree // Post.id -> *Post
	postsCtr  uint64   // increments Post.id
	createdAt time.Time
	deleted   avl.Tree // TODO reserved for fast-delete.
}

func newBoard(id BoardID, name string, creator std.Address) *Board {
	assertIsBoardName(name)

	if gBoardsByName.Has(name) {
		panic("board already exists")
	}

	return &Board{
		id:        id,
		name:      name,
		creator:   creator,
		threads:   avl.Tree{},
		createdAt: time.Now(),
		deleted:   avl.Tree{},
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

// GetURL returns the relative URL of the board.
func (board *Board) GetURL() string {
	return strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land") + ":" + board.name
}

func (board *Board) GetThread(threadID PostID) (_ *Post, found bool) {
	v, found := board.threads.Get(threadID.Key())
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

func (board *Board) AddThread(creator std.Address, title string, body string) *Post {
	pid := board.incGetPostID()
	thread := newPost(board, pid, creator, title, body, pid, 0, 0)
	board.threads.Set(pid.Key(), thread)
	return thread
}

// NOTE: this can be potentially very expensive for threads with many replies.
// TODO: implement optional fast-delete where thread is simply moved.
func (board *Board) DeleteThread(pid PostID) {
	_, removed := board.threads.Remove(pid.Key())
	if !removed {
		panic("thread does not exist with id " + pid.String())
	}
}

// TODO: Change HasPermission to use a new authorization interface's `CanDo()`
func (board *Board) HasPermission(addr std.Address, perm Permission) bool {
	if board.creator == addr {
		switch perm {
		case PermissionEdit:
			return true
		case PermissionDelete:
			return true
		default:
			return false
		}
	}
	return false
}

func (board *Board) Render() string {
	s := "\\[" + newLink("post", board.GetPostFormURL()) + "]\n\n"
	if board.threads.Size() > 0 {
		board.threads.Iterate("", "", func(_ string, v interface{}) bool {
			s += "----------------------------------------\n"
			s += v.(*Post).RenderSummary() + "\n"
			return false
		})
	}
	return s
}

func (board *Board) incGetPostID() PostID {
	board.postsCtr++
	return PostID(board.postsCtr)
}

func (board *Board) GetURLFromThreadID(threadID PostID) string {
	return board.GetURL() + "/" + threadID.String()
}

func (board *Board) GetURLFromReplyID(threadID, replyID PostID) string {
	return board.GetURL() + "/" + threadID.String() + "/" + replyID.String()
}

func (board *Board) GetPostFormURL() string {
	return txlink.URL("CreateThread", "bid", board.id.String())
}

func assertIsBoardName(name string) {
	if !reBoardName.MatchString(name) {
		panic("invalid board name: " + name)
	}
}
