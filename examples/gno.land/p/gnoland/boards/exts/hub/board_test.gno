package hub_test

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/hub"
	"gno.land/p/gnoland/boards/exts/permissions"
	"gno.land/p/nt/urequire/v0"
)

func TestNewSafeBoard(t *testing.T) {
	tests := []struct {
		name                     string
		setup                    func() *boards.Board
		threadCount, memberCount int
		createdAt                int64
	}{
		{
			name:      "new board",
			setup:     func() *boards.Board { return boards.New(1) },
			createdAt: testTime,
		},
		{
			name: "with threads",
			setup: func() *boards.Board {
				b := boards.New(1)
				b.Threads.Add(boards.MustNewThread(b, alice, "Title 1", "Body 1"))
				b.Threads.Add(boards.MustNewThread(b, bob, "Title 2", "Body 2"))
				return b
			},
			threadCount: 2,
			createdAt:   testTime,
		},
		{
			name: "with members",
			setup: func() *boards.Board {
				b := boards.New(1)
				perms := permissions.New()
				perms.SetUserRoles(alice)
				perms.SetUserRoles(bob)
				b.Permissions = perms
				return b
			},
			memberCount: 2,
			createdAt:   testTime,
		},
		{
			// A board with no storages assigned must not panic: the counts
			// are skipped and the zero creation time maps to a zero Unix time.
			name:  "bare board without storages",
			setup: func() *boards.Board { return &boards.Board{ID: 1} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := tt.setup()
			ref.Name = "test123"
			ref.Creator = alice

			board := hub.NewSafeBoard(ref)

			urequire.Equal(t, uint64(1), board.ID(), "expect ID to match")
			urequire.Equal(t, "test123", board.Name(), "expect name to match")
			urequire.Equal(t, alice, board.Creator(), "expect creator to match")
			urequire.False(t, board.Readonly(), "expect board not to be readonly")
			urequire.Equal(t, tt.threadCount, board.ThreadCount(), "expect thread count to match")
			urequire.Equal(t, tt.memberCount, board.MemberCount(), "expect member count to match")
			urequire.Equal(t, tt.createdAt, board.CreatedAt(), "expect creation time to match")
			urequire.Equal(t, int64(0), board.UpdatedAt(), "expect zero update time to map to zero")
		})
	}
}

// TestNewSafeBoardSnapshotsAliases asserts the invariant this package exists
// for: a safe type is a snapshot, so later writes to the source board must not
// be observable through a value already handed to a caller.
func TestNewSafeBoardSnapshotsAliases(t *testing.T) {
	ref := boards.New(1)
	ref.Name = "current"
	ref.Aliases = []string{"previous"}

	board := hub.NewSafeBoard(ref)

	ref.Aliases[0] = "mutated"
	ref.Aliases = append(ref.Aliases, "added")
	ref.Name = "renamed"
	ref.Readonly = true

	aliases := board.Aliases()
	urequire.Equal(t, 1, len(aliases), "expect alias count to be unaffected")
	urequire.Equal(t, "previous", aliases[0], "expect alias to be unaffected")
	urequire.Equal(t, "current", board.Name(), "expect name to be unaffected")
	urequire.False(t, board.Readonly(), "expect readonly to be unaffected")
}

func TestNewSafeBoardNilRef(cur realm, t *testing.T) {
	urequire.PanicsWithMessage(t, cur, "board reference is nil", func() {
		hub.NewSafeBoard(nil)
	}, "expect a nil board reference to be rejected")
}
