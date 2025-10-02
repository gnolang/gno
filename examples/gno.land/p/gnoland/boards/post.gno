package boards

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/nt/avl"
)

// A Post is a "thread" or a "reply" depending on context.
// A thread is a Post of a board that holds other replies.
type Post struct {
	ID              ID
	ParentID        ID     // Parent Post.ID for all replies or reposted thread
	ThreadID        ID     // Post.ID of the thread where reply is created
	OriginalBoardID ID     // Original Board.ID when thread is a repost
	Board           *Board // Board where post was created
	Title           string
	Body            string
	Hidden          bool
	Readonly        bool
	Creator         address
	CreatedAt       time.Time

	// TODO: Make these properties public introducing optional support for them (storages)
	flags      avl.Tree // address -> string(reason)
	replies    avl.Tree // string(Post.ID) -> *Post (direct post replies)
	repliesAll avl.Tree // string(Post.ID) -> *Post (all thread replies)
	reposts    avl.Tree // string(Board.ID) -> Post.ID
}

// Summary return a summary of the post's body.
// It returns the body making sure that the length is limited to 80 characters.
func (post Post) Summary() string {
	return SummaryOf(post.Body, 80)
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

// FlagsCount returns the number of times post was flagged.
func (post Post) FlagsCount() int {
	return post.flags.Size()
}

// HasReplies checks if a thread or reply has replies.
func (post Post) HasReplies() bool {
	return post.replies.Size() > 0
}

// Save stores the thread (or repost) within it's parent board, or
// if post is a reply it stores the reply within it's parent thread.
func (post *Post) Save() error {
	if post.Board == nil {
		return errors.New("board is required to save a post")
	}

	if IsThread(post) {
		return post.Board.Threads.Save(post)
	}

	if IsRepost(post) {
		thread, found := post.Board.Threads.Get(post.ParentID)
		if !found {
			return errors.New("repost origin thread not found: ID " + post.ParentID.String())
		}

		if err := post.Board.Threads.Save(post); err != nil {
			return err
		}
		return thread.saveRepost(post) // TODO: Replace by thread.Reposts.Save(post)
	}

	// Post is a reply, so get the thread where reply has been created
	thread, found := post.Board.Threads.Get(post.ThreadID)
	if !found {
		return errors.New("reply's thread not found: ID " + post.ThreadID.String())
	}

	// Add reply to the flat replies index. This index allows listing all thread
	// replies without needing to recursively explore the child replies of each reply.
	key := makePostKey(post.ID)
	thread.repliesAll.Set(key, post)

	// Add reply to it's parent index.
	// Reply parent can be a thread or another reply.
	parent := thread
	if post.ParentID != thread.ID {
		parentKey := makePostKey(post.ParentID)
		v, found := thread.repliesAll.Get(parentKey)
		if !found {
			return errors.New("reply's parent post not found: ID " + post.ParentID.String())
		}

		parent = v.(*Post)
	}

	parent.replies.Set(key, post)
	return nil
}

// IsThread checks if a post is a thread.
// When a post is not a thread it's considered a thread's reply/comment.
func IsThread(p *Post) bool {
	if p == nil {
		return false
	}
	return p.ThreadID == p.ID
}

// IsRepost checks if a thread is a repost.
func IsRepost(thread *Post) bool {
	if thread == nil {
		return false
	}
	return thread.OriginalBoardID != 0
}

// SummaryOf returns a summary of a text.
func SummaryOf(text string, length int) string {
	lines := strings.SplitN(text, "\n", 2)
	line := lines[0]
	if len(line) > length {
		line = line[:(length-3)] + "..."
	} else if len(lines) > 1 {
		line = line + "..."
	}
	return line
}
