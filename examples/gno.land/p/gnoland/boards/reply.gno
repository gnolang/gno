package boards

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/nt/ufmt"
)

// NewReply creates a new reply to a thread or another reply.
func NewReply(parent *Post, creator address, body string) (*Post, error) {
	if parent == nil {
		return nil, errors.New("reply requires a parent thread or reply")
	}

	if parent.ThreadID == 0 {
		return nil, errors.New("parent has no thread ID assigned")
	}

	if parent.Board == nil {
		return nil, errors.New("parent has no board assigned")
	}

	if !creator.IsValid() {
		return nil, ufmt.Errorf("invalid reply creator address: %s", creator)
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errors.New("reply body is required")
	}

	id := parent.Board.ThreadsSequence.Next()
	return &Post{
		ID:        id,
		ParentID:  parent.ID,
		ThreadID:  parent.ThreadID,
		Board:     parent.Board,
		Body:      body,
		Replies:   NewPostStorage(),
		Flags:     NewFlagStorage(),
		Creator:   creator,
		CreatedAt: time.Now(),
	}, nil
}

// MustNewReply creates a new reply or panics on error.
func MustNewReply(parent *Post, creator address, body string) *Post {
	p, err := NewReply(parent, creator, body)
	if err != nil {
		panic(err)
	}
	return p
}
