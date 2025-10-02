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

	if !IsThread(thread) {
		return nil, errors.New("post must be a thread to be reposted to another board")
	}

	if IsRepost(thread) {
		return nil, errors.New("reposting a thread that is a repost is not allowed")
	}

	if dst == nil {
		return nil, errors.New("thread repost requires a destination board")
	}

	if !creator.IsValid() {
		return nil, ufmt.Errorf("invalid thread repost creator address: %s", creator)
	}

	id := dst.ThreadsSequence.Next()
	return &Post{
		ID:              id,
		ThreadID:        id,
		Board:           dst,
		ParentID:        thread.ID,       // TODO: Check
		OriginalBoardID: thread.Board.ID, // TODO: Check
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

// DeleteReply deletes a reply from a thread.
func (thread *Post) DeleteReply(replyID ID) error {
	if !IsThread(thread) {
		// TODO: Allow removing replies from parent replies too
		return errors.New("post must be a thread to remove a reply")
	}

	if err := checkPostSupportsReplies(thread); err != nil {
		return err
	}

	// Remove reply from the thread
	reply, removed := thread.Replies.Remove(replyID)
	if !removed {
		return errors.New("reply not found")
	}

	// Remove reply from parent reply
	if reply.ParentID != thread.ID {
		parent, found := thread.Replies.Get(reply.ParentID)
		if !found {
			return ufmt.Errorf("parent reply ID %s not found", reply.ParentID)
		}

		if err := checkPostSupportsReplies(parent); err != nil {
			return err
		}

		parent.Replies.Remove(replyID)
	}
	return nil
}
