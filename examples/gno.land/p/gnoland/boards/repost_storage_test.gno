package boards_test

import (
	"testing"

	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards"
)

func TestRepostStorageGet(t *testing.T) {
	tests := []struct {
		name              string
		setup             func() boards.RepostStorage
		boardID, repostID boards.ID
		found             bool
	}{
		{
			name: "single repost",
			setup: func() boards.RepostStorage {
				s := boards.NewRepostStorage()
				s.Add(&boards.Post{
					ID:    1,
					Board: &boards.Board{ID: 1},
				})
				return s
			},
			boardID:  1,
			repostID: 1,
			found:    true,
		},
		{
			name: "multiple reposts",
			setup: func() boards.RepostStorage {
				s := boards.NewRepostStorage()
				s.Add(&boards.Post{
					ID:    2,
					Board: &boards.Board{ID: 1},
				})
				s.Add(&boards.Post{
					ID:    5,
					Board: &boards.Board{ID: 2},
				})
				s.Add(&boards.Post{
					ID:    10,
					Board: &boards.Board{ID: 3},
				})
				return s
			},
			boardID:  1,
			repostID: 2,
			found:    true,
		},
		{
			name: "not found",
			setup: func() boards.RepostStorage {
				return boards.NewRepostStorage()
			},
			boardID: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			repostID, found := s.Get(tt.boardID)

			if !tt.found {
				urequire.False(t, found, "expect repost not to be found")
				urequire.True(t, int(repostID) == 0, "expect repost ID to be 0")
				return
			}

			urequire.True(t, found, "expect post to be found")
			urequire.Equal(t, tt.repostID.String(), repostID.String(), "expect repost ID to match")
		})
	}
}

func TestRepostStorageAdd(t *testing.T) {
	tests := []struct {
		name   string
		repost *boards.Post
		errMsg string
	}{
		{
			name: "ok",
			repost: &boards.Post{
				ID:    1,
				Board: &boards.Board{ID: 1},
			},
		},
		{
			name:   "nil repost",
			repost: nil,
			errMsg: "saving nil reposts is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := boards.NewRepostStorage()

			err := s.Add(tt.repost)

			if tt.errMsg != "" {
				urequire.Error(t, err, "expect an error")
				urequire.ErrorContains(t, err, tt.errMsg, "expect error to match")
				return
			}

			urequire.NoError(t, err, "expect no error")

			_, found := s.Get(tt.repost.ID)
			urequire.True(t, found, "expect repost to be found")
		})
	}
}

func TestRepostStorageRemove(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() boards.RepostStorage
		boardID boards.ID
		removed bool
	}{
		{
			name: "ok",
			setup: func() boards.RepostStorage {
				s := boards.NewRepostStorage()
				s.Add(&boards.Post{
					ID:    100,
					Board: &boards.Board{ID: 1},
				})
				return s
			},
			boardID: 1,
			removed: true,
		},
		{
			name: "not found",
			setup: func() boards.RepostStorage {
				return boards.NewRepostStorage()
			},
			boardID: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			removed := s.Remove(tt.boardID)

			if !tt.removed {
				urequire.False(t, removed, "expect repost not to be removed")
				return
			}

			urequire.True(t, removed, "expect repost to be removed")

			_, found := s.Get(tt.boardID)
			urequire.False(t, found, "expect repost not to be found")
		})
	}
}

func TestRepostStorageSize(t *testing.T) {
	tests := []struct {
		name  string
		setup func() boards.RepostStorage
		size  int
	}{
		{
			name: "empty",
			setup: func() boards.RepostStorage {
				return boards.NewRepostStorage()
			},
			size: 0,
		},
		{
			name: "one repost",
			setup: func() boards.RepostStorage {
				s := boards.NewRepostStorage()
				s.Add(&boards.Post{
					ID:    1,
					Board: &boards.Board{ID: 1},
				})
				return s
			},
			size: 1,
		},
		{
			name: "multiple reposts",
			setup: func() boards.RepostStorage {
				s := boards.NewRepostStorage()
				s.Add(&boards.Post{
					ID:    1,
					Board: &boards.Board{ID: 1},
				})
				s.Add(&boards.Post{
					ID:    1,
					Board: &boards.Board{ID: 2},
				})
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

func TestRepostStorageIterate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() boards.RepostStorage
		reverse bool
		ids     [][2]boards.ID
	}{
		{
			name: "default",
			setup: func() boards.RepostStorage {
				s := boards.NewRepostStorage()
				s.Add(&boards.Post{
					ID:    10,
					Board: &boards.Board{ID: 1},
				})
				s.Add(&boards.Post{
					ID:    20,
					Board: &boards.Board{ID: 2},
				})
				s.Add(&boards.Post{
					ID:    30,
					Board: &boards.Board{ID: 3},
				})
				return s
			},
			ids: [][2]boards.ID{
				{1, 10},
				{2, 20},
				{3, 30},
			},
		},
		{
			name: "reverse",
			setup: func() boards.RepostStorage {
				s := boards.NewRepostStorage()
				s.Add(&boards.Post{
					ID:    10,
					Board: &boards.Board{ID: 1},
				})
				s.Add(&boards.Post{
					ID:    20,
					Board: &boards.Board{ID: 2},
				})
				s.Add(&boards.Post{
					ID:    30,
					Board: &boards.Board{ID: 3},
				})
				return s
			},
			reverse: true,
			ids: [][2]boards.ID{
				{3, 30},
				{2, 20},
				{1, 10},
			},
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
			s.Iterate(0, count, func(boardID, repostID boards.ID) bool {
				urequire.True(t, tt.ids[i][0] == boardID, "expect board ID to match")
				urequire.True(t, tt.ids[i][1] == repostID, "expect repost ID to match")

				i++
				return false
			})
		})
	}
}
