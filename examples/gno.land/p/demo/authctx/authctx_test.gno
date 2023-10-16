package authctx

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestNewOrigAuthContext(t *testing.T) {
	var test1 std.Address = testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)

	auth := NewOrigAuthContext()

	assertSameAddresses(t, test1, auth.Addr())
	assertSameStrings(t, "orig:"+string(test1), auth.String())
	if !auth.IsAuthorized("foo") {
		t.Errorf("should be authorized")
	}
}

func TestNewPrevAuthContext(t *testing.T) {
	var test1 std.Address = testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)

	auth := NewPrevAuthContext()

	assertSameAddresses(t, test1, auth.Addr())
	assertSameStrings(t, "prev:"+string(test1), auth.String())
	if !auth.IsAuthorized("foo") {
		t.Errorf("should be authorized")
	}
}

func TestPackage(t *testing.T) {
	// table test
	// - NewOrig
	// - NewPrev
	// - NewScoped
	// - owner is a contract
	// - non-address (ID, not Addr)

	/*	test1 := testutils.TestAddress("test1")
		test2 := testutils.TestAddress("test2")
		test3 := testutils.TestAddress("test3")
		test4 := testutils.TestAddress("test4")

		var db authctx.DelegationDB

		std.TestSetOrigCaller(test1)
		var customAuthFn authctx.AuthFunc = func(action string) bool { return action == "fooperm" }
		db.Approve(test2, customAuthFn)
		db.Approve(test2, authctx.AuthorizeAll())
		db.Approve(test3, authctx.AuthorizeNone())
		db.Approve(test4, authctx.AuthorizeActions("fooperm", "barperm"))

		{
			std.TestSetOrigCaller(test2)
			delegation := db.NewOrigDelegatedAuthContext(test1)
			println("test2", delegation, delegation.IsAuthorized("bazperm"))
		}

		{
			std.TestSetOrigCaller(test3)
			delegation := db.NewOrigDelegatedAuthContext(test1)
			println("test3", delegation, delegation.IsAuthorized("bazperm"))
		}

		// TODO: test db.NewPrevDelegatedAuthContext
	*/
}

// TODO: TestDelegations

func assertSameStrings(t *testing.T, expected, got string) {
	t.Helper()
	if expected != got {
		t.Errorf("expected %q, got %q.", expected, got)
	}
}

func assertSameAddresses(t *testing.T, expected, got std.Address) {
	t.Helper()
	if expected != got {
		t.Errorf("expected %q, got %q.", expected, got)
	}
}
