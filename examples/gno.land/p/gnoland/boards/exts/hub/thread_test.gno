package hub_test

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/hub"
	"gno.land/p/nt/urequire/v0"
)

func TestNewSafeThread(t *testing.T) {
	board := boards.New(1)
	ref := boards.MustNewThread(board, alice, "Title", "Body")
	ref.Flags.Add(boards.Flag{User: bob, Reason: "Spam"})

	thread := hub.NewSafeThread(ref)

	urequire.Equal(t, uint64(ref.ID), thread.ID(), "expect ID to match")
	urequire.Equal(t, uint64(1), thread.BoardID(), "expect board ID to match")
	urequire.Equal(t, "Title", thread.Title(), "expect title to match")
	urequire.Equal(t, "Body", thread.Body(), "expect body to match")
	urequire.Equal(t, alice, thread.Creator(), "expect creator to match")
	urequire.False(t, thread.Hidden(), "expect thread not to be hidden")
	urequire.False(t, thread.Readonly(), "expect thread not to be readonly")
	urequire.Equal(t, 0, thread.CommentCount(), "expect no comments")
	urequire.Equal(t, 0, thread.RepostCount(), "expect no reposts")
	urequire.Equal(t, 1, thread.FlagCount(), "expect one flag")
	urequire.Equal(t, testTime, thread.CreatedAt(), "expect creation time to match")
	urequire.Equal(t, int64(0), thread.UpdatedAt(), "expect zero update time to map to zero")
}

func TestNewSafeThreadCounts(t *testing.T) {
	board := boards.New(1)
	ref := boards.MustNewThread(board, alice, "Title", "Body")
	ref.Replies.Add(boards.MustNewReply(ref, bob, "Comment 1"))
	ref.Replies.Add(boards.MustNewReply(ref, bob, "Comment 2"))

	thread := hub.NewSafeThread(ref)

	urequire.Equal(t, 2, thread.CommentCount(), "expect comment count to match")
	urequire.Equal(t, 0, thread.FlagCount(), "expect no flags")
}

// TestNewSafeThreadNilStorages covers the branches that skip counting when a
// post carries no reply, repost or flag storage.
func TestNewSafeThreadNilStorages(t *testing.T) {
	ref := &boards.Post{ID: 1, ThreadID: 1, Board: boards.New(1)}

	thread := hub.NewSafeThread(ref)

	urequire.Equal(t, 0, thread.CommentCount(), "expect no comments")
	urequire.Equal(t, 0, thread.RepostCount(), "expect no reposts")
	urequire.Equal(t, 0, thread.FlagCount(), "expect no flags")
	urequire.Equal(t, int64(0), thread.CreatedAt(), "expect zero creation time to map to zero")
}

func TestNewSafeThreadRejectsInvalidRefs(cur realm, t *testing.T) {
	board := boards.New(1)
	thread := boards.MustNewThread(board, alice, "Title", "Body")
	comment := boards.MustNewReply(thread, alice, "Comment")

	urequire.PanicsWithMessage(t, cur, "post reference is nil", func() {
		hub.NewSafeThread(nil)
	}, "expect a nil post reference to be rejected")

	urequire.PanicsWithMessage(t, cur, "post is not a thread", func() {
		hub.NewSafeThread(comment)
	}, "expect a comment to be rejected")
}

func TestNewSafeFlag(t *testing.T) {
	flag := hub.NewSafeFlag(boards.Flag{User: alice, Reason: "Spam"})

	urequire.Equal(t, alice, flag.User(), "expect user to match")
	urequire.Equal(t, "Spam", flag.Reason(), "expect reason to match")
}
