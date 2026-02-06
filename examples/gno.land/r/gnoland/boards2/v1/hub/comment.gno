package hub

import (
	"time"

	"gno.land/p/gnoland/boards"
)

// Comment defines a type for threads comment/replies.
type Comment struct {
	ref *boards.Post

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
	// Total only includes top level replies, sub-replies are not included.
	ReplyCount int

	// FlagCount contains the number of flags that comment has.
	FlagCount int

	// Creator is the account address that created the comment or reply.
	Creator address

	// CreatedAt is thread's creation time.
	CreatedAt time.Time

	// UpdatedAt is thread's update time.
	UpdatedAt time.Time
}

// IterateFlags iterates comment moderation flags.
// To reverse iterate use a negative count.
func (c Comment) IterateFlags(start, count int, fn func(boards.Flag) bool) bool {
	return c.ref.Flags.Iterate(start, count, fn)
}

// IterateReplies iterates comment replies (sub-comments).
// Only top level replies are iterated, sub-replies are ignored.
// To reverse iterate use a negative count.
func (t Thread) IterateReplies(start, count int, fn func(Comment) bool) bool {
	return t.ref.Replies.Iterate(start, count, func(comment *boards.Post) bool {
		return fn(newSafeComment(comment))
	})
}

func newSafeComment(ref *boards.Post) Comment {
	if boards.IsThread(ref) {
		panic("post is not a comment or reply")
	}

	return Comment{
		ref:        ref,
		ID:         uint64(ref.ID),
		BoardID:    uint64(ref.Board.ID),
		ThreadID:   uint64(ref.ThreadID),
		ParentID:   uint64(ref.ParentID),
		Body:       ref.Body,
		Hidden:     ref.Hidden,
		ReplyCount: ref.Replies.Size(),
		FlagCount:  ref.Flags.Size(),
		Creator:    ref.Creator,
		CreatedAt:  ref.CreatedAt,
		UpdatedAt:  ref.UpdatedAt,
	}
}
