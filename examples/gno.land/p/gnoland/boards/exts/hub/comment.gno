package hub

import (
	"gno.land/p/gnoland/boards"
)

// Comment defines a type for threads comment/replies.
type Comment struct {
	// id is the unique identifier of the comment.
	id uint64

	// boardID is the board ID where comment is created.
	boardID uint64

	// threadID is the ID of the thread where comment is created.
	threadID uint64

	// parentID is the ID of the parent comment or reply.
	parentID uint64

	// body contains the comment's content.
	body string

	// hidden indicates that comment is hidden.
	hidden bool

	// replyCount contains the number of comments replies.
	// Count only includes top level replies, sub-replies are not included.
	replyCount int

	// flagCount contains the number of flags that comment has.
	flagCount int

	// creator is the account address that created the comment or reply.
	creator address

	// createdAt is thread's creation time as Unix time.
	createdAt int64

	// updatedAt is thread's update time as Unix time.
	updatedAt int64
}

// ID returns the unique identifier of the comment.
func (c Comment) ID() uint64 { return c.id }

// BoardID returns the board ID where the comment is created.
func (c Comment) BoardID() uint64 { return c.boardID }

// ThreadID returns the ID of the thread where the comment is created.
func (c Comment) ThreadID() uint64 { return c.threadID }

// ParentID returns the ID of the parent comment or reply.
func (c Comment) ParentID() uint64 { return c.parentID }

// Body returns the comment's content.
func (c Comment) Body() string { return c.body }

// Hidden indicates that the comment is hidden.
func (c Comment) Hidden() bool { return c.hidden }

// ReplyCount returns the number of comment replies.
// Count only includes top level replies, sub-replies are not included.
func (c Comment) ReplyCount() int { return c.replyCount }

// FlagCount returns the number of flags that the comment has.
func (c Comment) FlagCount() int { return c.flagCount }

// Creator returns the account address that created the comment or reply.
func (c Comment) Creator() address { return c.creator }

// CreatedAt returns the comment's creation time as Unix time.
func (c Comment) CreatedAt() int64 { return c.createdAt }

// UpdatedAt returns the comment's update time as Unix time.
func (c Comment) UpdatedAt() int64 { return c.updatedAt }

// NewSafeComment creates a safe comment.
func NewSafeComment(ref *boards.Post) Comment {
	if boards.IsThread(ref) {
		panic("post is not a comment or reply")
	}

	var replyCount int
	if ref.Replies != nil {
		replyCount = ref.Replies.Size()
	}

	var flagCount int
	if ref.Flags != nil {
		flagCount = ref.Flags.Size()
	}

	return Comment{
		id:         uint64(ref.ID),
		boardID:    uint64(ref.Board.ID),
		threadID:   uint64(ref.ThreadID),
		parentID:   uint64(ref.ParentID),
		body:       ref.Body,
		hidden:     ref.Hidden,
		replyCount: replyCount,
		flagCount:  flagCount,
		creator:    ref.Creator,
		createdAt:  timeToUnix(ref.CreatedAt),
		updatedAt:  timeToUnix(ref.UpdatedAt),
	}
}
