package hub

import (
	"gno.land/p/gnoland/boards"
)

// Comment defines a type for threads comment/replies.
type Comment struct {
	// ID is the unique identifier of the comment.
	ID uint64

	// BoardID is the board ID where comment is created.
	BoardID uint64

	// ThreadID contains is the ID of the thread where comment is created.
	ThreadID uint64

	// ParentID is the ID of the parent comment or reply.
	ParentID uint64

	// Body contains the comment's content.
	Body string

	// Hidden indicates that comment is hidden.
	Hidden bool

	// ReplyCount contains the number of comments replies.
	// Count only includes top level replies, sub-replies are not included.
	ReplyCount int

	// FlagCount contains the number of flags that comment has.
	FlagCount int

	// Creator is the account address that created the comment or reply.
	Creator address

	// CreatedAt is thread's creation time as Unix time.
	CreatedAt int64

	// UpdatedAt is thread's update time as Unix time.
	UpdatedAt int64
}

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
		ID:         uint64(ref.ID),
		BoardID:    uint64(ref.Board.ID),
		ThreadID:   uint64(ref.ThreadID),
		ParentID:   uint64(ref.ParentID),
		Body:       ref.Body,
		Hidden:     ref.Hidden,
		ReplyCount: replyCount,
		FlagCount:  flagCount,
		Creator:    ref.Creator,
		CreatedAt:  timeToUnix(ref.CreatedAt),
		UpdatedAt:  timeToUnix(ref.UpdatedAt),
	}
}
