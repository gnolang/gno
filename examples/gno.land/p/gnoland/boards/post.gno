package boards

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/nt/ufmt"
)

// Post defines a generic type for posts.
// A post can be either a thread or a reply.
type Post struct {
	// ID is the unique identifier of the post.
	ID ID

	// ParentID is the ID of the parent post.
	ParentID ID

	// ThreadID contains the post ID of the thread where current post is created.
	// If current post is a thread it contains post's ID.
	// It should be used when current post is a thread or reply.
	ThreadID ID

	// OriginalBoardID contains the board ID of the original post when current post is a repost.
	OriginalBoardID ID

	//  Board contains the board where post is created.
	Board *Board

	// Title contains the post's title.
	Title string

	// Body contains content of the post.
	Body string

	// Hidden indicates that the post is hidden.
	Hidden bool

	// Readonly indicates that the post is readonly.
	Readonly bool

	// Replies stores post replies.
	Replies PostStorage

	// Reposts stores reposts of the current post.
	// It should be used when post is a thread.
	Reposts RepostStorage

	// Flags stores users flags for the current post.
	Flags FlagStorage

	// Creator is the account address that created the post.
	Creator address

	// CreatedAt is the post's creation time.
	CreatedAt time.Time

	// UpdatedAt is the post's update time.
	UpdatedAt time.Time
}

// Summary return a summary of the post's body.
// It returns the body making sure that the length is limited to 80 characters.
func (post Post) Summary() string {
	return SummaryOf(post.Body, 80)
}

// Save stores the thread (or repost) within it's parent board, or
// if post is a reply it stores the reply within it's parent thread.
func (post *Post) Save() error {
	if post.Board == nil {
		return errors.New("board is required to save a post")
	}

	if post.Board.Threads == nil {
		return ufmt.Errorf("save error, threads are not supported board ID %s", post.Board.ID)
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
	return thread.Board.Threads.Add(thread)
}

func (repost *Post) saveRepost() error {
	thread, found := repost.Board.Threads.Get(repost.ParentID)
	if !found {
		return ufmt.Errorf("repost save error, origin thread ID %s not found", repost.ParentID)
	}

	if thread.Reposts == nil {
		return ufmt.Errorf("repost save error, reposts are not supported in thread ID %s", thread.ID)
	}

	thread.Reposts.Add(repost)
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
		return ufmt.Errorf("replies are not supported in post ID %s", p.ID)
	}
	return nil
}
