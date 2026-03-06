package boards_test

import (
	"testing"

	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards"
)

func TestStorageGet(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() boards.Storage
		boardID boards.ID
		found   bool
	}{
		{
			name: "single board",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1})
				return s
			},
			boardID: 1,
			found:   true,
		},
		{
			name: "multiple boards",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1})
				s.Add(&boards.Board{ID: 2})
				s.Add(&boards.Board{ID: 3})
				return s
			},
			boardID: 2,
			found:   true,
		},
		{
			name: "not found",
			setup: func() boards.Storage {
				return boards.NewStorage()
			},
			boardID: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			board, found := s.Get(tt.boardID)

			if !tt.found {
				urequire.False(t, found, "expect board not to be found")
				urequire.True(t, board == nil, "expect board to be nil")
				return
			}

			urequire.True(t, found, "expect board to be found")
			urequire.False(t, board == nil, "expect board not to be nil")
			urequire.Equal(t, tt.boardID.String(), board.ID.String(), "expect board ID to match")
		})
	}
}

func TestStorageGetByName(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() boards.Storage
		boardName string
		found     bool
	}{
		{
			name: "single board",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1, Name: "A"})
				return s
			},
			boardName: "A",
			found:     true,
		},
		{
			name: "multiple boards",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1, Name: "A"})
				s.Add(&boards.Board{ID: 2, Name: "B"})
				s.Add(&boards.Board{ID: 3, Name: "C"})
				return s
			},
			boardName: "B",
			found:     true,
		},
		{
			name: "not found",
			setup: func() boards.Storage {
				return boards.NewStorage()
			},
			boardName: "foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			board, found := s.GetByName(tt.boardName)

			if !tt.found {
				urequire.False(t, found, "expect board not to be found")
				urequire.True(t, board == nil, "expect board to be nil")
				return
			}

			urequire.True(t, found, "expect board to be found")
			urequire.False(t, board == nil, "expect board not to be nil")
			urequire.Equal(t, tt.boardName, board.Name, "expect board name to match")
		})
	}
}

func TestStorageRemove(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() boards.Storage
		boardID    boards.ID
		boardNames []string
		removed    bool
	}{
		{
			name: "ok",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1, Name: "A"})
				s.Add(&boards.Board{ID: 2, Name: "B"})
				return s
			},
			boardID:    2,
			boardNames: []string{"B"},
			removed:    true,
		},
		{
			name: "ok with aliases",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1, Name: "A"})

				b := &boards.Board{ID: 2, Name: "B"}
				s.Add(b)

				b.Aliases = []string{"A"}
				b.Name = "C"
				s.Add(b)

				return s
			},
			boardID:    2,
			boardNames: []string{"B", "C"},
			removed:    true,
		},
		{
			name: "not found",
			setup: func() boards.Storage {
				return boards.NewStorage()
			},
			boardID: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			board, removed := s.Remove(tt.boardID)

			if !tt.removed {
				urequire.False(t, removed, "expect board not to be removed")
				urequire.True(t, board == nil, "expect board to be nil")
				return
			}

			urequire.True(t, removed, "expect board to be removed")
			urequire.False(t, board == nil, "expect board not to be nil")
			urequire.Equal(t, tt.boardID.String(), board.ID.String(), "expect board ID to match")

			_, found := s.Get(tt.boardID)
			urequire.False(t, found, "expect board not to be found by ID")

			for _, name := range tt.boardNames {
				_, found = s.GetByName(name)
				urequire.False(t, found, "expect board not to be found by name: "+name)
			}
		})
	}
}

func TestStorageAdd(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() boards.Storage
		board  *boards.Board
		errMsg string
	}{
		{
			name:  "ok",
			board: &boards.Board{ID: 1, Name: "A"},
		},
		{
			name:   "nil board",
			board:  nil,
			errMsg: "adding nil boards to the storage is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := boards.NewStorage()

			err := s.Add(tt.board)

			if tt.errMsg != "" {
				urequire.Error(t, err, "expect an error")
				urequire.ErrorContains(t, err, tt.errMsg, "expect error to match")
				return
			}

			urequire.NoError(t, err, "expect no error")

			_, found := s.Get(tt.board.ID)
			urequire.True(t, found, "expect board to be found by ID")

			_, found = s.GetByName(tt.board.Name)
			urequire.True(t, found, "expect board to be found by name")
		})
	}
}

func TestStorageSize(t *testing.T) {
	tests := []struct {
		name  string
		setup func() boards.Storage
		size  int
	}{
		{
			name: "empty",
			setup: func() boards.Storage {
				return boards.NewStorage()
			},
			size: 0,
		},
		{
			name: "one board",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1})
				return s
			},
			size: 1,
		},
		{
			name: "multiple boards",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1})
				s.Add(&boards.Board{ID: 2})
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

func TestStorageIterate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() boards.Storage
		reverse bool
		ids     []boards.ID
	}{
		{
			name: "default",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1})
				s.Add(&boards.Board{ID: 2})
				s.Add(&boards.Board{ID: 3})
				return s
			},
			ids: []boards.ID{1, 2, 3},
		},
		{
			name: "reverse",
			setup: func() boards.Storage {
				s := boards.NewStorage()
				s.Add(&boards.Board{ID: 1})
				s.Add(&boards.Board{ID: 2})
				s.Add(&boards.Board{ID: 3})
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
			s.Iterate(0, count, func(p *boards.Board) bool {
				urequire.True(t, tt.ids[i] == p.ID, "expect board ID to match")

				i++
				return false
			})
		})
	}
}
