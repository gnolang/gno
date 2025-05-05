package home

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/r/mouss/config"
)

var (
	user1    = testutils.TestAddress("user1")
	user2    = testutils.TestAddress("user2")
	mainAddr = config.OwnableMain.Owner()
)

func TestProfile(t *testing.T) {
	uassert.NotEmpty(t, profile.AboutMe, "AboutMe should not be empty")
	uassert.NotEmpty(t, profile.Avatar, "Avatar should not be empty")
	uassert.NotEmpty(t, profile.Email, "Email should not be empty")
	uassert.NotEmpty(t, profile.Github, "Github should not be empty")
	uassert.NotEmpty(t, profile.LinkedIn, "LinkedIn should not be empty")
}

func TestAddRecipe(t *testing.T) {
	testing.SetOriginCaller(user1)
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
			return true
		}
		return false
	})
	uassert.Equal(t, true, found)
}

func TestFollow(t *testing.T) {
	// Test user following admin's profile
	testing.SetOriginCaller(user1)
	err := Follow()
	uassert.NoError(t, err, "user should be able to follow admin's profile")

	// Test admin trying to follow themselves
	testing.SetOriginCaller(mainAddr)
	err = Follow()
	uassert.Error(t, err, "you cannot follow yourself")

	// Test following same address twice
	testing.SetOriginCaller(user1)
	err = Follow()
	uassert.Error(t, err, "should not be able to follow same address twice")

	// Test multiple users following admin
	testing.SetOriginCaller(user2)
	err = Follow()
	uassert.NoError(t, err, "another user should be able to follow admin's profile")
}

func TestUnfollow(t *testing.T) {
	// Test successful unfollow
	testing.SetOriginCaller(user1)
	err := Unfollow()
	uassert.NoError(t, err)
	uassert.False(t, profile.Followers.Has(user1))

	// Test unfollowing when not following
	err = Unfollow()
	uassert.Error(t, err)

}
