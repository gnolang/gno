package boards_test

import (
	"testing"

	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards"
)

func TestNewThread(t *testing.T) {
	tests := []struct {
		name        string
		board       *boards.Board
		creator     address
		title, body string
		errMsg      string
	}{
		{
			name:    "ok",
			board:   boards.New(1),
			creator: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			title:   "Test",
			body:    "Foo",
		},
		{
			name:   "nil board",
			board:  nil,
			errMsg: "thread requires a parent board",
		},
		{
			name:    "invalid creator",
			board:   boards.New(1),
			creator: "foo",
			errMsg:  "invalid thread creator address: foo",
		},
		{
			name:    "empty title",
			board:   boards.New(1),
			creator: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			title:   "",
			errMsg:  "thread title is required",
		},
		{
			name:    "empty body",
			board:   boards.New(1),
			creator: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			title:   "Test",
			body:    "",
			errMsg:  "thread body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thread, err := boards.NewThread(tt.board, tt.creator, tt.title, tt.body)

			if tt.errMsg != "" {
				urequire.Error(t, err, "expect an error")
				urequire.ErrorContains(t, err, tt.errMsg, "expect error to match")
				return
			}

			urequire.NoError(t, err, "expect no error")
			urequire.True(t, tt.board.ThreadsSequence.Last() == thread.ID, "expect ID to match")
			urequire.True(t, thread.ThreadID == thread.ID, "expect thread ID to match")
			urequire.False(t, thread.Board == nil, "expect board to be assigned")
			urequire.True(t, tt.board.ID == thread.Board.ID, "expect board ID to match")
			urequire.Equal(t, tt.title, thread.Title, "expect title to match")
			urequire.Equal(t, tt.body, thread.Body, "expect body to match")
			urequire.True(t, thread.Replies != nil, "expect thread to support sub-replies")
			urequire.True(t, thread.Reposts != nil, "expect thread to support reposts")
			urequire.True(t, thread.Flags != nil, "expect thread to support flagging")
			urequire.Equal(t, tt.creator, thread.Creator, "expect creator to match")
			urequire.False(t, thread.CreatedAt.IsZero(), "expect creation date to be assigned")
		})
	}
}

func TestNewRepost(t *testing.T) {
	tests := []struct {
		name       string
		origThread *boards.Post
		dstBoard   *boards.Board
		creator    address
		errMsg     string
	}{
		{
			name: "ok",
			origThread: boards.MustNewThread(
				boards.New(1),
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"Title",
				"Body",
			),
			dstBoard: boards.New(2),
			creator:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		},
		{
			name:       "nil original thread",
			origThread: nil,
			errMsg:     "thread to repost is required",
		},
		{
			name:       "original thread without board",
			origThread: &boards.Post{ID: 1},
			errMsg:     "original thread has no board assigned",
		},
		{
			name: "nil destination board",
			origThread: boards.MustNewThread(
				boards.New(1),
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"Title",
				"Body",
			),
			dstBoard: nil,
			errMsg:   "thread repost requires a destination board",
		},
		{
			name: "original thread is not a thread",
			origThread: &boards.Post{
				ID:       1,
				ThreadID: 2,
				Board:    boards.New(1),
			},
			dstBoard: boards.New(2),
			errMsg:   "post must be a thread to be reposted to another board",
		},
		{
			name: "original thread is a repost",
			origThread: &boards.Post{
				ID:              1,
				ThreadID:        1,
				OriginalBoardID: 1,
				Board:           boards.New(1),
			},
			dstBoard: boards.New(2),
			errMsg:   "reposting a thread that is a repost is not allowed",
		},
		{
			name: "invalid creator",
			origThread: boards.MustNewThread(
				boards.New(1),
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"Title",
				"Body",
			),
			dstBoard: boards.New(2),
			creator:  "foo",
			errMsg:   "invalid thread repost creator address: foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thread, err := boards.NewRepost(tt.origThread, tt.dstBoard, tt.creator)

			if tt.errMsg != "" {
				urequire.Error(t, err, "expect an error")
				urequire.ErrorContains(t, err, tt.errMsg, "expect error to match")
				return
			}

			urequire.NoError(t, err, "expect no error")
			urequire.True(t, thread.ID == tt.dstBoard.ThreadsSequence.Last(), "expect ID to match")
			urequire.True(t, thread.ThreadID == thread.ID, "expect thread ID to match")
			urequire.True(t, thread.ParentID == tt.origThread.ID, "expect parent ID to match")
			urequire.True(t, thread.OriginalBoardID == tt.origThread.Board.ID, "expect original board ID to match")
			urequire.False(t, thread.Board == nil, "expect board to be assigned")
			urequire.True(t, thread.Board.ID == tt.dstBoard.ID, "expect board ID to match")
			urequire.Empty(t, thread.Title, "expect title to be empty")
			urequire.Empty(t, thread.Body, "expect body to be empty")
			urequire.True(t, thread.Replies != nil, "expect thread to support sub-replies")
			urequire.True(t, thread.Reposts != nil, "expect thread to support reposts")
			urequire.True(t, thread.Flags != nil, "expect thread to support flagging")
			urequire.Equal(t, tt.creator, thread.Creator, "expect creator to match")
			urequire.False(t, thread.CreatedAt.IsZero(), "expect creation date to be assigned")
		})
	}
}
