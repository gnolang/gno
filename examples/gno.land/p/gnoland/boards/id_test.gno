package boards_test

import (
	"testing"

	"gno.land/p/nt/urequire"

	"gno.land/p/gnoland/boards"
)

func TestID(t *testing.T) {
	id := boards.ID(42)

	urequire.Equal(t, "42", id.String(), "expect string to match")
	urequire.Equal(t, "0000000042", id.PaddedString(), "expect padded string to match")
	urequire.Equal(t, "000001a", id.Key(), "expect key to match")
}
