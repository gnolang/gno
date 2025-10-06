package boards2

import (
	"errors"
	"strconv"
	"time"

	"gno.land/p/nt/avl"
)

const dateFormat = "2006-01-02 3:04pm MST"

// PostID defines a type for Post (Threads/Replies) identifiers.
type PostID uint64

// String returns the ID as a string.
func (id PostID) String() string {
	return strconv.Itoa(int(id))
}

// Key returns the ID as a string with 10 characters padded with zeroes.
// This value can be used for indexing by ID.
func (id PostID) Key() string {
	return padZero(uint64(id), 10)
}

// A Post is a "thread" or a "reply" depending on context.
// A thread is a Post of a Board that holds other replies.
type Post struct {
	ID            PostID
	Board         *Board
	Creator       address
	Title         string
	Body          string
	Hidden        bool
	Readonly      bool
	ThreadID      PostID  // Original Post.ID
	ParentID      PostID  // Parent Post.ID (if reply or repost)
	RepostBoardID BoardID // Original Board.ID (if repost)
	UpdatedAt     time.Time

	flags      avl.Tree // string(address) -> string(reason)
	replies    avl.Tree // Post.ID -> *Post
	repliesAll avl.Tree // Post.ID -> *Post (all replies, for top-level posts)
	reposts    avl.Tree // Board.ID -> Post.ID
	createdAt  time.Time
}

func newPost(board *Board, threadID, id PostID, creator address, title, body string) *Post {
	return &Post{
		Board:     board,
		ThreadID:  threadID,
		ID:        id,
		Creator:   creator,
		Title:     title,
		Body:      body,
		createdAt: time.Now(),
	}
}

// CreatedAt returns the time when post was created.
func (post *Post) CreatedAt() time.Time {
	return post.createdAt
}

// IsRepost checks if current post is repost.
func (post *Post) IsRepost() bool {
	return post.RepostBoardID != 0
}

// IsThread checks if current post is a thread.
func (post *Post) IsThread() bool {
	// repost threads also have parent ID
	return post.ParentID == 0 || post.IsRepost()
}

// Flag add a flag to the post.
// It returns false when the user flagging the post already flagged it.
func (post *Post) Flag(user address, reason string) bool {
	if post.flags.Has(user.String()) {
		return false
	}

	post.flags.Set(user.String(), reason)
	return true
}

// FlagsCount returns the number of time post was flagged.
func (post *Post) FlagsCount() int {
	return post.flags.Size()
}

// AddReply adds a new reply to the post.
// Replies can be added to threads and also to other replies.
func (post *Post) AddReply(creator address, body string) *Post {
	board := post.Board
	pid := board.generateNextPostID()
	pKey := pid.Key()
	reply := newPost(board, post.ThreadID, pid, creator, "", body)
	reply.ParentID = post.ID
	// TODO: Figure out how to remove this redundancy of data "replies==repliesAll" in threads
	post.replies.Set(pKey, reply)
	if post.ThreadID == post.ID {
		post.repliesAll.Set(pKey, reply)
	} else {
		thread, _ := board.GetThread(post.ThreadID)
		thread.repliesAll.Set(pKey, reply)
	}
	return reply
}

// HasReplies checks if post has replies.
func (post *Post) HasReplies() bool {
	return post.replies.Size() > 0
}

// Get returns a post reply.
func (thread *Post) GetReply(pid PostID) (_ *Post, found bool) {
	v, found := thread.repliesAll.Get(pid.Key())
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

// Repost reposts a thread into another boards.
func (post *Post) Repost(creator address, dst *Board, title, body string) *Post {
	if !post.IsThread() {
		panic("post must be a thread to be reposted to another board")
	}

	repost := dst.AddThread(creator, title, body)
	repost.ParentID = post.ID
	repost.RepostBoardID = post.Board.ID

	dst.threads.Set(repost.ID.Key(), repost)
	post.reposts.Set(dst.ID.Key(), repost.ID)
	return repost
}

// DeleteReply deletes a reply from a thread.
func (post *Post) DeleteReply(replyID PostID) error {
	if !post.IsThread() {
		// TODO: Allow removing replies from parent replies too
		panic("cannot delete reply from a non-thread post")
	}

	if post.ID == replyID {
		return errors.New("expected an ID of an inner reply")
	}

	key := replyID.Key()
	v, removed := post.repliesAll.Remove(key)
	if !removed {
		return errors.New("reply not found in thread")
	}

	// TODO: Shouldn't reply be hidden instead of deleted? Maybe replace reply by a deleted message.
	reply := v.(*Post)
	if reply.ParentID != post.ID {
		parent, _ := post.GetReply(reply.ParentID)
		parent.replies.Remove(key)
	} else {
		post.replies.Remove(key)
	}
	return nil
}

// Summary return a summary of the post's body.
// It returns the body making sure that the length is limited to 80 characters.
func (post *Post) Summary() string {
	return summaryOf(post.Body, 80)
}
