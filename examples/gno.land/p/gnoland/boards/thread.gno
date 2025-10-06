package boards

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/nt/ufmt"
)

// NewThread creates a new board thread.
func NewThread(b *Board, creator address, title, body string) (*Post, error) {
	if b == nil {
		return nil, errors.New("thread requires a parent board")
	}

	if !creator.IsValid() {
		return nil, ufmt.Errorf("invalid thread creator address: %s", creator)
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("thread title is required")
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errors.New("thread body is required")
	}

	id := b.ThreadsSequence.Next()
	return &Post{
		ID:        id,
		ThreadID:  id,
		Board:     b,
		Title:     title,
		Body:      body,
		Replies:   NewPostStorage(),
		Reposts:   NewRepostStorage(),
		Flags:     NewFlagStorage(),
		Creator:   creator,
		CreatedAt: time.Now(),
	}, nil
}

// MustNewThread creates a new thread or panics on error.
func MustNewThread(b *Board, creator address, title, body string) *Post {
	t, err := NewThread(b, creator, title, body)
	if err != nil {
		panic(err)
	}
	return t
}

// NewRepost creates a new thread that is a repost of a thread from another board.
func NewRepost(thread *Post, dst *Board, creator address) (*Post, error) {
	if thread == nil {
		return nil, errors.New("thread to repost is required")
	}

	if thread.Board == nil {
		return nil, errors.New("original thread has no board assigned")
	}

	if dst == nil {
		return nil, errors.New("thread repost requires a destination board")
	}

	if IsRepost(thread) {
		return nil, errors.New("reposting a thread that is a repost is not allowed")
	}

	if !IsThread(thread) {
		return nil, errors.New("post must be a thread to be reposted to another board")
	}

	if !creator.IsValid() {
		return nil, ufmt.Errorf("invalid thread repost creator address: %s", creator)
	}

	id := dst.ThreadsSequence.Next()
	return &Post{
		ID:              id,
		ThreadID:        id,
		ParentID:        thread.ID,
		OriginalBoardID: thread.Board.ID,
		Board:           dst,
		Replies:         NewPostStorage(),
		Reposts:         NewRepostStorage(),
		Flags:           NewFlagStorage(),
		Creator:         creator,
		CreatedAt:       time.Now(),
	}, nil
}

// MustNewRepost creates a new thread that is a repost of a thread from another board or panics on error.
func MustNewRepost(thread *Post, dst *Board, creator address) *Post {
	r, err := NewRepost(thread, dst, creator)
	if err != nil {
		panic(err)
	}
	return r
}
