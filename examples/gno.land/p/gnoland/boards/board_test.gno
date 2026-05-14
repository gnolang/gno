package boards_test

import (
	"testing"

	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards"
)

func TestNew(t *testing.T) {
	board := boards.New(42)

	urequire.Equal(t, 42, int(board.ID), "expect board ID to match")
	urequire.True(t, board.Threads != nil, "expect board to support threads")
	urequire.True(t, board.ThreadsSequence != nil, "expect board to initialize a thread ID generator")
	urequire.False(t, board.CreatedAt.IsZero(), "expect board to have a creation date")
}
