package boards2

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestBoardURL(t *testing.T) {
	pkgPath := strings.TrimPrefix(string(gRealmLink), "gno.land")
	name := "foobar_test_get_url123"
	want := pkgPath + ":" + name

	addr := testutils.TestAddress("creator")
	perms := createBasicBoardPermissions(addr)
	board := newBoard(1, name, addr, perms)
	got := board.URL()
	uassert.Equal(t, want, got)
}

func TestBoardGetThread(t *testing.T) {
	addr := testutils.TestAddress("creator")
	perms := createBasicBoardPermissions(addr)
	b := newBoard(1, "test123", addr, perms)

	_, ok := b.GetThread(12345)
	uassert.False(t, ok)

	post := b.AddThread(addr, "foo", "bar")
	_, ok = b.GetThread(post.ID)
	uassert.True(t, ok)
}

func TestBoardDeleteThread(t *testing.T) {
	addr := testutils.TestAddress("creator")
	perms := createBasicBoardPermissions(addr)
	b := newBoard(1, "test123", addr, perms)

	post := b.AddThread(addr, "foo", "bar")
	b.DeleteThread(post.ID)

	_, ok := b.GetThread(post.ID)
	uassert.False(t, ok)
}

var boardUrlPrefix = strings.TrimPrefix(string(gRealmLink), "gno.land")

func TestBoardGetURLFromThreadID(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoland/boards2/v1"))

	boardName := "test12345"
	addr := testutils.TestAddress("creator")
	perms := createBasicBoardPermissions(addr)
	b := newBoard(BoardID(11), boardName, addr, perms)
	want := boardUrlPrefix + ":" + boardName + "/10"

	got := b.GetURLFromThreadID(10)
	uassert.Equal(t, want, got)
}

func TestBoardGetURLFromReplyID(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoland/boards2/v1"))

	boardName := "test12345"
	addr := testutils.TestAddress("creator")
	perms := createBasicBoardPermissions(addr)
	b := newBoard(BoardID(11), boardName, addr, perms)
	want := boardUrlPrefix + ":" + boardName + "/10/20"

	got := b.GetURLFromReplyID(10, 20)
	uassert.Equal(t, want, got)
}

func TestBoardGetPostFormURL(t *testing.T) {
	bid := BoardID(386)
	addr := testutils.TestAddress("creator")
	perms := createBasicBoardPermissions(addr)
	b := newBoard(bid, "foo1234", addr, perms)
	expect := gRealmLink.Call("CreateThread", "boardID", bid.String(), "title", "", "body", "")

	got := b.GetPostFormURL()
	uassert.Equal(t, expect, got)
}
