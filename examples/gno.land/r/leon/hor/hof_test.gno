package hor

import (
	"chain/runtime"
	"testing"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
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
	admin      address = "g125em6arxsnj49vx35f0n0z34putv5ty3376fg5" // hardcoded from config.OwnableMain
	adminRealm runtime.Realm
	alice      = testutils.TestAddress("alice")
)

func init() {
	adminRealm = testing.NewUserRealm(admin)
}

func TestRegister(t *testing.T) {
	// Test user realm register
	aliceRealm := testing.NewUserRealm(alice)
	testing.SetRealm(aliceRealm)

	Register(cross, validTitle, validDesc)
	uassert.False(t, itemExists(t, rlmPath))

	// Test register while paused
	testing.SetRealm(adminRealm)
	err := Pause(cross)
	uassert.NoError(t, err)

	// Set legitimate caller
	testing.SetRealm(testing.NewCodeRealm(rlmPath))

	Register(cross, validTitle, validDesc)
	uassert.False(t, itemExists(t, rlmPath))

	// Unpause
	testing.SetRealm(adminRealm)
	err = Unpause(cross)
	uassert.NoError(t, err)

	// Set legitimate caller
	testing.SetRealm(testing.NewCodeRealm(rlmPath))
	Register(cross, validTitle, validDesc)

	// Find registered items
	uassert.True(t, itemExists(t, rlmPath))

	// Test register with invalid title
	testing.SetRealm(testing.NewCodeRealm(rlmPath2))
	Register(cross, invalidTitle, validDesc)
	uassert.False(t, itemExists(t, rlmPath2))

	// Test register with invalid description
	testing.SetRealm(testing.NewCodeRealm(rlmPath2))
	Register(cross, validTitle, invalidDescription)
	uassert.False(t, itemExists(t, rlmPath2))
}

func TestUpvote(t *testing.T) {
	raw, _ := exhibition.items.Get(rlmPath)
	item := raw.(*Item)

	// 0 upvotes by default
	urequire.Equal(t, item.upvote.Size(), 0)

	testing.SetRealm(adminRealm)

	urequire.NotPanics(t, func() {
		Upvote(cross, rlmPath)
	})

	// Check both trees for 1 upvote
	uassert.Equal(t, item.upvote.Size(), 1)

	// Check double upvote
	uassert.AbortsWithMessage(t, ErrDoubleUpvote.Error(), func() {
		Upvote(cross, rlmPath)
	})
}

func TestDownvote(t *testing.T) {
	raw, _ := exhibition.items.Get(rlmPath)
	item := raw.(*Item)

	// 0 downvotes by default
	urequire.Equal(t, item.downvote.Size(), 0)

	userRealm := testing.NewUserRealm(alice)
	testing.SetRealm(userRealm)

	urequire.NotPanics(t, func() {
		Downvote(cross, rlmPath)
	})

	// Check both trees for 1 upvote
	uassert.Equal(t, item.downvote.Size(), 1)

	// Check double downvote
	uassert.AbortsWithMessage(t, ErrDoubleDownvote.Error(), func() {
		Downvote(cross, rlmPath)
	})
}

func TestDelete(t *testing.T) {
	userRealm := testing.NewUserRealm(admin)
	testing.SetRealm(userRealm)
	testing.SetOriginCaller(admin)

	uassert.AbortsWithMessage(t, ErrNoSuchItem.Error(), func() {
		Delete(cross, "nonexistentpkgpath")
	})

	i, _ := exhibition.items.Get(rlmPath)
	_ = i.(*Item).id

	uassert.NotPanics(t, func() {
		Delete(cross, rlmPath)
	})

	uassert.False(t, exhibition.items.Has(rlmPath))
}

func itemExists(t *testing.T, rlmPath string) bool {
	t.Helper()

	_, ok1 := exhibition.items.Get(rlmPath)

	return ok1
}

func TestgetVoteSortKey(t *testing.T) {
	i := &Item{
		id:          1,
		title:       validTitle,
		description: validDesc,
		pkgpath:     rlmPath,
		blockNum:    runtime.ChainHeight(),
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
	testing.SetRealm(testing.NewCodeRealm(rlmPath3))
	Register(cross, validTitle, validDesc)

	testing.SetRealm(testing.NewCodeRealm(rlmPath4))
	Register(cross, validTitle, validDesc)

	testing.SetRealm(testing.NewCodeRealm(rlmPath5))
	Register(cross, validTitle, validDesc)

	user1 := testutils.TestAddress("user1")
	user2 := testutils.TestAddress("user2")
	user3 := testutils.TestAddress("user3")

	testing.SetOriginCaller(user1)
	testing.SetRealm(testing.NewUserRealm(user1))
	Upvote(cross, rlmPath3)
	Upvote(cross, rlmPath4)
	Upvote(cross, rlmPath5)

	testing.SetOriginCaller(user2)
	testing.SetRealm(testing.NewUserRealm(user2))
	Upvote(cross, rlmPath4)
	Upvote(cross, rlmPath5)

	testing.SetOriginCaller(user3)
	testing.SetRealm(testing.NewUserRealm(user3))
	Upvote(cross, rlmPath5)

	// We are displaying data in reverse order in render, so items should be sorted in reverse order
	_, firstRawValue := exhibition.itemsSortedByUpvotes.GetByIndex(0)
	firstValue := firstRawValue.(*Item)
	uassert.Equal(t, firstValue.pkgpath, rlmPath3)

	_, secondRawValue := exhibition.itemsSortedByUpvotes.GetByIndex(1)
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
	testing.SetRealm(testing.NewCodeRealm(rlmPath3))
	Register(cross, validTitle, validDesc)

	testing.SetRealm(testing.NewCodeRealm(rlmPath4))
	Register(cross, validTitle, validDesc)

	testing.SetRealm(testing.NewCodeRealm(rlmPath5))
	Register(cross, validTitle, validDesc)

	user1 := testutils.TestAddress("user1")
	user2 := testutils.TestAddress("user2")
	user3 := testutils.TestAddress("user3")

	testing.SetOriginCaller(user1)
	testing.SetRealm(testing.NewUserRealm(user1))
	Downvote(cross, rlmPath3)
	Downvote(cross, rlmPath4)
	Downvote(cross, rlmPath5)

	testing.SetOriginCaller(user2)
	testing.SetRealm(testing.NewUserRealm(user2))
	Downvote(cross, rlmPath4)
	Downvote(cross, rlmPath5)

	testing.SetOriginCaller(user3)
	testing.SetRealm(testing.NewUserRealm(user3))
	Downvote(cross, rlmPath5)

	// We are dispalying data is reverse order in render, so items should be sorted in reverse order
	_, firstRawValue := exhibition.itemsSortedByDownvotes.GetByIndex(0)

	firstValue := firstRawValue.(*Item)

	uassert.Equal(t, firstValue.pkgpath, rlmPath3)

	_, secondRawValue := exhibition.itemsSortedByDownvotes.GetByIndex(1)

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
	testing.SetRealm(testing.NewCodeRealm(rlmPath3))
	Register(cross, validTitle, validDesc)

	testing.SkipHeights(10)
	testing.SetRealm(testing.NewCodeRealm(rlmPath4))
	Register(cross, validTitle, validDesc)

	testing.SkipHeights(10)
	testing.SetRealm(testing.NewCodeRealm(rlmPath5))
	Register(cross, validTitle, validDesc)

	// We are dispalying data is reverse order in render, so items should be sorted in reverse order
	_, firstRawValue := exhibition.itemsSortedByCreation.GetByIndex(0)

	firstValue := firstRawValue.(*Item)

	uassert.Equal(t, firstValue.pkgpath, rlmPath3)

	_, secondRawValue := exhibition.itemsSortedByCreation.GetByIndex(1)

	secondValue := secondRawValue.(*Item)

	uassert.Equal(t, secondValue.pkgpath, rlmPath4)
}

/* XXX fix @moul
func TestGetScore(t *testing.T) {
	score, err := h.GetScore(42)
	uassert.Nil(t, err)
	uassert.Equal(t, 0, score)
	uassert.NotPanics(t, func() {
		uassert.Equal(t, 0, h.MustGetScore(42))
	})
}
*/
