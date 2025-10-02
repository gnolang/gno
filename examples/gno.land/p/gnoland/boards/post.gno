package boards

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
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
	Replies         PostStorage
	Creator         address
	CreatedAt       time.Time

	// TODO: Make these properties public introducing optional support for them (storages)
	flags   avl.Tree // address -> string(reason)
	reposts avl.Tree // string(Board.ID) -> Post.ID
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

// Save stores the thread (or repost) within it's parent board, or
// if post is a reply it stores the reply within it's parent thread.
func (post *Post) Save() error {
	if post.Board == nil {
		return errors.New("board is required to save a post")
	}

	if post.Board.Threads == nil {
		return ufmt.Errorf("save error, threads support not available in board ID %s", post.Board.ID)
	}

	// Post is a thread
	if IsThread(post) {
		return post.saveThread()
	}

	// Post is a repost of a thread
	if IsRepost(post) {
		return post.saveRepost()
	}

	// Post is a reply
	return post.saveReply()
}

func (thread *Post) saveThread() error {
	// Add current thread to the board
	return thread.Board.Threads.Add(thread)
}

func (repost *Post) saveRepost() error {
	thread, found := repost.Board.Threads.Get(repost.ParentID)
	if !found {
		return ufmt.Errorf("repost save error, origin thread ID %s not found", repost.ParentID)
	}

	// TODO: Check if thread support reposts
	thread.reposts.Set(repost.Board.ID.Key(), repost.ID) // TODO: Replace by thread.Reposts.Add(post)

	return repost.Board.Threads.Add(repost)
}

func (reply *Post) saveReply() error {
	// Get the thread where reply has been created
	thread, found := reply.Board.Threads.Get(reply.ThreadID)
	if !found {
		return ufmt.Errorf("reply save error, thread ID %s not found", reply.ThreadID)
	}

	if err := checkPostSupportsReplies(thread); err != nil {
		return err
	}

	// Add reply to the flat replies index. This index allows listing all thread
	// replies without needing to recursively explore the child replies of each reply.
	thread.Replies.Add(reply)

	// Add reply to it's parent reply.
	// Reply parent can be a thread or another reply.
	if reply.ParentID != thread.ID {
		parent, found := thread.Replies.Get(reply.ParentID)
		if !found {
			return ufmt.Errorf("reply save error, parent reply ID %s not found", reply.ParentID)
		}

		if err := checkPostSupportsReplies(parent); err != nil {
			return err
		}

		parent.Replies.Add(reply)
	}
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
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	lines := strings.SplitN(text, "\n", 2)
	line := lines[0]
	if len(line) > length {
		line = line[:(length-3)] + "..."
	} else if len(lines) > 1 {
		line = line + "..."
	}
	return line
}

func checkPostSupportsReplies(p *Post) error {
	if p.Replies == nil {
		return ufmt.Errorf("replies support not available in post ID %s", p.ID)
	}
	return nil
}
