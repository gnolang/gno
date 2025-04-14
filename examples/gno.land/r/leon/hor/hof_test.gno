package hor

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/moul/addrset"
)

const (
	rlmPath  = "gno.land/r/gnoland/home"
	rlmPath2 = "gno.land/r/gnoland/test2"

	rlmPath3 = "gno.land/r/gnoland/test3"
	rlmPath4 = "gno.land/r/gnoland/test4"
	rlmPath5 = "gno.land/r/gnoland/test5"

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
	testing.SetRealm(aliceRealm)

	Register(validTitle, validDesc)
	uassert.False(t, itemExists(t, rlmPath))

	// Test register while paused
	testing.SetRealm(adminRealm)
	Pausable.Pause()

	// Set legitimate caller
	testing.SetRealm(std.NewCodeRealm(rlmPath))

	Register(validTitle, validDesc)
	uassert.False(t, itemExists(t, rlmPath))

	// Unpause
	testing.SetRealm(adminRealm)
	Pausable.Unpause()

	// Set legitimate caller
	testing.SetRealm(std.NewCodeRealm(rlmPath))
	Register(validTitle, validDesc)

	// Find registered items
	uassert.True(t, itemExists(t, rlmPath))

	// Test register with invalid title
	testing.SetRealm(std.NewCodeRealm(rlmPath2))
	Register(invalidTitle, validDesc)
	uassert.False(t, itemExists(t, rlmPath2))

	// Test register with invalid description
	testing.SetRealm(std.NewCodeRealm(rlmPath2))
	Register(validTitle, invalidDescription)
	uassert.False(t, itemExists(t, rlmPath2))
}

func TestUpvote(t *testing.T) {
	raw, _ := exhibition.items.Get(rlmPath)
	item := raw.(*Item)

	// 0 upvotes by default
	urequire.Equal(t, item.upvote.Size(), 0)

	testing.SetRealm(adminRealm)

	urequire.NotPanics(t, func() {
		Upvote(rlmPath)
	})

	// Check both trees for 1 upvote
	uassert.Equal(t, item.upvote.Size(), 1)

	// Check double upvote
	uassert.PanicsWithMessage(t, ErrDoubleUpvote.Error(), func() {
		Upvote(rlmPath)
	})
}

func TestDownvote(t *testing.T) {
	raw, _ := exhibition.items.Get(rlmPath)
	item := raw.(*Item)

	// 0 downvotes by default
	urequire.Equal(t, item.downvote.Size(), 0)

	userRealm := std.NewUserRealm(alice)
	testing.SetRealm(userRealm)

	urequire.NotPanics(t, func() {
		Downvote(rlmPath)
	})

	// Check both trees for 1 upvote
	uassert.Equal(t, item.downvote.Size(), 1)

	// Check double downvote
	uassert.PanicsWithMessage(t, ErrDoubleDownvote.Error(), func() {
		Downvote(rlmPath)
	})
}

func TestDelete(t *testing.T) {
	userRealm := std.NewUserRealm(admin)
	testing.SetRealm(userRealm)
	testing.SetOriginCaller(admin)

	uassert.PanicsWithMessage(t, ErrNoSuchItem.Error(), func() {
		Delete("nonexistentpkgpath")
	})

	i, _ := exhibition.items.Get(rlmPath)
	id := i.(*Item).id

	uassert.NotPanics(t, func() {
		Delete(rlmPath)
	})

	uassert.False(t, exhibition.items.Has(rlmPath))
}

func itemExists(t *testing.T, rlmPath string) bool {
	t.Helper()

	i, ok1 := exhibition.items.Get(rlmPath)

	return ok1
}

func TestgetVoteSortKey(t *testing.T) {
	i := &Item{
		id:          1,
		title:       validTitle,
		description: validDesc,
		pkgpath:     rlmPath,
		blockNum:    std.ChainHeight(),
		upvote:      &addrset.Set{},
		downvote:    &addrset.Set{},
	}

	i.upvote.Add(alice)

	generatedKey := getVoteSortKey(i.upvote.Size(), i.id)
	expectedKey := "0000000001:1"

	urequire.Equal(t, generatedKey, expectedKey)
}

func TestSortByUpvote(t *testing.T) {
	// Remove all items from all trees
	exhibition.items.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.items.Remove(key)
		return false
	})
	exhibition.itemsSortedByUpvotes.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.itemsSortedByUpvotes.Remove(key)
		return false
	})
	exhibition.itemsSortedByDownvotes.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.itemsSortedByDownvotes.Remove(key)
		return false
	})
	exhibition.itemsSortedByCreation.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.itemsSortedByCreation.Remove(key)
		return false
	})

	// Add items
	testing.SetRealm(std.NewCodeRealm(rlmPath3))
	Register(validTitle, validDesc)

	testing.SetRealm(std.NewCodeRealm(rlmPath4))
	Register(validTitle, validDesc)

	testing.SetRealm(std.NewCodeRealm(rlmPath5))
	Register(validTitle, validDesc)

	user1 := testutils.TestAddress("user1")
	user2 := testutils.TestAddress("user2")
	user3 := testutils.TestAddress("user3")

	testing.SetOriginCaller(user1)
	testing.SetRealm(std.NewUserRealm(user1))
	Upvote(rlmPath3)
	Upvote(rlmPath4)
	Upvote(rlmPath5)

	testing.SetOriginCaller(user2)
	testing.SetRealm(std.NewUserRealm(user2))
	Upvote(rlmPath4)
	Upvote(rlmPath5)

	testing.SetOriginCaller(user3)
	testing.SetRealm(std.NewUserRealm(user3))
	Upvote(rlmPath5)

	// We are displaying data in reverse order in render, so items should be sorted in reverse order
	firstKey, firstRawValue := exhibition.itemsSortedByUpvotes.GetByIndex(0)
	firstValue := firstRawValue.(*Item)
	uassert.Equal(t, firstValue.pkgpath, rlmPath3)

	secondKey, secondRawValue := exhibition.itemsSortedByUpvotes.GetByIndex(1)
	secondValue := secondRawValue.(*Item)
	uassert.Equal(t, secondValue.pkgpath, rlmPath4)
}

func TestSortByDownvote(t *testing.T) {
	// Remove all items from all trees
	exhibition.items.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.items.Remove(key)
		return false
	})
	exhibition.itemsSortedByUpvotes.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.itemsSortedByUpvotes.Remove(key)
		return false
	})
	exhibition.itemsSortedByDownvotes.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.itemsSortedByDownvotes.Remove(key)
		return false
	})
	exhibition.itemsSortedByCreation.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.itemsSortedByCreation.Remove(key)
		return false
	})

	// Add items
	testing.SetRealm(std.NewCodeRealm(rlmPath3))
	Register(validTitle, validDesc)

	testing.SetRealm(std.NewCodeRealm(rlmPath4))
	Register(validTitle, validDesc)

	testing.SetRealm(std.NewCodeRealm(rlmPath5))
	Register(validTitle, validDesc)

	user1 := testutils.TestAddress("user1")
	user2 := testutils.TestAddress("user2")
	user3 := testutils.TestAddress("user3")

	testing.SetOriginCaller(user1)
	testing.SetRealm(std.NewUserRealm(user1))
	Downvote(rlmPath3)
	Downvote(rlmPath4)
	Downvote(rlmPath5)

	testing.SetOriginCaller(user2)
	testing.SetRealm(std.NewUserRealm(user2))
	Downvote(rlmPath4)
	Downvote(rlmPath5)

	testing.SetOriginCaller(user3)
	testing.SetRealm(std.NewUserRealm(user3))
	Downvote(rlmPath5)

	// We are dispalying data is reverse order in render, so items should be sorted in reverse order
	firstKey, firstRawValue := exhibition.itemsSortedByDownvotes.GetByIndex(0)

	firstValue := firstRawValue.(*Item)

	uassert.Equal(t, firstValue.pkgpath, rlmPath3)

	secondKey, secondRawValue := exhibition.itemsSortedByDownvotes.GetByIndex(1)

	secondValue := secondRawValue.(*Item)

	uassert.Equal(t, secondValue.pkgpath, rlmPath4)
}

func TestSortByCreation(t *testing.T) {
	// Remove all items from all trees
	exhibition.items.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.items.Remove(key)
		return false
	})
	exhibition.itemsSortedByUpvotes.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.itemsSortedByUpvotes.Remove(key)
		return false
	})
	exhibition.itemsSortedByDownvotes.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.itemsSortedByDownvotes.Remove(key)
		return false
	})
	exhibition.itemsSortedByCreation.Iterate("", "", func(key string, value interface{}) bool {
		exhibition.itemsSortedByCreation.Remove(key)
		return false
	})

	testing.SkipHeights(10)
	testing.SetRealm(std.NewCodeRealm(rlmPath3))
	Register(validTitle, validDesc)

	testing.SkipHeights(10)
	testing.SetRealm(std.NewCodeRealm(rlmPath4))
	Register(validTitle, validDesc)

	testing.SkipHeights(10)
	testing.SetRealm(std.NewCodeRealm(rlmPath5))
	Register(validTitle, validDesc)

	// We are dispalying data is reverse order in render, so items should be sorted in reverse order
	firstKey, firstRawValue := exhibition.itemsSortedByCreation.GetByIndex(0)

	firstValue := firstRawValue.(*Item)

	uassert.Equal(t, firstValue.pkgpath, rlmPath3)

	secondKey, secondRawValue := exhibition.itemsSortedByCreation.GetByIndex(1)

	secondValue := secondRawValue.(*Item)

	uassert.Equal(t, secondValue.pkgpath, rlmPath4)
}
