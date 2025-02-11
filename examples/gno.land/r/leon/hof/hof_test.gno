package hof

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

const (
	rlmPath            = "gno.land/r/gnoland/home"
	rlmPath2           = "gno.land/r/gnoland/test2"

	rlmPath3           = "gno.land/r/gnoland/test3"
	rlmPath4           = "gno.land/r/gnoland/test4"
	rlmPath5           = "gno.land/r/gnoland/test5"

	validTitle         = "valid title"
	invalidTitle       = "This title is very very very long, longer than 30 characters"
	validDesc          = "valid description"
	invalidDescription = "This description is very very very long, longer than 50 characters"
)

var (
	admin      = Ownable.Owner()
	adminRealm = std.NewUserRealm(admin)
	alice      = testutils.TestAddress("alice")
)

func TestRegister(t *testing.T) {
	// Test user realm register
	aliceRealm := std.NewUserRealm(alice)
	std.TestSetRealm(aliceRealm)

	Register(validTitle, validDesc)
	uassert.False(t, itemExists(t, rlmPath))

	// Test register while paused
	std.TestSetRealm(adminRealm)
	Pausable.Pause()

	// Set legitimate caller
	std.TestSetRealm(std.NewCodeRealm(rlmPath))

	Register(validTitle, validDesc)
	uassert.False(t, itemExists(t, rlmPath))

	// Unpause
	std.TestSetRealm(adminRealm)
	Pausable.Unpause()

	// Set legitimate caller
	std.TestSetRealm(std.NewCodeRealm(rlmPath))
	Register(validTitle, validDesc)

	// Find registered items
	uassert.True(t, itemExists(t, rlmPath))

	//Test register with invalid title
	std.TestSetRealm(std.NewCodeRealm(rlmPath2))
	Register(invalidTitle, validDesc)
	uassert.False(t, itemExists(t, rlmPath2))

	//Test register with invalid description
	std.TestSetRealm(std.NewCodeRealm(rlmPath2))
	Register(validTitle, invalidDescription)
	uassert.False(t, itemExists(t, rlmPath2))
}

func TestUpvote(t *testing.T) {
	raw, _ := exhibition.items.Get(rlmPath)
	item := raw.(*Item)

	rawSorted, _ := exhibition.itemsSorted.Get(item.id.String())
	itemSorted := rawSorted.(*Item)

	// 0 upvotes by default
	urequire.Equal(t, item.upvote.Size(), 0)

	std.TestSetRealm(adminRealm)

	urequire.NotPanics(t, func() {
		Upvote(rlmPath)
	})

	// Check both trees for 1 upvote
	uassert.Equal(t, item.upvote.Size(), 1)
	uassert.Equal(t, itemSorted.upvote.Size(), 1)

	// Check double upvote
	uassert.PanicsWithMessage(t, ErrDoubleUpvote.Error(), func() {
		Upvote(rlmPath)
	})
}

func TestDownvote(t *testing.T) {
	raw, _ := exhibition.items.Get(rlmPath)
	item := raw.(*Item)

	rawSorted, _ := exhibition.itemsSorted.Get(item.id.String())
	itemSorted := rawSorted.(*Item)

	// 0 downvotes by default
	urequire.Equal(t, item.downvote.Size(), 0)

	userRealm := std.NewUserRealm(alice)
	std.TestSetRealm(userRealm)

	urequire.NotPanics(t, func() {
		Downvote(rlmPath)
	})

	// Check both trees for 1 upvote
	uassert.Equal(t, item.downvote.Size(), 1)
	uassert.Equal(t, itemSorted.downvote.Size(), 1)

	// Check double downvote
	uassert.PanicsWithMessage(t, ErrDoubleDownvote.Error(), func() {
		Downvote(rlmPath)
	})
}

func TestDelete(t *testing.T) {
	userRealm := std.NewUserRealm(admin)
	std.TestSetRealm(userRealm)
	std.TestSetOrigCaller(admin)

	uassert.PanicsWithMessage(t, ErrNoSuchItem.Error(), func() {
		Delete("nonexistentpkgpath")
	})

	i, _ := exhibition.items.Get(rlmPath)
	id := i.(*Item).id

	uassert.NotPanics(t, func() {
		Delete(rlmPath)
	})

	uassert.False(t, exhibition.items.Has(rlmPath))
	uassert.False(t, exhibition.itemsSorted.Has(id.String()))
}

func itemExists(t *testing.T, rlmPath string) bool {
	t.Helper()

	i, ok1 := exhibition.items.Get(rlmPath)
	ok2 := false

	if ok1 {
		_, ok2 = exhibition.itemsSorted.Get(i.(*Item).id.String())
	}

	return ok1 && ok2
}

func TestgenerateUpvoteSortKey(t *testing.T) {
	i := &Item{
		id:          1,
		title:       validTitle,
		description: validDesc,
		pkgpath:     rlmPath,
		blockNum:    std.GetHeight(),
		upvote:      avl.NewTree(),
		downvote:    avl.NewTree(),
	}

	i.upvote.Set(alice.String(), struct{}{})

	generatedKey := generateUpvoteSortKey(i)
	expectedKey := "0000000001:1"

	urequire.Equal(t, generatedKey, expectedKey)
}
