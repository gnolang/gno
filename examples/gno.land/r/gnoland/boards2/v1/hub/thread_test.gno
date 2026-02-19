package hub_test

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/urequire"

	"gno.land/r/gnoland/boards2/v1/hub"
)

func TestThreadIterateFlags(t *testing.T) {
	tests := []struct {
		name  string
		flags []boards.Flag
	}{
		{
			name: "no flags",
		},
		{
			name: "one flag",
			flags: []boards.Flag{
				{User: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Reason: "Reason 1"},
			},
		},
		{
			name: "multiple flags",
			flags: []boards.Flag{
				{User: "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq", Reason: "Reason 1"},
				{User: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Reason: "Reason 2"},
				{User: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", Reason: "Reason 3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i int
			ref := createComment(t, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Comment")
			for _, f := range tt.flags {
				ref.Flags.Add(f)
			}

			comment := hub.NewSafeComment(ref)

			urequire.Equal(t, len(tt.flags), comment.FlagCount, "expect number of flags to match")
			comment.IterateFlags(0, comment.FlagCount, func(f boards.Flag) bool {
				urequire.Equal(t, tt.flags[i].User, f.User, "expect user to match")
				urequire.Equal(t, tt.flags[i].Reason, f.Reason, "expect reason to match")
				i++
				return false
			})
		})
	}
}

func TestThreadIterateReposts(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*boards.Post)
		expected [][2]uint64 // {boardID, repostThreadID}
	}{
		{
			name: "no reposts",
		},
		{
			name: "one repost",
			setup: func(thread *boards.Post) {
				r := boards.MustNewRepost(thread, boards.New(1), "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				thread.Reposts.Add(r)
			},
			expected: [][2]uint64{
				{1, 1},
			},
		},
		{
			name: "multiple reposts",
			setup: func(thread *boards.Post) {
				reposts := []*boards.Post{
					boards.MustNewRepost(thread, boards.New(1), "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
					boards.MustNewRepost(thread, boards.New(2), "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"),
					boards.MustNewRepost(thread, boards.New(5), "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"),
				}
				for _, r := range reposts {
					thread.Reposts.Add(r)
				}
			},
			expected: [][2]uint64{
				{1, 1},
				{2, 1},
				{5, 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i int
			ref := createThread(t, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
			if tt.setup != nil {
				tt.setup(ref)
			}

			thread := hub.NewSafeThread(ref)

			urequire.Equal(t, ref.Reposts.Size(), thread.RepostCount, "expect number of reposts to match")
			thread.IterateReposts(0, thread.RepostCount, func(boardID, repostThreadID uint64) bool {
				urequire.Equal(t, tt.expected[i][0], boardID, "expect repost board ID to match")
				urequire.Equal(t, tt.expected[i][1], repostThreadID, "expect repost thread ID to match")
				i++
				return false
			})
		})
	}
}

func TestThreadIterateComments(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*boards.Post)
	}{
		{
			name: "no comments",
		},
		{
			name: "one comment",
			setup: func(thread *boards.Post) {
				r := boards.MustNewReply(thread, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Body 1")
				thread.Replies.Add(r)
			},
		},
		{
			name: "multiple comments",
			setup: func(thread *boards.Post) {
				replies := []*boards.Post{
					boards.MustNewReply(thread, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Body 1"),
					boards.MustNewReply(thread, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", "Body 2"),
					boards.MustNewReply(thread, "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq", "Body 3"),
				}
				for _, t := range replies {
					thread.Replies.Add(t)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := createThread(t, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
			if tt.setup != nil {
				tt.setup(ref)
			}

			thread := hub.NewSafeThread(ref)

			urequire.Equal(t, ref.Replies.Size(), thread.CommentCount, "expect number of replies to match")
			thread.IterateComments(0, thread.CommentCount, func(comment hub.Comment) bool {
				id := boards.ID(comment.ID)
				expected, found := ref.Replies.Get(id)

				urequire.True(t, found, "expect comment to be found")
				urequire.Equal(t, expected.Creator, comment.Creator, "expect creator to match")
				urequire.Equal(t, expected.Body, comment.Body, "expect body to match")
				return false
			})
		})
	}
}

func createThread(t *testing.T, user address, title, body string) *boards.Post {
	t.Helper()

	board := boards.New(1)
	return boards.MustNewThread(board, user, title, body)
}
