package hof

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

const rlm1Path = "gno.land/r/leon/home"

var admin = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

func TestRegister(t *testing.T) {
	userRealm := std.NewUserRealm(admin)
	std.TestSetRealm(userRealm)
	uassert.PanicsWithMessage(t, ErrNonCodeCall.Error(), Register)

	rlm1 := std.NewCodeRealm(rlm1Path)
	std.TestSetRealm(rlm1)
	urequire.NotPanics(t, Register)
	uassert.PanicsWithMessage(t, ErrAlreadyExists.Error(), Register)

	// Find registered items
	i, ok := exhibition.items.Get(rlm1Path)
	uassert.True(t, ok)

	_, ok = exhibition.itemsSorted.Get(i.(*Item).id.String())
	uassert.True(t, ok)

	// Check paused
	std.TestSetRealm(userRealm)
	urequire.NotPanics(t, func() { exhibition.Pause() })
	uassert.PanicsWithMessage(t, ErrPaused.Error(), Register)
}

func TestUpvote(t *testing.T) {
	raw, _ := exhibition.items.Get(rlm1Path)
	item := raw.(*Item)

	rawSorted, _ := exhibition.itemsSorted.Get(item.id.String())
	itemSorted := rawSorted.(*Item)

	// 0 upvotes by default
	urequire.Equal(t, item.upvote.Size(), 0)

	userRealm := std.NewUserRealm(admin)
	std.TestSetRealm(userRealm)

	urequire.NotPanics(t, func() {
		Upvote(rlm1Path)
	})

	// Check both trees for 1 upvote
	uassert.Equal(t, item.upvote.Size(), 1)
	uassert.Equal(t, itemSorted.upvote.Size(), 1)

	// Check double upvote
	uassert.PanicsWithMessage(t, ErrDoubleUpvote.Error(), func() {
		Upvote(rlm1Path)
	})
}

func TestDownvote(t *testing.T) {
	raw, _ := exhibition.items.Get(rlm1Path)
	item := raw.(*Item)

	rawSorted, _ := exhibition.itemsSorted.Get(item.id.String())
	itemSorted := rawSorted.(*Item)

	// 0 upvotes by default
	urequire.Equal(t, item.downvote.Size(), 0)

	userRealm := std.NewUserRealm(admin)
	std.TestSetRealm(userRealm)

	urequire.NotPanics(t, func() {
		Downvote(rlm1Path)
	})

	// Check both trees for 1 upvote
	uassert.Equal(t, item.downvote.Size(), 1)
	uassert.Equal(t, itemSorted.downvote.Size(), 1)

	// Check double upvote
	uassert.PanicsWithMessage(t, ErrDoubleDownvote.Error(), func() {
		Downvote(rlm1Path)
	})
}

func TestDelete(t *testing.T) {
	userRealm := std.NewUserRealm(admin)
	std.TestSetRealm(userRealm)
	std.TestSetOrigCaller(admin)

	uassert.PanicsWithMessage(t, ErrNoSuchItem.Error(), func() {
		Delete("nonexistentpkgpath")
	})

	i, ok := exhibition.items.Get(rlm1Path)
	id := i.(*Item).id

	uassert.NotPanics(t, func() {
		Delete(rlm1Path)
	})

	uassert.False(t, exhibition.items.Has(rlm1Path))
	uassert.False(t, exhibition.itemsSorted.Has(id.String()))
}
