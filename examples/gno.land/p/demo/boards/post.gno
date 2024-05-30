package boards

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
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

func NewPost(board *Board, id PostID, creator std.Address, title, body string, threadID, parentID PostID, repostBoard BoardID) *Post {
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

func (post *Post) Board() Board {
	return *post.board
}

func (post *Post) Id() PostID {
	return post.id
}

func (post *Post) Creator() std.Address {
	return post.creator
}

func (post *Post) Title() string {
	return post.title
}

func (post *Post) Body() string {
	return post.body
}

func (post *Post) Replies() avl.Tree {
	return post.replies
}

func (post *Post) RepliesAll() avl.Tree {
	return post.repliesAll
}

func (post *Post) Reposts() avl.Tree {
	return post.reposts
}

func (post *Post) ThreadID() PostID {
	return post.threadID
}

func (post *Post) ParentID() PostID {
	return post.parentID
}

func (post *Post) IsThread() bool {
	return post.parentID == 0
}

func (post *Post) RepostBoard() BoardID {
	return post.repostBoard
}

func (post *Post) CreatedAt() time.Time {
	return post.createdAt
}

func (post *Post) AddReply(creator std.Address, body string) *Post {
	board := post.board
	pid := board.incGetPostID()
	pidkey := postIDKey(pid)
	reply := NewPost(board, pid, creator, "", body, post.threadID, post.id, 0)
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

func (post *Post) AddRepostTo(creator std.Address, title, body string, dst *Board) *Post {
	if !post.IsThread() {
		panic("cannot repost non-thread post")
	}
	pid := dst.incGetPostID()
	pidkey := postIDKey(pid)
	repost := NewPost(dst, pid, creator, title, body, pid, post.id, post.board.id)
	dst.threads.Set(pidkey, repost)
	if !dst.IsPrivate() {
		bidkey := BoardIDKey(dst.id)
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
		return post.board.GetURLFromThreadAndReplyID(
			post.id, 0)
	} else {
		return post.board.GetURLFromThreadAndReplyID(
			post.threadID, post.id)
	}
}
