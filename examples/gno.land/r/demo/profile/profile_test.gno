package profile

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// Global addresses for test users
var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	charlie = testutils.TestAddress("charlie")
	dave    = testutils.TestAddress("dave")
	eve     = testutils.TestAddress("eve")
	frank   = testutils.TestAddress("frank")
	user1   = testutils.TestAddress("user1")
	user2   = testutils.TestAddress("user2")
)

func TestStringFields(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	// Get before setting
	name := GetStringField(alice, DisplayName, "anon")
	uassert.Equal(t, "anon", name)

	// Set new key
	updated := SetStringField(cross, DisplayName, "Alice foo")
	uassert.Equal(t, updated, false)
	updated = SetStringField(cross, Homepage, "https://example.com")
	uassert.Equal(t, updated, false)

	// Update the key
	updated = SetStringField(cross, DisplayName, "Alice foo")
	uassert.Equal(t, updated, true)

	// Get after setting
	name = GetStringField(alice, DisplayName, "anon")
	homepage := GetStringField(alice, Homepage, "")
	bio := GetStringField(alice, Bio, "42")

	uassert.Equal(t, "Alice foo", name)
	uassert.Equal(t, "https://example.com", homepage)
	uassert.Equal(t, "42", bio)
}

func TestIntFields(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(bob))

	// Get before setting
	age := GetIntField(bob, Age, 25)
	uassert.Equal(t, 25, age)

	// Set new key
	updated := SetIntField(cross, Age, 30)
	uassert.Equal(t, updated, false)

	// Update the key
	updated = SetIntField(cross, Age, 30)
	uassert.Equal(t, updated, true)

	// Get after setting
	age = GetIntField(bob, Age, 25)
	uassert.Equal(t, 30, age)
}

func TestBoolFields(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(charlie))

	// Get before setting
	hiring := GetBoolField(charlie, AvailableForHiring, false)
	uassert.Equal(t, false, hiring)

	// Set
	updated := SetBoolField(cross, AvailableForHiring, true)
	uassert.Equal(t, updated, false)

	// Update the key
	updated = SetBoolField(cross, AvailableForHiring, true)
	uassert.Equal(t, updated, true)

	// Get after setting
	hiring = GetBoolField(charlie, AvailableForHiring, false)
	uassert.Equal(t, true, hiring)
}

func TestMultipleProfiles(t *testing.T) {
	// Set profile for user1
	testing.SetRealm(testing.NewUserRealm(user1))
	updated := SetStringField(cross, DisplayName, "User One")
	uassert.Equal(t, updated, false)

	// Set profile for user2
	testing.SetRealm(testing.NewUserRealm(user2))
	updated = SetStringField(cross, DisplayName, "User Two")
	uassert.Equal(t, updated, false)

	// Get profiles
	testing.SetRealm(testing.NewUserRealm(user1)) // Switch back to user1
	name1 := GetStringField(user1, DisplayName, "anon")
	testing.SetRealm(testing.NewUserRealm(user2)) // Switch back to user2
	name2 := GetStringField(user2, DisplayName, "anon")

	uassert.Equal(t, "User One", name1)
	uassert.Equal(t, "User Two", name2)
}

func TestArbitraryStringField(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(user1))

	// Set arbitrary string field
	updated := SetStringField(cross, "MyEmail", "my@email.com")
	uassert.Equal(t, updated, false)

	val := GetStringField(user1, "MyEmail", "")
	uassert.Equal(t, val, "my@email.com")
}

func TestArbitraryIntField(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(user1))

	// Set arbitrary int field
	updated := SetIntField(cross, "MyIncome", 100_000)
	uassert.Equal(t, updated, false)

	val := GetIntField(user1, "MyIncome", 0)
	uassert.Equal(t, val, 100_000)
}

func TestArbitraryBoolField(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(user1))

	// Set arbitrary bool field
	updated := SetBoolField(cross, "IsWinner", true)
	uassert.Equal(t, updated, false)

	val := GetBoolField(user1, "IsWinner", false)
	uassert.Equal(t, val, true)
}
