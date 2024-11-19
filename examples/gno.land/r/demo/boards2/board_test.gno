package boards

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestBoardID_String(t *testing.T) {
	input := BoardID(32)

	uassert.Equal(t, "32", input.String())
}

func TestBoardID_Key(t *testing.T) {
	input := BoardID(128)
	want := strings.Repeat("0", 7) + "128"
	uassert.Equal(t, want, input.Key())
}

func TestBoard_IsPrivate(t *testing.T) {
	b := new(Board)
	b.id = 0
	uassert.True(t, b.IsPrivate())

	b.id = 128
	uassert.False(t, b.IsPrivate())
}

func TestBoard_GetID(t *testing.T) {
	want := int(92)
	b := new(Board)
	b.id = BoardID(want)
	got := int(b.GetID())

	uassert.Equal(t, got, want)
	uassert.NotEqual(t, got, want*want)
}

func TestBoard_GetURL(t *testing.T) {
	// TODO
}
