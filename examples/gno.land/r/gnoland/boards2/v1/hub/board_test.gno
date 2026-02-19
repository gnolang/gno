package hub_test

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/urequire"

	"gno.land/r/gnoland/boards2/v1/hub"
	"gno.land/r/gnoland/boards2/v1/permissions"
)

func TestBoardIterateThreads(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*boards.Board)
	}{
		{
			name: "no threads",
		},
		{
			name: "one thread",
			setup: func(b *boards.Board) {
				t := boards.MustNewThread(b, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title 1", "Body 1")
				b.Threads.Add(t)
			},
		},
		{
			name: "multiple threads",
			setup: func(b *boards.Board) {
				threads := []*boards.Post{
					boards.MustNewThread(b, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title 1", "Body 1"),
					boards.MustNewThread(b, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", "Title 2", "Body 2"),
					boards.MustNewThread(b, "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq", "Title 3", "Body 3"),
				}
				for _, t := range threads {
					b.Threads.Add(t)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := boards.New(1)
			if tt.setup != nil {
				tt.setup(ref)
			}

			board := hub.NewSafeBoard(ref)

			urequire.Equal(t, ref.Threads.Size(), board.ThreadCount, "expect number of threads to match")
			board.IterateThreads(0, board.ThreadCount, func(thread hub.Thread) bool {
				id := boards.ID(thread.ID)
				expected, found := ref.Threads.Get(id)

				urequire.True(t, found, "expect thread to be found")
				urequire.Equal(t, expected.Creator, thread.Creator, "expect creator to match")
				urequire.Equal(t, expected.Title, thread.Title, "expect title to match")
				urequire.Equal(t, expected.Body, thread.Body, "expect body to match")
				return false
			})
		})
	}
}

func TestBoardIterateMembers(t *testing.T) {
	tests := []struct {
		name  string
		users []boards.User
	}{
		{
			name: "no members",
		},
		{
			name: "one member",
			users: []boards.User{
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
			},
		},
		{
			name: "multiple members",
			users: []boards.User{
				{Address: "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"},
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
				{Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i int
			perms := permissions.New()
			for _, u := range tt.users {
				perms.SetUserRoles(cross, u.Address)
			}

			ref := boards.New(1)
			ref.Permissions = perms
			board := hub.NewSafeBoard(ref)

			urequire.Equal(t, len(tt.users), board.MemberCount, "expect number of members to match")
			board.IterateMembers(0, board.MemberCount, func(user boards.User) bool {
				urequire.Equal(t, tt.users[i].Address, user.Address, "expect address to match")
				i++
				return false
			})
		})
	}
}
