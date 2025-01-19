package v1

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"

	"gno.land/r/sys/users"
)

var (
	alice     = "alice123"
	bob       = "bob123"
	aliceAddr = testutils.TestAddress(alice)
	bobAddr   = testutils.TestAddress(bob)
)

func TestRegister_Valid(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(aliceAddr))
	std.TestSetOrigCaller(aliceAddr)

	uassert.NoError(t, Register(alice))
	res, latest := users.ResolveName(alice)

	uassert.NotEqual(t, nil, res)
	uassert.Equal(t, alice, res.Name())
	uassert.Equal(t, aliceAddr, res.Addr())
	uassert.False(t, res.IsDeleted())
	uassert.True(t, latest)
}

func TestRegister_Invalid(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(bobAddr))
	std.TestSetOrigCaller(bobAddr)

	// Invalid usernames
	uassert.Error(t, Register("alice"), ErrInvalidUsername.Error())      // vanity
	uassert.Error(t, Register(""), ErrInvalidUsername.Error())           // empty
	uassert.Error(t, Register("   "), ErrInvalidUsername.Error())        // empty
	uassert.Error(t, Register("123"), ErrInvalidUsername.Error())        // only numbers
	uassert.Error(t, Register("alice&#($)"), ErrInvalidUsername.Error()) // non-allowed chars
	uassert.Error(t, Register("Alice123"), ErrInvalidUsername.Error())   // upper-case
	uassert.Error(t, Register("toolongusernametoolongusernametoolongusername123"),
		ErrInvalidUsername.Error()) // too long

	// Name taken
	urequire.NoError(t, Register(bob))
	uassert.Error(t, Register(bob), users.ErrNameTaken.Error())
}

func TestUpdateName_Valid(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(aliceAddr))
	std.TestSetOrigCaller(aliceAddr)

	newalice := "newalice123"
	// resolve old name
	urequire.NoError(t, UpdateName(newalice))

	res, latest := users.ResolveName(alice)
	uassert.NotEqual(t, nil, res)
	uassert.Equal(t, newalice, res.Name())
	uassert.Equal(t, aliceAddr, res.Addr())
	uassert.False(t, res.IsDeleted())
	uassert.False(t, latest)
}

func TestUpdateName_Invalid(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(aliceAddr))
	std.TestSetOrigCaller(aliceAddr)

	// Invalid names
	uassert.Error(t, UpdateName("alice"), ErrInvalidUsername.Error())      // vanity
	uassert.Error(t, UpdateName(""), ErrInvalidUsername.Error())           // empty
	uassert.Error(t, UpdateName("   "), ErrInvalidUsername.Error())        // empty
	uassert.Error(t, UpdateName("123"), ErrInvalidUsername.Error())        // only numbers
	uassert.Error(t, UpdateName("alice&#($)"), ErrInvalidUsername.Error()) // non-allowed chars
	uassert.Error(t, UpdateName("Alice123"), ErrInvalidUsername.Error())   // upper-case
	uassert.Error(t, UpdateName("toolongusernametoolongusernametoolongusername123"),
		ErrInvalidUsername.Error()) // too long

	urequire.Error(t, UpdateName(bob), users.ErrNameTaken.Error())
}

func TestDeleteUser(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(aliceAddr))
	std.TestSetOrigCaller(aliceAddr)

	urequire.NoError(t, DeleteUser())
	res, _ := users.ResolveName(alice)
	uassert.Equal(t, nil, res)
	res = users.ResolveAddress(aliceAddr)
	uassert.Equal(t, nil, res)
}

func TestDeleteUser_Invalid(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(aliceAddr))
	std.TestSetOrigCaller(aliceAddr)

	// Already deleted user
	urequire.Error(t, DeleteUser(), users.ErrUserNotExist.Error())
}
