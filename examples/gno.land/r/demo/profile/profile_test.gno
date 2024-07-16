package profile

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
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
	std.TestSetRealm(std.NewUserRealm(alice))

	// Get before setting
	name := GetStringField(alice, DisplayName, "anon")
	uassert.Equal(t, "anon", name)

	// Set
	err := SetStringField(DisplayName, "Alice foo")
	uassert.NoError(t, err)
	err = SetStringField(Homepage, "https://example.com")
	uassert.NoError(t, err)

	// Get after setting
	name = GetStringField(alice, DisplayName, "anon")
	homepage := GetStringField(alice, Homepage, "")
	bio := GetStringField(alice, Bio, "42")

	uassert.Equal(t, "Alice foo", name)
	uassert.Equal(t, "https://example.com", homepage)
	uassert.Equal(t, "42", bio)
}

func TestIntFields(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(bob))

	// Get before setting
	age := GetIntField(bob, Age, 25)
	uassert.Equal(t, 25, age)

	// Set
	err := SetIntField(Age, 30)
	uassert.NoError(t, err)

	// Get after setting
	age = GetIntField(bob, Age, 25)
	uassert.Equal(t, 30, age)
}

func TestBoolFields(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(charlie))

	// Get before setting
	hiring := GetBoolField(charlie, AvailableForHiring, false)
	uassert.Equal(t, false, hiring)

	// Set
	err := SetBoolField(AvailableForHiring, true)
	uassert.NoError(t, err)

	// Get after setting
	hiring = GetBoolField(charlie, AvailableForHiring, false)
	uassert.Equal(t, true, hiring)
}

func TestInvalidStringField(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(dave))

	err := SetStringField(InvalidField, "test")
	uassert.Error(t, err)
}

func TestInvalidIntField(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(eve))

	err := SetIntField(InvalidField, 123)
	uassert.Error(t, err)
}

func TestInvalidBoolField(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(frank))

	err := SetBoolField(InvalidField, true)
	uassert.Error(t, err)
}

func TestMultipleProfiles(t *testing.T) {
	// Set profile for user1
	std.TestSetRealm(std.NewUserRealm(user1))
	err := SetStringField(DisplayName, "User One")
	uassert.NoError(t, err)

	// Set profile for user2
	std.TestSetRealm(std.NewUserRealm(user2))
	err = SetStringField(DisplayName, "User Two")
	uassert.NoError(t, err)

	// Get profiles
	std.TestSetRealm(std.NewUserRealm(user1)) // Switch back to user1
	name1 := GetStringField(user1, DisplayName, "anon")
	std.TestSetRealm(std.NewUserRealm(user2)) // Switch back to user2
	name2 := GetStringField(user2, DisplayName, "anon")

	uassert.Equal(t, "User One", name1)
	uassert.Equal(t, "User Two", name2)
}
