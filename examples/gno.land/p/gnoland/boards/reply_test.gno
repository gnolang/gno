package boards_test

import (
	"testing"

	"gno.land/p/nt/urequire"

	"gno.land/p/gnoland/boards"
)

func TestNewReply(t *testing.T) {
	tests := []struct {
		name    string
		parent  func() *boards.Post
		creator address
		body    string
		errMsg  string
	}{
		{
			name: "ok",
			parent: func() *boards.Post {
				board := boards.New(1)
				id := board.ThreadsSequence.Next()
				return &boards.Post{
					ID:       id,
					ThreadID: id,
					Board:    board,
				}
			},
			creator: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			body:    "Foo",
		},
		{
			name:   "nil parent",
			parent: func() *boards.Post { return nil },
			errMsg: "reply requires a parent thread or reply",
		},
		{
			name: "parent without thread ID",
			parent: func() *boards.Post {
				return &boards.Post{ID: 1}
			},
			errMsg: "parent has no thread ID assigned",
		},
		{
			name: "parent without board",
			parent: func() *boards.Post {
				return &boards.Post{ID: 1, ThreadID: 1}
			},
			errMsg: "parent has no board assigned",
		},
		{
			name: "invalid creator",
			parent: func() *boards.Post {
				return &boards.Post{ID: 1, ThreadID: 1, Board: boards.New(1)}
			},
			creator: "foo",
			errMsg:  "invalid reply creator address: foo",
		},
		{
			name: "empty body",
			parent: func() *boards.Post {
				return &boards.Post{ID: 1, ThreadID: 1, Board: boards.New(1)}
			},
			creator: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			body:    "",
			errMsg:  "reply body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := tt.parent()

			reply, err := boards.NewReply(parent, tt.creator, tt.body)

			if tt.errMsg != "" {
				urequire.Error(t, err, "expect an error")
				urequire.ErrorContains(t, err, tt.errMsg, "expect error to match")
				return
			}

			urequire.NoError(t, err, "expect no error")
			urequire.True(t, parent.Board.ThreadsSequence.Last() == reply.ID, "expect ID to match")
			urequire.True(t, parent.ID == reply.ParentID, "expect parent ID to match")
			urequire.True(t, parent.ThreadID == reply.ThreadID, "expect thread ID to match")
			urequire.False(t, reply.Board == nil, "expect board to be assigned")
			urequire.True(t, parent.Board.ID == reply.Board.ID, "expect board ID to match")
			urequire.Equal(t, tt.body, reply.Body, "expect body to match")
			urequire.True(t, reply.Replies != nil, "expect reply to support sub-replies")
			urequire.True(t, reply.Flags != nil, "expect reply to support flagging")
			urequire.Equal(t, tt.creator, reply.Creator, "expect creator to match")
			urequire.False(t, reply.CreatedAt.IsZero(), "expect creation date to be assigned")
		})
	}
}
