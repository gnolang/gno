package boards_test

import (
	"testing"

	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards"
)

func TestPostStorageGet(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() boards.PostStorage
		postID boards.ID
		found  bool
	}{
		{
			name: "single post",
			setup: func() boards.PostStorage {
				s := boards.NewPostStorage()
				s.Add(&boards.Post{ID: 1})
				return s
			},
			postID: 1,
			found:  true,
		},
		{
			name: "multiple posts",
			setup: func() boards.PostStorage {
				s := boards.NewPostStorage()
				s.Add(&boards.Post{ID: 1})
				s.Add(&boards.Post{ID: 2})
				s.Add(&boards.Post{ID: 3})
				return s
			},
			postID: 2,
			found:  true,
		},
		{
			name: "not found",
			setup: func() boards.PostStorage {
				return boards.NewPostStorage()
			},
			postID: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			post, found := s.Get(tt.postID)

			if !tt.found {
				urequire.False(t, found, "expect post not to be found")
				urequire.True(t, post == nil, "expect post to be nil")
				return
			}

			urequire.True(t, found, "expect post to be found")
			urequire.False(t, post == nil, "expect post not to be nil")
			urequire.Equal(t, tt.postID.String(), post.ID.String(), "expect post ID to match")
		})
	}
}

func TestPostStorageRemove(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() boards.PostStorage
		postID  boards.ID
		removed bool
	}{
		{
			name: "ok",
			setup: func() boards.PostStorage {
				s := boards.NewPostStorage()
				s.Add(&boards.Post{ID: 1})
				s.Add(&boards.Post{ID: 2})
				return s
			},
			postID:  2,
			removed: true,
		},
		{
			name: "not found",
			setup: func() boards.PostStorage {
				return boards.NewPostStorage()
			},
			postID: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			post, removed := s.Remove(tt.postID)

			if !tt.removed {
				urequire.False(t, removed, "expect post not to be removed")
				urequire.True(t, post == nil, "expect post to be nil")
				return
			}

			urequire.True(t, removed, "expect post to be removed")
			urequire.False(t, post == nil, "expect post not to be nil")
			urequire.Equal(t, tt.postID.String(), post.ID.String(), "expect post ID to match")

			_, found := s.Get(tt.postID)
			urequire.False(t, found, "expect post not to be found")
		})
	}
}

func TestPostStorageAdd(t *testing.T) {
	tests := []struct {
		name   string
		post   *boards.Post
		errMsg string
	}{
		{
			name: "ok",
			post: &boards.Post{ID: 1},
		},
		{
			name:   "nil post",
			post:   nil,
			errMsg: "saving nil posts is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := boards.NewPostStorage()

			err := s.Add(tt.post)

			if tt.errMsg != "" {
				urequire.Error(t, err, "expect an error")
				urequire.ErrorContains(t, err, tt.errMsg, "expect error to match")
				return
			}

			urequire.NoError(t, err, "expect no error")

			_, found := s.Get(tt.post.ID)
			urequire.True(t, found, "expect post to be found")
		})
	}
}

func TestPostStorageSize(t *testing.T) {
	tests := []struct {
		name  string
		setup func() boards.PostStorage
		size  int
	}{
		{
			name: "empty",
			setup: func() boards.PostStorage {
				return boards.NewPostStorage()
			},
			size: 0,
		},
		{
			name: "one post",
			setup: func() boards.PostStorage {
				s := boards.NewPostStorage()
				s.Add(&boards.Post{ID: 1})
				return s
			},
			size: 1,
		},
		{
			name: "multiple posts",
			setup: func() boards.PostStorage {
				s := boards.NewPostStorage()
				s.Add(&boards.Post{ID: 1})
				s.Add(&boards.Post{ID: 2})
				return s
			},
			size: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			urequire.Equal(t, tt.size, s.Size())
		})
	}
}

func TestPostStorageIterate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() boards.PostStorage
		reverse bool
		ids     []boards.ID
	}{
		{
			name: "default",
			setup: func() boards.PostStorage {
				s := boards.NewPostStorage()
				s.Add(&boards.Post{ID: 1})
				s.Add(&boards.Post{ID: 2})
				s.Add(&boards.Post{ID: 3})
				return s
			},
			ids: []boards.ID{1, 2, 3},
		},
		{
			name: "reverse",
			setup: func() boards.PostStorage {
				s := boards.NewPostStorage()
				s.Add(&boards.Post{ID: 1})
				s.Add(&boards.Post{ID: 2})
				s.Add(&boards.Post{ID: 3})
				return s
			},
			reverse: true,
			ids:     []boards.ID{3, 2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
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
