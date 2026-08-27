package hub_test

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/hub"
	"gno.land/p/nt/urequire/v0"
)

// Addresses and the fixed time the test VM reports for time.Now().
const (
	alice    address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	bob      address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	testTime int64   = 1234567890
)

func TestNewSafeComment(t *testing.T) {
	board := boards.New(1)
	thread := boards.MustNewThread(board, alice, "Title", "Body")
	ref := boards.MustNewReply(thread, bob, "Comment")
	ref.Flags.Add(boards.Flag{User: alice, Reason: "Spam"})

	comment := hub.NewSafeComment(ref)

	urequire.Equal(t, uint64(ref.ID), comment.ID(), "expect ID to match")
	urequire.Equal(t, uint64(1), comment.BoardID(), "expect board ID to match")
	urequire.Equal(t, uint64(thread.ID), comment.ThreadID(), "expect thread ID to match")
	urequire.Equal(t, "Comment", comment.Body(), "expect body to match")
	urequire.Equal(t, bob, comment.Creator(), "expect creator to match")
	urequire.False(t, comment.Hidden(), "expect comment not to be hidden")
	urequire.Equal(t, 0, comment.ReplyCount(), "expect no replies")
	urequire.Equal(t, 1, comment.FlagCount(), "expect one flag")
	urequire.Equal(t, testTime, comment.CreatedAt(), "expect creation time to match")
	urequire.Equal(t, int64(0), comment.UpdatedAt(), "expect zero update time to map to zero")
}

// TestNewSafeCommentParentID documents that a top level comment reports its
// thread as parent rather than a zero parent, so callers must compare
// ParentID against ThreadID to tell comments and replies apart.
func TestNewSafeCommentParentID(t *testing.T) {
	board := boards.New(1)
	thread := boards.MustNewThread(board, alice, "Title", "Body")
	commentRef := boards.MustNewReply(thread, bob, "Comment")
	replyRef := boards.MustNewReply(commentRef, alice, "Reply")

	comment := hub.NewSafeComment(commentRef)
	reply := hub.NewSafeComment(replyRef)

	urequire.Equal(t, comment.ThreadID(), comment.ParentID(), "expect a comment's parent to be its thread")
	urequire.Equal(t, comment.ID(), reply.ParentID(), "expect a reply's parent to be its comment")
	urequire.Equal(t, comment.ThreadID(), reply.ThreadID(), "expect a reply to keep the thread ID")
}

func TestNewSafeCommentReplyCount(t *testing.T) {
	board := boards.New(1)
	thread := boards.MustNewThread(board, alice, "Title", "Body")
	ref := boards.MustNewReply(thread, bob, "Comment")
	ref.Replies.Add(boards.MustNewReply(ref, alice, "Reply 1"))
	ref.Replies.Add(boards.MustNewReply(ref, alice, "Reply 2"))

	comment := hub.NewSafeComment(ref)

	urequire.Equal(t, 2, comment.ReplyCount(), "expect reply count to match")
}

// TestNewSafeCommentNilStorages covers the branches that skip counting when a
// post carries no reply or flag storage.
func TestNewSafeCommentNilStorages(t *testing.T) {
	ref := &boards.Post{ID: 2, ParentID: 1, ThreadID: 1, Board: boards.New(1)}

	comment := hub.NewSafeComment(ref)

	urequire.Equal(t, 0, comment.ReplyCount(), "expect no replies")
	urequire.Equal(t, 0, comment.FlagCount(), "expect no flags")
	urequire.Equal(t, int64(0), comment.CreatedAt(), "expect zero creation time to map to zero")
}

func TestNewSafeCommentRejectsInvalidRefs(cur realm, t *testing.T) {
	board := boards.New(1)
	thread := boards.MustNewThread(board, alice, "Title", "Body")

	urequire.PanicsWithMessage(t, cur, "post reference is nil", func() {
		hub.NewSafeComment(nil)
	}, "expect a nil post reference to be rejected")

	urequire.PanicsWithMessage(t, cur, "post is not a comment or reply", func() {
		hub.NewSafeComment(thread)
	}, "expect a thread to be rejected")
}
