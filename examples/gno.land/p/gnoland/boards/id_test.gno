package boards_test

import (
	"testing"

	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards"
)

func TestID(t *testing.T) {
	id := boards.ID(42)

	urequire.Equal(t, "42", id.String(), "expect string to match")
	urequire.Equal(t, "0000000042", id.PaddedString(), "expect padded string to match")
	urequire.Equal(t, "000001a", id.Key(), "expect key to match")
}

func TestIdentifierGenerator(t *testing.T) {
	g := boards.NewIdentifierGenerator()

	urequire.Equal(t, uint64(0), uint64(g.Last()), "expect default to be 0")
	urequire.Equal(t, uint64(1), uint64(g.Next()), "expect next to be 1")
	urequire.Equal(t, uint64(1), uint64(g.Last()), "expect last to be 1")
	urequire.Equal(t, uint64(2), uint64(g.Next()), "expect next to be 2")
	urequire.Equal(t, uint64(2), uint64(g.Last()), "expect last to be 2")
}
