package authorizable

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	charlie = testutils.TestAddress("charlie")
)

func TestNewAuthorizable(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(alice))

	a := NewAuthorizable()
	got := a.Owner()

	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func TestNewAuthorizableWithAddress(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)

	got := a.Owner()

	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func TestOnAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	testing.SetRealm(std.NewUserRealm(alice))

	if err := a.OnAuthList(); err == ErrNotInAuthList {
		t.Fatalf("expected alice to be on the list")
	}
}

func TestNotOnAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	testing.SetRealm(std.NewUserRealm(bob))

	if err := a.OnAuthList(); err == nil {
		t.Fatalf("expected bob to not be on the list")
	}
}

func TestAddToAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	testing.SetRealm(std.NewUserRealm(alice))

	if err := a.AddToAuthList(bob); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	testing.SetRealm(std.NewUserRealm(bob))

	if err := a.AddToAuthList(bob); err == nil {
		t.Fatalf("Expected AddToAuth to error while bob called it, but it didn't")
	}
}

func TestDeleteFromList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	testing.SetRealm(std.NewUserRealm(alice))

	if err := a.AddToAuthList(bob); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if err := a.AddToAuthList(charlie); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	testing.SetRealm(std.NewUserRealm(bob))

	// Try an unauthorized deletion
	if err := a.DeleteFromAuthList(alice); err == nil {
		t.Fatalf("Expected DelFromAuth to error with %v", err)
	}

	testing.SetRealm(std.NewUserRealm(alice))

	if err := a.DeleteFromAuthList(charlie); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAssertOnList(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(alice))

	a := NewAuthorizableWithAddress(alice)

	testing.SetRealm(std.NewUserRealm(bob))

	uassert.PanicsWithMessage(t, ErrNotInAuthList.Error(), func() {
		a.AssertOnAuthList()
	})
}
