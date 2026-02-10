package boards

import (
	"strconv"
	"time"

	"gno.land/p/nt/avl"
)

//----------------------------------------
// Post

// NOTE: a PostID is relative to the board.
type PostID uint64

func (pid PostID) String() string {
	return strconv.Itoa(int(pid))
}

// A Post is a "thread" or a "reply" depending on context.
// A thread is a Post of a Board that holds other replies.
type Post struct {
	board       *Board
	id          PostID
	creator     address
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

func newPost(board *Board, id PostID, creator address, title, body string, threadID, parentID PostID, repostBoard BoardID) *Post {
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

func (post *Post) AddReply(creator address, body string) *Post {
	board := post.board
	pid := board.incGetPostID()
	pidkey := postIDKey(pid)
	reply := newPost(board, pid, creator, "", body, post.threadID, post.id, 0)
	post.replies.Set(pidkey, reply)
	if post.threadID == post.id {
		post.repliesAll.Set(pidkey, reply)
	} else {
		thread := board.GetThread(post.threadID)
		thread.repliesAll.Set(pidkey, reply)
	}
	return reply
}

func (post *Post) Update(title string, body string) {
	post.title = title
	post.body = body
	post.updatedAt = time.Now()
}

func (thread *Post) GetReply(pid PostID) *Post {
	pidkey := postIDKey(pid)
	replyI, ok := thread.repliesAll.Get(pidkey)
	if !ok {
		return nil
	} else {
		return replyI.(*Post)
	}
}

func (post *Post) AddRepostTo(creator address, title, body string, dst *Board) *Post {
	if !post.IsThread() {
		panic("cannot repost non-thread post")
	}
	pid := dst.incGetPostID()
	pidkey := postIDKey(pid)
	repost := newPost(dst, pid, creator, title, body, pid, post.id, post.board.id)
	dst.threads.Set(pidkey, repost)
	if !dst.IsPrivate() {
		bidkey := boardIDKey(dst.id)
		post.reposts.Set(bidkey, pid)
	}
	return repost
}

func (thread *Post) DeletePost(pid PostID) {
	if thread.id == pid {
		panic("should not happen")
	}
	pidkey := postIDKey(pid)
	postI, removed := thread.repliesAll.Remove(pidkey)
	if !removed {
		panic("post not found in thread")
	}
	post := postI.(*Post)
	if post.parentID != thread.id {
		parent := thread.GetReply(post.parentID)
		parent.replies.Remove(pidkey)
	} else {
		thread.replies.Remove(pidkey)
	}
}

func (post *Post) HasPermission(addr address, perm Permission) bool {
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
		return post.board.GetURLFromThreadAndReplyID(
			post.id, 0)
	} else {
		return post.board.GetURLFromThreadAndReplyID(
			post.threadID, post.id)
	}
}

func (post *Post) GetReplyFormURL() string {
	return gRealmLink.Call("CreateReply",
		"bid", post.board.id.String(),
		"threadid", post.threadID.String(),
		"postid", post.id.String(),
	)
}

func (post *Post) GetRepostFormURL() string {
	return gRealmLink.Call("CreateRepost",
		"bid", post.board.id.String(),
		"postid", post.id.String(),
	)
}

func (post *Post) GetDeleteFormURL() string {
	return gRealmLink.Call("DeletePost",
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
	str := ""
	if post.title != "" {
		str += "## [" + summaryOf(post.title, 80) + "](" + post.GetURL() + ")\n"
		str += "\n"
	}
	str += post.GetSummary() + "\n"
	str += "\\- " + displayAddressMD(post.creator) + ","
	str += " [" + post.createdAt.Format("2006-01-02 3:04pm MST") + "](" + post.GetURL() + ")"
	str += " \\[[x](" + post.GetDeleteFormURL() + ")]"
	str += " (" + strconv.Itoa(post.replies.Size()) + " replies)"
	str += " (" + strconv.Itoa(post.reposts.Size()) + " reposts)" + "\n"
	return str
}

func (post *Post) RenderPost(indent string, levels int) string {
	if post == nil {
		return "nil post"
	}
	str := ""
	if post.title != "" {
		str += indent + "# " + post.title + "\n"
		str += indent + "\n"
	}
	str += indentBody(indent, post.body) + "\n" // TODO: indent body lines.
	str += indent + "\\- " + displayAddressMD(post.creator) + ", "
	str += "[" + post.createdAt.Format("2006-01-02 3:04pm (MST)") + "](" + post.GetURL() + ")"
	str += " \\[[reply](" + post.GetReplyFormURL() + ")]"
	if post.IsThread() {
		str += " \\[[repost](" + post.GetRepostFormURL() + ")]"
	}
	str += " \\[[x](" + post.GetDeleteFormURL() + ")]\n"
	if levels > 0 {
		if post.replies.Size() > 0 {
			post.replies.Iterate("", "", func(key string, value any) bool {
				str += indent + "\n"
				str += value.(*Post).RenderPost(indent+"> ", levels-1)
				return false
			})
		}
	} else {
		if post.replies.Size() > 0 {
			str += indent + "\n"
			str += indent + "_[see all " + strconv.Itoa(post.replies.Size()) + " replies](" + post.GetURL() + ")_\n"
		}
	}
	return str
}

// render reply and link to context thread
func (post *Post) RenderInner() string {
	if post.IsThread() {
		panic("unexpected thread")
	}
	threadID := post.threadID
	// replyID := post.id
	parentID := post.parentID
	str := ""
	str += "_[see thread](" + post.board.GetURLFromThreadAndReplyID(
		threadID, 0) + ")_\n\n"
	thread := post.board.GetThread(post.threadID)
	var parent *Post
	if thread.id == parentID {
		parent = thread
	} else {
		parent = thread.GetReply(parentID)
	}
	str += parent.RenderPost("", 0)
	str += "\n"
	str += post.RenderPost("> ", 5)
	return str
}
