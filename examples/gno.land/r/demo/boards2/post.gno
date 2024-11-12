package boards

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/txlink"
)

// TODO: Move post and board to a package to improve API and reduce realm size?

type PostID uint64

func (id PostID) String() string {
	return strconv.Itoa(int(id))
}

func (id PostID) Key() string {
	return padZero(uint64(id), 10)
}

// A Post is a "thread" or a "reply" depending on context.
// A thread is a Post of a Board that holds other replies.
// TODO: Figure out "repliesAll", "threadID"
type Post struct {
	board       *Board
	id          PostID
	creator     std.Address
	title       string // optional
	body        string
	replies     avl.Tree // Post.id -> *Post
	repliesAll  avl.Tree // Post.id -> *Post (all replies, for top-level posts)
	reposts     avl.Tree // Board.id -> Post.id
	threadID    PostID   // original Post.id
	parentID    PostID   // parent Post.id (if reply or repost)
	repostBoard BoardID  // original Board.id (if repost)
	createdAt   time.Time
	updatedAt   time.Time
}

func newPost(board *Board, id PostID, creator std.Address, title, body string, threadID, parentID PostID, repostBoard BoardID) *Post {
	return &Post{
		board:       board,
		id:          id,
		creator:     creator,
		title:       title,
		body:        body,
		replies:     avl.Tree{},
		repliesAll:  avl.Tree{},
		reposts:     avl.Tree{},
		threadID:    threadID,
		parentID:    parentID,
		repostBoard: repostBoard,
		createdAt:   time.Now(),
	}
}

func (post *Post) IsThread() bool {
	return post.parentID == 0
}

func (post *Post) GetPostID() PostID {
	return post.id
}

// TODO: Rename to Reply()
func (post *Post) AddReply(creator std.Address, body string) *Post {
	board := post.board
	pid := board.incGetPostID()
	pKey := pid.Key()
	reply := newPost(board, pid, creator, "", body, post.threadID, post.id, 0)
	// TODO: Figure out how to remove this redundancy of data "replies==repliesAll" in threads
	post.replies.Set(pKey, reply)
	if post.threadID == post.id {
		post.repliesAll.Set(pKey, reply)
	} else {
		thread := board.GetThread(post.threadID)
		thread.repliesAll.Set(pKey, reply)
	}
	return reply
}

func (post *Post) Update(title string, body string) {
	post.title = title
	post.body = body
	post.updatedAt = time.Now()
}

func (thread *Post) GetReply(pid PostID) *Post {
	v, found := thread.repliesAll.Get(pid.Key())
	if !found {
		return nil
	}
	return v.(*Post)
}

func (post *Post) AddRepostTo(creator std.Address, title, body string, dst *Board) *Post {
	if !post.IsThread() {
		panic("cannot repost non-thread post")
	}

	pid := dst.incGetPostID()
	repost := newPost(dst, pid, creator, title, body, pid, post.id, post.board.id)
	dst.threads.Set(pid.Key(), repost)
	if !dst.IsPrivate() {
		post.reposts.Set(dst.id.Key(), pid)
	}
	return repost
}

func (thread *Post) DeletePost(pid PostID) {
	if thread.id == pid {
		panic("should not happen")
	}

	pKey := pid.Key()
	v, removed := thread.repliesAll.Remove(pKey)
	if !removed {
		panic("post not found in thread")
	}

	post := v.(*Post)
	if post.parentID != thread.id {
		parent := thread.GetReply(post.parentID)
		parent.replies.Remove(pKey)
	} else {
		thread.replies.Remove(pKey)
	}
}

// TODO: Change HasPermission to use a new authorization interface's `CanDo()`
func (post *Post) HasPermission(addr std.Address, perm Permission) bool {
	if post.creator == addr {
		switch perm {
		case EditPermission:
			return true
		case DeletePermission:
			return true
		default:
			return false
		}
	}
	// post notes inherit permissions of the board.
	return post.board.HasPermission(addr, perm)
}

func (post *Post) GetSummary() string {
	return summaryOf(post.body, 80)
}

func (post *Post) GetURL() string {
	if post.IsThread() {
		return post.board.GetURLFromThreadAndReplyID(post.id, 0)
	}
	return post.board.GetURLFromThreadAndReplyID(post.threadID, post.id)
}

func (post *Post) GetReplyFormURL() string {
	return txlink.URL("CreateReply",
		"bid", post.board.id.String(),
		"threadid", post.threadID.String(),
		"postid", post.id.String(),
	)
}

func (post *Post) GetRepostFormURL() string {
	return txlink.URL("CreateRepost",
		"bid", post.board.id.String(),
		"postid", post.id.String(),
	)
}

func (post *Post) GetDeleteFormURL() string {
	return txlink.URL("DeletePost",
		"bid", post.board.id.String(),
		"threadid", post.threadID.String(),
		"postid", post.id.String(),
	)
}

func (post *Post) RenderSummary() string {
	if post.repostBoard != 0 {
		dstBoard := getBoard(post.repostBoard)
		if dstBoard == nil {
			panic("repostBoard does not exist")
		}

		thread := dstBoard.GetThread(PostID(post.parentID))
		if thread == nil {
			return "reposted post does not exist"
		}
		return "Repost: " + post.GetSummary() + "\n" + thread.RenderSummary()
	}

	var s string
	if post.title != "" { // TODO: Add a newLink() function
		s += "## [" + summaryOf(post.title, 80) + "](" + post.GetURL() + ")\n"
		s += "\n"
	}
	s += post.GetSummary() + "\n"
	s += "\\- " + displayAddressMD(post.creator) + ","
	s += " [" + post.createdAt.Format("2006-01-02 3:04pm MST") + "](" + post.GetURL() + ")"
	s += " \\[[x](" + post.GetDeleteFormURL() + ")]"
	s += " (" + strconv.Itoa(post.replies.Size()) + " replies)"
	s += " (" + strconv.Itoa(post.reposts.Size()) + " reposts)" + "\n"
	return s
}

func (post *Post) RenderPost(indent string, levels int) string {
	if post == nil {
		return "nil post"
	}

	var s string
	if post.title != "" {
		s += indent + "# " + post.title + "\n"
		s += indent + "\n"
	}

	s += indentBody(indent, post.body) + "\n" // TODO: indent body lines.

	s += indent + "\\- " + displayAddressMD(post.creator) + ", "
	s += "[" + post.createdAt.Format("2006-01-02 3:04pm (MST)") + "](" + post.GetURL() + ")"
	s += " \\[[reply](" + post.GetReplyFormURL() + ")]"
	if post.IsThread() {
		s += " \\[[repost](" + post.GetRepostFormURL() + ")]"
	}
	s += " \\[[x](" + post.GetDeleteFormURL() + ")]\n"

	if levels > 0 {
		if post.replies.Size() > 0 {
			post.replies.Iterate("", "", func(key string, value interface{}) bool {
				s += indent + "\n"
				s += value.(*Post).RenderPost(indent+"> ", levels-1)
				return false
			})
		}
	} else {
		if post.replies.Size() > 0 {
			s += indent + "\n"
			s += indent + "_[see all " + strconv.Itoa(post.replies.Size()) + " replies](" + post.GetURL() + ")_\n"
		}
	}
	return s
}

// render reply and link to context thread
func (post *Post) RenderInner() string {
	if post.IsThread() {
		panic("unexpected thread")
	}

	threadID := post.threadID
	// replyID := post.id
	parentID := post.parentID
	s := "_[see thread](" + post.board.GetURLFromThreadAndReplyID(threadID, 0) + ")_\n\n"
	thread := post.board.GetThread(post.threadID)

	var parent *Post
	if thread.id == parentID {
		parent = thread
	} else {
		parent = thread.GetReply(parentID)
	}

	s += parent.RenderPost("", 0)
	s += "\n"
	s += post.RenderPost("> ", 5)
	return s
}
