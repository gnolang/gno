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

func NewAuthorizable(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // TODO(bug, issue #2371): should not be needed

	a := NewAuthorizable()
	got := a.Owner()

	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func NewAuthorizableWithAddress(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)

	got := a.Owner()

	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func CallerOnAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice)

	if err := a.CallerOnAuthList(); err == ErrNotInAuthList {
		t.Fatalf("expected alice to be on the list")
	}
}

func TestNotCallerOnAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOrigCaller(bob)

	if err := a.CallerOnAuthList(); err == nil {
		t.Fatalf("expected bob to not be on the list")
	}
}

func TestAddToAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice)

	if err := a.AddToAuthList(bob); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOrigCaller(bob)

	if err := a.AddToAuthList(bob); err == nil {
		t.Fatalf("Expected AddToAuth to error while bob called it, but it didn't")
	}
}

func TestDelFromList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice)

	if err := a.AddToAuthList(bob); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if err := a.AddToAuthList(charlie); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOrigCaller(bob)

	// Try an unauthorized deletion
	if err := a.DelFromAuthList(alice); err == nil {
		t.Fatalf("Expected DelFromAuth to error with %v", err)
	}

	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice)

	if err := a.DelFromAuthList(charlie); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAssertOnList(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice)
	a := NewAuthorizableWithAddress(alice)

	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOrigCaller(bob)

	uassert.PanicsWithMessage(t, ErrNotInAuthList.Error(), func() {
		a.AssertOnAuthList()
	})

}
