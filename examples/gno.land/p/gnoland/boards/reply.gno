package boards

import (
	"errors"
	"std"
	"strings"
	"time"
)

// NewReply creates a new reply to a thread or another reply.
func NewReply(parent *Post, creator std.Address, body string) (*Post, error) {
	if parent == nil {
		return nil, errors.New("reply requires a parent thread or reply")
	}

	if !creator.IsValid() {
		return nil, errors.New("invalid reply creator address: " + creator.String())
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errors.New("reply body is required")
	}

	board := parent.Board()
	id := board.generateNextPostID()
	return &Post{
		Body: body,

		id:        id,
		parentID:  parent.ID(),
		threadID:  parent.ThreadID(),
		board:     board,
		creator:   creator,
		createdAt: time.Now(),
	}, nil
}

// MustNewReply creates a new reply or panics on error.
func MustNewReply(parent *Post, creator std.Address, body string) *Post {
	p, err := NewReply(parent, creator, body)
	if err != nil {
		panic(err)
	}
	return p
}

// ThreadID returns the unique identified of the thread where reply was created.
// If post is a thread it returns the unique identifier of the thread, same as ID().
func (reply Post) ThreadID() ID {
	return reply.threadID
}
