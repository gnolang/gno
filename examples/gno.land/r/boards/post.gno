package boards

import (
	"std"
	"strconv"

	"gno.land/p/avl"
)

//----------------------------------------
// Post

// NOTE: a PostID is relative to the board.
type PostID uint64

// A Post is a "thread" or a "reply" depending on context.
// A thread is a Post of a Board that holds other replies.
type Post struct {
	board       *Board
	id          PostID
	creator     std.Address
	title       string // optional
	body        string
	replies     *avl.Tree // Post.id -> *Post
	repliesAll  *avl.Tree // Post.id -> *Post (all replies, for top-level posts)
	reposts     *avl.Tree // Board.id -> Post.id
	threadID    PostID    // original Post.id
	replyTo     PostID    // parent Post.id (if reply or repost)
	repostBoard BoardID   // original Board.id (if repost)
	createdAt   std.Time
}

func (post *Post) IsThread() bool {
	return post.replyTo == 0
}

func (post *Post) GetPostID() PostID {
	return post.id
}

func (post *Post) AddReply(creator std.Address, body string) *Post {
	board := post.board
	pid := board.incGetPostID()
	pidkey := postIDKey(pid)
	reply := &Post{
		board:     board,
		id:        pid,
		creator:   creator,
		body:      body,
		threadID:  post.threadID,
		replyTo:   post.id,
		createdAt: std.GetTimestamp(),
	}
	post.replies, _ = post.replies.Set(pidkey, reply)
	if post.threadID == post.id {
		post.repliesAll, _ = post.repliesAll.Set(pidkey, reply)
	} else {
		thread := board.GetThread(post.threadID)
		thread.repliesAll, _ = thread.repliesAll.Set(pidkey, reply)
	}
	return reply
}

func (thread *Post) GetReply(pid PostID) *Post {
	pidkey := postIDKey(pid)
	_, replyI, ok := thread.repliesAll.Get(pidkey)
	if !ok {
		return nil
	} else {
		return replyI.(*Post)
	}
}

func (post *Post) AddRepostTo(creator std.Address, title, body string, dst *Board) *Post {
	pid := dst.incGetPostID()
	pidkey := postIDKey(pid)
	repost := &Post{
		board:       dst,
		id:          pid,
		creator:     creator,
		title:       title,
		body:        body,
		threadID:    pid,
		replyTo:     post.id,
		repostBoard: post.board.id,
		createdAt:   std.GetTimestamp(),
	}
	dst.posts, _ = dst.posts.Set(pidkey, repost)
	if !dst.IsPrivate() {
		bidkey := boardIDKey(dst.id)
		post.reposts, _ = post.reposts.Set(bidkey, pid)
	}
	return repost
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
	return "/r/boards?help&__func=CreateReply" +
		"&bid=" + strconv.Itoa(int(post.board.id)) +
		"&threadid=" + strconv.Itoa(int(post.threadID)) +
		"&postid=" + strconv.Itoa(int(post.id)) +
		"&body.type=textarea"
}

func (post *Post) RenderSummary() string {
	str := ""
	if post.title != "" {
		str += "## [" + summaryOf(post.title, 80) + "](" + post.GetURL() + ")\n"
		str += "\n"
	}
	str += post.GetSummary() + "\n"
	str += "\\- " + displayAddressMD(post.creator) + ", "
	str += "[" + std.FormatTimestamp(post.createdAt, "2006-01-02 3:04pm MST") + "](" + post.GetURL() + ") "

	str += "(" + strconv.Itoa(post.replies.Size()) + " replies)" + "\n"
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
	str += "[" + std.FormatTimestamp(post.createdAt, "2006-01-02 3:04pm (MST)") + "](" + post.GetURL() + ")"
	str += " [reply](" + post.GetReplyFormURL() + ")" + "\n"
	if levels > 0 {
		if post.replies.Size() > 0 {
			post.replies.Iterate("", "", func(n *avl.Tree) bool {
				str += indent + "\n"
				str += n.Value().(*Post).RenderPost(indent+"> ", levels-1)
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
	replyID := post.id
	parentID := post.replyTo
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
