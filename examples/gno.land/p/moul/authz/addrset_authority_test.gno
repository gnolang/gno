package authz_test

/*

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/moul/authz"
)

func Example_basic() {
	// create a new addrset authority.
	// Auth is the safe-object authority implementing the Can() method and exposing Add(), Remove() and RemoveAll() that an admin can call to change the memberset.
	// privcfg is the unrestricted object allowing this contract to Add(), Remove(), and RemoveAll().
	Auth, privcfg := authz.New()
	privcfg.Add(authz.Admin, "g1admin")
	privcfg.Add("propose", "g1user1", "g1user2")
	privcfg.Add("vote", std.CurrentRealm().Address())
	// privcfg.Add("veto", externalAlreadyDefinedAuthority)

	/*
		if Auth.CanByCurrent(authz.Admin) {
			print("hello")
		}
		if Auth.Can(std.CurrentRealm().Address(), authz.Admin) {
			print("world")
		}
		Auth.AssertCurrentCan("arbitrary action")
		println("!")
		// Output: hello world!
*/ /*
}

func TestAddrsetAuthority_Basic(t *testing.T) {
	admin := std.Address("g1admin")
	user1 := std.Address("g1user1")
	user2 := std.Address("g1user2")
	user3 := std.Address("g1user3")
	user4 := std.Address("g1user4")
	// Create a new authority with an admin
	auth, privcfg := authz.New(admin)

	// Add some test users and permissions
	privcfg.Add("propose", user1, user2)
	privcfg.Add("vote", user3)
	privcfg.Add("public_action", authz.Anyone)

	// Test admin permissions
	uassert.True(t, auth.Can(admin, authz.Admin), "admin should be able to perform admin actions")
	uassert.True(t, auth.Can(admin, "propose"), "admin should be able to perform propose action")
	uassert.True(t, auth.Can(admin, "arbitrary_action"), "admin should be able to perform arbitrary action")
	uassert.True(t, auth.Can(admin, "public_action"), "admin should be able to perform public action")

	// Test user permissions
	uassert.True(t, auth.Can(user1, "propose"), "user1 should be able to propose")
	uassert.False(t, auth.Can(user1, "vote"), "user1 should not be able to vote")
	uassert.False(t, auth.Can(user1, "arbitrary_action"), "user1 should not be able to perform arbitrary action")
	uassert.True(t, auth.Can(user1, "public_action"), "user1 should be able to perform public action")
	uassert.False(t, auth.Can(user1, authz.Admin), "user1 should not be able to perform admin action")

	// Test Anyone special address
	uassert.True(t, auth.Can(user4, "public_action"), "anyone should be able to perform public actions")
	uassert.False(t, auth.Can(user4, "arbitrary_action"), "anyone should not be able to perform arbitrary action")
	uassert.False(t, auth.Can(user4, "vote"), "anyone should not be able to vote")
}

func TestAddrsetAuthority_String(t *testing.T) {
	// Create a new authority with multiple users and actions
	admin := std.Address("g1admin")
	auth, privcfg := authz.New(admin)

	// Add multiple users to different actions
	user1 := std.Address("g1user1")
	user2 := std.Address("g1user2")
	user3 := std.Address("g1user3")
	user4 := std.Address("g1user4")
	privcfg.Add("action1", user1)
	privcfg.Add("action2", user1, user2)
	privcfg.Add("action3", user1, user2, user3, user4)
	privcfg.Add("action4", authz.Anyone)

	expected := `_admin_: g1admin
action1: g1user1
action2: g1user1, g1user2
action3: g1user1, g1user2, g1user3, g1user4
action4: g1anyone`
	uassert.Equal(t, expected, auth.String())
}

func TestAddrsetAuthority_NewWithCurrent(t *testing.T) {
	auth, _ := authz.NewWithCurrent()
	uassert.True(t, auth.Can(std.CurrentRealm().Address(), authz.Admin), "current realm should be admin")
}

func TestAddrsetAuthority_NewWithPrevious(t *testing.T) {
	auth, _ := authz.NewWithPrevious()
	uassert.True(t, auth.Can(std.PreviousRealm().Address(), authz.Admin), "previous realm should be admin")
}

func TestAddrsetAuthority_AddRemove(t *testing.T) {
	admin := std.Address("g1admin")
	user := std.Address("g1user")
	auth, privcfg := authz.New(admin)

	// Test adding and removing a single user
	privcfg.Add("test_action", user)
	uassert.True(t, auth.Can(user, "test_action"), "user should be able to perform test_action after adding")

	privcfg.Remove("test_action", user)
	uassert.False(t, auth.Can(user, "test_action"), "user should not be able to perform test_action after removing")

	// Test adding and removing multiple users
	user2 := std.Address("g1user2")
	privcfg.Add("test_action", user, user2)
	uassert.True(t, auth.Can(user, "test_action") && auth.Can(user2, "test_action"), "both users should be able to perform test_action")

	privcfg.Remove("test_action", user)
	uassert.False(t, auth.Can(user, "test_action"), "user1 should not be able to perform test_action after removing")
	uassert.True(t, auth.Can(user2, "test_action"), "user2 should still be able to perform test_action")
}

func TestAddrsetAuthority_RemoveAll(t *testing.T) {
	admin := std.Address("g1admin")
	user1 := std.Address("g1user1")
	user2 := std.Address("g1user2")
	auth, privcfg := authz.New(admin)

	// Test removing all from non-existent action
	privcfg.RemoveAll("non_existent") // Should not panic

	// Test removing all from action with multiple users
	privcfg.Add("test_action", user1, user2)
	uassert.True(t, auth.Can(user1, "test_action") && auth.Can(user2, "test_action"), "both users should be able to perform test_action")

	privcfg.RemoveAll("test_action")
	uassert.False(t, auth.Can(user1, "test_action") || auth.Can(user2, "test_action"), "no users should be able to perform test_action after RemoveAll")
}

func TestAddrsetAuthority_AdminOperations(t *testing.T) {
	admin := std.Address("g1admin")
	nonAdmin := std.Address("g1nonadmin")
	user := std.Address("g1user")
	auth, _ := authz.New(admin)

	// Test that non-admin can't add
	uassert.PanicsWithMessage(t, "unauthorized: only admin can add members", func() {
		auth.Add("test_action", user)
	})

	// Test that non-admin can't remove
	uassert.PanicsWithMessage(t, "unauthorized: only admin can remove members", func() {
		auth.Remove("test_action", user)
	})

	// Test that non-admin can't remove all
	uassert.PanicsWithMessage(t, "unauthorized: only admin can remove all policies", func() {
		auth.RemoveAll("test_action")
	})
}

func TestAddrsetAuthority_EmptyAuthority(t *testing.T) {
	auth, _ := authz.New()
	user := std.Address("g1user")

	// Test that no one can perform any action in empty authority
	uassert.False(t, auth.Can(user, "any_action"), "no one should be able to perform any action in empty authority")
	uassert.False(t, auth.Can(user, authz.Admin), "no one should be able to perform admin action in empty authority")
}

func TestAddrsetAuthority_MultipleAdmins(t *testing.T) {
	admin1 := std.Address("g1admin1")
	admin2 := std.Address("g1admin2")
	user := std.Address("g1user")
	auth, _ := authz.New(admin1, admin2)

	// Test that both admins can perform admin actions
	uassert.True(t, auth.Can(admin1, authz.Admin) && auth.Can(admin2, authz.Admin), "both admins should be able to perform admin actions")

	testing.SetRealm(std.NewUserRealm(admin1))
	// Test that both admins can add/remove
	auth.Add("test_action", user)
	uassert.True(t, auth.Can(user, "test_action"), "user should be able to perform test_action after admin1 adds")

	auth.Remove("test_action", user)
	uassert.False(t, auth.Can(user, "test_action"), "user should not be able to perform test_action after admin2 removes")
}
*/
