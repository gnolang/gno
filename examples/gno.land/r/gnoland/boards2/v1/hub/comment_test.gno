package hub_test

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/urequire"

	"gno.land/r/gnoland/boards2/v1/hub"
)

func TestCommentIterateFlags(t *testing.T) {
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

func TestCommentIterateReplies(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*boards.Post)
	}{
		{
			name: "no replies",
		},
		{
			name: "one reply",
			setup: func(comment *boards.Post) {
				r := boards.MustNewReply(comment, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Body 1")
				comment.Replies.Add(r)
			},
		},
		{
			name: "multiple replies",
			setup: func(comment *boards.Post) {
				replies := []*boards.Post{
					boards.MustNewReply(comment, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Body 1"),
					boards.MustNewReply(comment, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", "Body 2"),
					boards.MustNewReply(comment, "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq", "Body 3"),
				}
				for _, t := range replies {
					comment.Replies.Add(t)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := createComment(t, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Comment")
			if tt.setup != nil {
				tt.setup(ref)
			}

			comment := hub.NewSafeComment(ref)

			urequire.Equal(t, ref.Replies.Size(), comment.ReplyCount, "expect number of replies to match")
			comment.IterateReplies(0, comment.ReplyCount, func(reply hub.Comment) bool {
				id := boards.ID(reply.ID)
				expected, found := ref.Replies.Get(id)

				urequire.True(t, found, "expect reply to be found")
				urequire.Equal(t, expected.Creator, reply.Creator, "expect creator to match")
				urequire.Equal(t, expected.Body, reply.Body, "expect body to match")
				return false
			})
		})
	}
}

func createComment(t *testing.T, user address, body string) *boards.Post {
	t.Helper()

	thread := createThread(t, user, "Title", "Body")
	return boards.MustNewReply(thread, user, body)
}
