package boards

import (
	"strings"
	"time"
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

	// Board contains the board where post is created.
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

	// Meta allows storing post metadata.
	Meta any

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

// SetID sets post ID value.
func (post *Post) SetID(v ID) {
	post.ID = v
}

// SetParentID sets post's parent ID value.
func (post *Post) SetParentID(v ID) {
	post.ParentID = v
}

// SetThreadID sets thread ID value.
func (post *Post) SetThreadID(v ID) {
	post.ThreadID = v
}

// SetOriginalBoardID sets the board ID of the original post when current post is a repost.
func (post *Post) SetOriginalBoardID(v ID) {
	post.OriginalBoardID = v
}

// SetBoard sets the board where post was created.
func (post *Post) SetBoard(v *Board) {
	post.Board = v
}

// SetTitle sets title value.
func (post *Post) SetTitle(v string) {
	post.Title = v
}

// SetBody sets post's content.
func (post *Post) SetBody(v string) {
	post.Body = v
}

// SetHidden sets hidden value.
func (post *Post) SetHidden(v bool) {
	post.Hidden = v
}

// SetReadonly sets readonly value.
func (post *Post) SetReadonly(v bool) {
	post.Readonly = v
}

// SetReplyStorage sets the storage where post replies are stored.
func (post *Post) SetReplyStorage(v PostStorage) {
	post.Replies = v
}

// SetRepostStorage sets the storage where thread reposts are stored.
func (post *Post) SetRepostStorage(v RepostStorage) {
	post.Reposts = v
}

// SetFlagStorage sets the storage where post flags are stored.
func (post *Post) SetFlagStorage(v FlagStorage) {
	post.Flags = v
}

// SetCreator sets the address of the account that created the post.
func (post *Post) SetCreator(v address) {
	post.Creator = v
}

// SetCreatedAt sets the time when post was created.
func (post *Post) SetCreatedAt(v time.Time) {
	post.CreatedAt = v
}

// SetUpdatedAt sets the time when a post value was updated.
func (post *Post) SetUpdatedAt(v time.Time) {
	post.UpdatedAt = v
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
