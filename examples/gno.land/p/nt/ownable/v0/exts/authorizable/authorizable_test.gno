package authorizable

import (
	"testing"

	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	charlie = testutils.TestAddress("charlie")
)

func TestNew(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	a := New(ownable.NewWithAddress(alice))
	got := a.Owner()

	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func TestOnAuthList(cur realm, t *testing.T) {
	a := New(ownable.NewWithAddress(alice))
	testing.SetRealm(testing.NewUserRealm(alice))

	// After SetRealm, cur is the test-frame HIV with addr=alice;
	// OnAuthList reports on rlm.Address() (i.e. cur.Address()).
	if err := a.OnAuthList(0, cur); err == ErrNotInAuthList {
		t.Fatalf("expected alice to be on the list")
	}
}

func TestNotOnAuthList(cur realm, t *testing.T) {
	a := New(ownable.NewWithAddress(alice))
	testing.SetRealm(testing.NewUserRealm(bob))

	if err := a.OnAuthList(0, cur); err == nil {
		t.Fatalf("expected bob to not be on the list")
	}
}

func TestAddToAuthList(t *testing.T) {
	a := New(ownable.NewWithAddress(alice))

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	func(cur realm) { err = a.AddToAuthList(0, cur, bob) }(cross)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) { err = a.AddToAuthList(0, cur, bob) }(cross)
	if err == nil {
		t.Fatalf("Expected AddToAuth to error while bob called it, but it didn't")
	}
}

func TestDeleteFromList(t *testing.T) {
	a := New(ownable.NewWithAddress(alice))

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	func(cur realm) { err = a.AddToAuthList(0, cur, bob) }(cross)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	func(cur realm) { err = a.AddToAuthList(0, cur, charlie) }(cross)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Try an unauthorized deletion (bob is not the superuser).
	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) { err = a.DeleteFromAuthList(0, cur, alice) }(cross)
	if err == nil {
		t.Fatalf("Expected DelFromAuth to error with %v", err)
	}

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) { err = a.DeleteFromAuthList(0, cur, charlie) }(cross)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAssertOnList(cur realm, t *testing.T) {
	a := New(ownable.NewWithAddress(alice))

	testing.SetRealm(testing.NewUserRealm(bob))

	uassert.PanicsWithMessage(t, ErrNotInAuthList.Error(), func() {
		a.AssertOnAuthList(0, cur)
	})
}
