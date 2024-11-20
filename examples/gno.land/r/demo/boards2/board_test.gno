package boards

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/moul/txlink"
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
	pkgPath := strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land")
	name := "foobar_test_get_url123"
	want := pkgPath + ":" + name

	var addr std.Address

	board := newBoard(1, name, addr)
	got := board.GetURL()
	uassert.Equal(t, want, got)
}

func TestBoard_GetThread(t *testing.T) {
	var addr std.Address
	b := newBoard(1, "test123", addr)

	_, ok := b.GetThread(12345)
	uassert.False(t, ok)

	post := b.AddThread(addr, "foo", "bar")
	_, ok = b.GetThread(post.GetPostID())
	uassert.True(t, ok)
}

func TestBoard_DeleteThread(t *testing.T) {
	var addr std.Address
	b := newBoard(1, "test123", addr)

	post := b.AddThread(addr, "foo", "bar")
	id := post.GetPostID()

	b.DeleteThread(id)

	_, ok := b.GetThread(id)
	uassert.False(t, ok)
}

var boardUrlPrefix = strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land")

func TestBoard_GetURLFromThreadID(t *testing.T) {
	boardName := "test12345"
	b := newBoard(BoardID(11), boardName, "")
	want := boardUrlPrefix + ":" + boardName + "/10"

	got := b.GetURLFromThreadID(10)
	uassert.Equal(t, want, got)
}

func TestBoard_GetURLFromReplyID(t *testing.T) {
	boardName := "test12345"
	b := newBoard(BoardID(11), boardName, "")
	want := boardUrlPrefix + ":" + boardName + "/10/20"

	got := b.GetURLFromReplyID(10, 20)
	uassert.Equal(t, want, got)
}

func TestBoard_GetPostFormURL(t *testing.T) {
	bid := BoardID(386)
	b := newBoard(bid, "foo1234", "")
	expect := txlink.URL("CreateThread", "bid", bid.String())
	got := b.GetPostFormURL()
	uassert.Equal(t, expect, got)
}
