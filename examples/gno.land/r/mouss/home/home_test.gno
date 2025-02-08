package home

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/r/mouss/config"
)

var (
	user1    = testutils.TestAddress("user1")
	user2    = testutils.TestAddress("user2")
	mainAddr = config.Address()
)

func TestProfile(t *testing.T) {
	uassert.NotEmpty(t, profile.AboutMe, "AboutMe should not be empty")
	uassert.NotEmpty(t, profile.Avatar, "Avatar should not be empty")
	uassert.NotEmpty(t, profile.Email, "Email should not be empty")
	uassert.NotEmpty(t, profile.Github, "Github should not be empty")
	uassert.NotEmpty(t, profile.LinkedIn, "LinkedIn should not be empty")
}

func TestAddRecipe(t *testing.T) {
	std.TestSetOrigCaller(user1)
	name := "Test Recipe"
	origin := "Test Origin"
	ingredients := "Ingredient 1\nIngredient 2"
	instructions := "Step 1\nStep 2"
	tips := "Test Tips"

	result := AddRecipe(name, origin, ingredients, instructions, tips)
	uassert.Equal(t, "Recipe added successfully", result)
	uassert.Equal(t, 1, recipes.Size())
	value, exist := recipes.Get(name)
	uassert.True(t, exist)
	recipe := value.(*Recipe)
	uassert.Equal(t, name, recipe.Name)
	uassert.Equal(t, origin, recipe.Origin)
	uassert.Equal(t, ingredients, recipe.Ingredients)
	uassert.Equal(t, instructions, recipe.Instructions)
	uassert.Equal(t, tips, recipe.Tips)
	uassert.Equal(t, user1, recipe.Author)

	// Verify recipe is correctly stored in AVL tree with matching fields
	var found bool
	recipes.Iterate("", "", func(key string, value interface{}) bool {
		if key == name {
			found = true
			foundRecipe := value.(*Recipe)
			uassert.Equal(t, recipe.Name, foundRecipe.Name)
			uassert.Equal(t, recipe.Origin, foundRecipe.Origin)
			uassert.Equal(t, recipe.Ingredients, foundRecipe.Ingredients)
			uassert.Equal(t, recipe.Instructions, foundRecipe.Instructions)
			uassert.Equal(t, recipe.Tips, foundRecipe.Tips)
			uassert.Equal(t, recipe.Author, foundRecipe.Author)
			return true // stop iteration
		}
		return false
	})
	uassert.Equal(t, true, found)
}

func TestFollow(t *testing.T) {
	// Test user following admin's profile
	std.TestSetOrigCaller(user1)
	err := Follow(user1)
	uassert.NoError(t, err, "regular user should be able to follow admin's profile")

	// Test admin trying to follow themselves
	std.TestSetOrigCaller(mainAddr)
	err = Follow(mainAddr)
	uassert.ErrorIs(t, err, config.ErrUnauthorized, "admin should not be able to follow their own profile")

	// Test following same address twice
	std.TestSetOrigCaller(user1)
	err = Follow(user1)
	uassert.Error(t, err, "should not be able to follow same address twice")

	// Test multiple users following admin
	std.TestSetOrigCaller(user2)
	err = Follow(user2)
	uassert.NoError(t, err, "another user should be able to follow admin's profile")
}

func TestUnfollow(t *testing.T) {
	// user1 is already follower
	// Test successful unfollow
	err := Unfollow(user1)
	uassert.NoError(t, err)
	uassert.False(t, profile.Followers.Has(user1))

	// Test unfollowing when not following
	err = Unfollow(user1)
	uassert.Error(t, err)

	// Test unauthorized unfollow
	std.TestSetOrigCaller(mainAddr)
	err = Unfollow(user1)
	uassert.ErrorIs(t, err, config.ErrUnauthorized)
}

func TestIsUser(t *testing.T) {
	uassert.True(t, isUser(user1), "should recognize normal user")
	uassert.False(t, isUser(mainAddr), "should not recognize admin as normal user")
}
