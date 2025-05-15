package authz_test

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/moul/authz"
)

func TestCombinedAuthority_Basic(t *testing.T) {
	// Create mock authorities
	mock1 := newAuthorityMock(func(addr std.Address, action string) bool {
		return action == "action1"
	})
	mock2 := newAuthorityMock(func(addr std.Address, action string) bool {
		return action == "action2"
	})
	mock3 := newAuthorityMock(func(addr std.Address, action string) bool {
		return false
	})

	// Create combined authority
	combined := authz.NewCombinedAuthority(mock1, mock2, mock3)

	// Test cases
	user := std.Address("g1user")
	uassert.True(t, combined.Can(user, "action1"), "should allow action1 from mock1")
	uassert.True(t, combined.Can(user, "action2"), "should allow action2 from mock2")
	uassert.False(t, combined.Can(user, "action3"), "should not allow action3")
}

func TestCombinedAuthority_Empty(t *testing.T) {
	// Test with no authorities
	combined := authz.NewCombinedAuthority()
	user := std.Address("g1user")
	uassert.False(t, combined.Can(user, "any_action"), "empty combined authority should not allow any action")
}

func TestCombinedAuthority_WithAddrsetAuthority(t *testing.T) {
	// Create an AddrsetAuthority
	admin := std.Address("g1admin")
	auth, privcfg := authz.New(admin)
	privcfg.Add("action1", std.Address("g1user1"))

	// Create a mock authority
	mock := newAuthorityMock(func(addr std.Address, action string) bool {
		return action == "action2"
	})

	// Create combined authority
	combined := authz.NewCombinedAuthority(auth, mock)

	// Test cases
	user1 := std.Address("g1user1")
	user2 := std.Address("g1user2")
	uassert.True(t, combined.Can(user1, "action1"), "should allow action1 from AddrsetAuthority")
	uassert.True(t, combined.Can(user2, "action2"), "should allow action2 from mock")
	uassert.False(t, combined.Can(user2, "action1"), "should not allow action1 for user2")
	uassert.False(t, combined.Can(user1, "action3"), "should not allow action3")
}

func TestCombinedAuthority_AdminPermissions(t *testing.T) {
	// Create an AddrsetAuthority with admin
	admin := std.Address("g1admin")
	auth, _ := authz.New(admin)

	// Create a mock authority that allows everything
	mock := newAuthorityMock(func(addr std.Address, action string) bool {
		return true
	})

	// Create combined authority
	combined := authz.NewCombinedAuthority(auth, mock)

	// Test cases
	user := std.Address("g1user")
	uassert.True(t, combined.Can(admin, authz.Admin), "admin should have admin permissions")
	uassert.True(t, combined.Can(user, "any_action"), "user should be allowed by mock")
	uassert.True(t, combined.Can(admin, "any_action"), "admin should be allowed by both")
}

func TestCombinedAuthority_Anyone(t *testing.T) {
	// Create an AddrsetAuthority with Anyone
	auth, privcfg := authz.New()
	privcfg.Add("public_action", authz.Anyone)

	// Create a mock authority that allows nothing
	mock := newAuthorityMock(func(addr std.Address, action string) bool {
		return false
	})

	// Create combined authority
	combined := authz.NewCombinedAuthority(auth, mock)

	// Test cases
	user := std.Address("g1user")
	uassert.True(t, combined.Can(user, "public_action"), "anyone should be allowed for public_action")
	uassert.False(t, combined.Can(user, "private_action"), "no one should be allowed for private_action")
}
