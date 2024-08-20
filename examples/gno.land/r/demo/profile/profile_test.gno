package profile

import (
	"std"
	"testing"
	"errors"

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

func TestSetMultiFields(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(user1))

	// Set new arbitrary string field
	err := SetField("NewStrField", "This is new string value")
	uassert.NoError(t, err)

	strVal := GetField(user1, "NewStrField", "")
	uassert.Equal(t, strVal , "This is new string value")

	// Set new arbitrary int field
	err = SetField("NewIntField", 40)
	uassert.NoError(t, err)

	intVal := GetField(user1, "NewIntField", 0)
	uassert.Equal(t, intVal , 40)

	// Set new arbitrary bool field
	err = SetField("NewBoolField", true)
	uassert.NoError(t, err)

	boolVal := GetField(user1, "NewBoolField", false)
	uassert.Equal(t, boolVal , true)

	// Set reserved field
	err = SetField("DisplayName", "My Name")
	uassert.NoError(t, err)

	reservedStrVal := GetStringField(user1, "DisplayName", "")
	uassert.Equal(t, reservedStrVal , "My Name")

	// Get reserved field
	err = SetIntField("Age", 40)
	uassert.NoError(t, err)

	reservedIntVal := GetField(user1, "Age", 0)
	uassert.Equal(t, reservedIntVal , 40)

	// Get string field that does not exist should return default string
	inexistStrVal := GetField(user1, "InexistStrField", "default")
	uassert.Equal(t, inexistStrVal , "default")

	// Get int field that does not exist should return default int
	inexistIntVal := GetField(user1, "inexistIntField", 0)
	uassert.Equal(t, inexistIntVal , 0)

	// Get bool field that does not exist should return default bool
	inexistBoolVal := GetField(user1, "inexistBoolField", false)
	uassert.Equal(t, inexistBoolVal , false)
}