package boards2_test

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/urequire/v0"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

func TestReplyStorageGet(t *testing.T) {
	creator := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	tests := []struct {
		name   string
		setup  func(thread *boards.Post) boards.PostStorage
		postID boards.ID
		found  bool
	}{
		{
			name: "single reply",
			setup: func(t *boards.Post) boards.PostStorage {
				s := boards2.NewReplyStorage()
				s.Add(boards.MustNewReply(t, creator, "A")) // ID=2
				return s
			},
			postID: 2,
			found:  true,
		},
		{
			name: "multiple replies",
			setup: func(t *boards.Post) boards.PostStorage {
				s := boards2.NewReplyStorage()
				s.Add(boards.MustNewReply(t, creator, "A"))
				s.Add(boards.MustNewReply(t, creator, "B")) // ID=3
				s.Add(boards.MustNewReply(t, creator, "C"))
				return s
			},
			postID: 3,
			found:  true,
		},
		{
			name: "single sub reply",
			setup: func(t *boards.Post) boards.PostStorage {
				parent := boards.MustNewReply(t, creator, "A")
				s := boards2.NewReplyStorage()
				s.Add(parent)
				s.Add(boards.MustNewReply(parent, creator, "B")) // ID=3
				return s
			},
			postID: 3,
			found:  true,
		},
		{
			name: "multiple sub replies",
			setup: func(t *boards.Post) boards.PostStorage {
				parent := boards.MustNewReply(t, creator, "A")
				s := boards2.NewReplyStorage()
				s.Add(parent)
				s.Add(boards.MustNewReply(parent, creator, "C"))
				s.Add(boards.MustNewReply(parent, creator, "D")) // ID=4
				s.Add(boards.MustNewReply(parent, creator, "E"))
				return s
			},
			postID: 4,
			found:  true,
		},
		{
			name: "not found",
			setup: func(*boards.Post) boards.PostStorage {
				return boards2.NewReplyStorage()
			},
			postID: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := boards.New(1)
			thread := boards.MustNewThread(board, creator, "Title", "Body")
			s := tt.setup(thread)

			reply, found := s.Get(tt.postID)

			if !tt.found {
				urequire.False(t, found, "expect reply not to be found")
				urequire.True(t, reply == nil, "expect reply to be nil")
				return
			}

			urequire.True(t, found, "expect reply to be found")
			urequire.False(t, reply == nil, "expect reply not to be nil")
			urequire.Equal(t, tt.postID.String(), reply.ID.String(), "expect reply ID to match")
		})
	}
}

func TestReplyStorageRemove(t *testing.T) {
	creator := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	tests := []struct {
		name    string
		setup   func(thread *boards.Post) boards.PostStorage
		postID  boards.ID
		removed bool
	}{
		{
			name: "reply",
			setup: func(t *boards.Post) boards.PostStorage {
				s := boards2.NewReplyStorage()
				s.Add(boards.MustNewReply(t, creator, "A"))
				s.Add(boards.MustNewReply(t, creator, "B")) // ID=3
				return s
			},
			postID:  3,
			removed: true,
		},
		{
			name: "sub reply",
			setup: func(t *boards.Post) boards.PostStorage {
				parent := boards.MustNewReply(t, creator, "A")
				s := boards2.NewReplyStorage()
				s.Add(parent)
				s.Add(boards.MustNewReply(parent, creator, "A"))
				s.Add(boards.MustNewReply(parent, creator, "B")) // ID=4
				return s
			},
			postID:  4,
			removed: true,
		},
		{
			name: "not found",
			setup: func(*boards.Post) boards.PostStorage {
				return boards2.NewReplyStorage()
			},
			postID: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := boards.New(1)
			thread := boards.MustNewThread(board, creator, "Title", "Body")
			s := tt.setup(thread)

			reply, removed := s.Remove(tt.postID)

			if !tt.removed {
				urequire.False(t, removed, "expect reply not to be removed")
				urequire.True(t, reply == nil, "expect reply to be nil")
				return
			}

			urequire.True(t, removed, "expect reply to be removed")
			urequire.False(t, reply == nil, "expect reply not to be nil")
			urequire.Equal(t, tt.postID.String(), reply.ID.String(), "expect reply ID to match")

			_, found := s.Get(tt.postID)
			urequire.False(t, found, "expect reply not to be found")
		})
	}
}

func TestReplyStorageAdd(t *testing.T) {
	tests := []struct {
		name   string
		post   *boards.Post
		errMsg string
	}{
		{
			name: "reply",
			post: &boards.Post{ID: 2, ParentID: 1, ThreadID: 1},
		},
		{
			name: "sub reply",
			post: &boards.Post{ID: 3, ParentID: 2, ThreadID: 1},
		},
		{
			name:   "nil reply",
			post:   nil,
			errMsg: "saving nil replies is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := boards2.NewReplyStorage()

			err := s.Add(tt.post)

			if tt.errMsg != "" {
				urequire.Error(t, err, "expect an error")
				urequire.ErrorContains(t, err, tt.errMsg, "expect error to match")
				return
			}

			urequire.NoError(t, err, "expect no error")

			_, found := s.Get(tt.post.ID)
			urequire.True(t, found, "expect reply to be found")
		})
	}
}

func TestReplyStorageSize(t *testing.T) {
	creator := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	tests := []struct {
		name  string
		setup func(thread *boards.Post) boards.PostStorage
		size  int
	}{
		{
			name: "empty",
			setup: func(*boards.Post) boards.PostStorage {
				return boards2.NewReplyStorage()
			},
			size: 0,
		},
		{
			name: "one reply",
			setup: func(t *boards.Post) boards.PostStorage {
				s := boards2.NewReplyStorage()
				s.Add(boards.MustNewReply(t, creator, "A"))
				return s
			},
			size: 1,
		},
		{
			name: "multiple replies",
			setup: func(t *boards.Post) boards.PostStorage {
				s := boards2.NewReplyStorage()
				s.Add(boards.MustNewReply(t, creator, "A"))
				s.Add(boards.MustNewReply(t, creator, "B"))
				return s
			},
			size: 2,
		},
		{
			name: "multiple replies and sub reply",
			setup: func(t *boards.Post) boards.PostStorage {
				parent := boards.MustNewReply(t, creator, "A")
				s := boards2.NewReplyStorage()
				s.Add(parent)
				s.Add(boards.MustNewReply(t, creator, "B"))
				s.Add(boards.MustNewReply(parent, creator, "A2"))
				return s
			},
			size: 2, // Sub-replies are not counted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := boards.New(1)
			thread := boards.MustNewThread(board, creator, "Title", "Body")
			s := tt.setup(thread)

			urequire.Equal(t, tt.size, s.Size())
		})
	}
}

func TestReplyStorageIterate(t *testing.T) {
	creator := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	tests := []struct {
		name    string
		setup   func(thread *boards.Post) boards.PostStorage
		reverse bool
		ids     []boards.ID
	}{
		{
			name: "default",
			setup: func(t *boards.Post) boards.PostStorage {
				parent := boards.MustNewReply(t, creator, "A")
				s := boards2.NewReplyStorage()
				s.Add(parent)
				s.Add(boards.MustNewReply(t, creator, "B"))
				s.Add(boards.MustNewReply(t, creator, "C"))
				s.Add(boards.MustNewReply(parent, creator, "A2")) // Sub-replies are ignored
				return s
			},
			ids: []boards.ID{2, 3, 4},
		},
		{
			name: "reverse",
			setup: func(t *boards.Post) boards.PostStorage {
				parent := boards.MustNewReply(t, creator, "A")
				s := boards2.NewReplyStorage()
				s.Add(parent)
				s.Add(boards.MustNewReply(t, creator, "B"))
				s.Add(boards.MustNewReply(t, creator, "C"))
				s.Add(boards.MustNewReply(parent, creator, "A2")) // Sub-replies are ignored
				return s
			},
			reverse: true,
			ids:     []boards.ID{4, 3, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := boards.New(1)
			thread := boards.MustNewThread(board, creator, "Title", "Body")
			s := tt.setup(thread)

			count := s.Size()
			if tt.reverse {
				count = -count
			}

			var i int
			s.Iterate(0, count, func(p *boards.Post) bool {
				urequire.True(t, tt.ids[i] == p.ID, "expect post ID to match")

				i++
				return false
			})
		})
	}
}
