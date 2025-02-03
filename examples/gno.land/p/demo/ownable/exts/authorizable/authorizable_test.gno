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
	t.SetOriginCaller(alice)

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

func TestCallerOnAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	t.SetOriginCaller(alice)

	if err := a.CallerOnAuthList(); err == ErrNotInAuthList {
		t.Fatalf("expected alice to be on the list")
	}
}

func TestNotCallerOnAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	t.SetOriginCaller(bob)

	if err := a.CallerOnAuthList(); err == nil {
		t.Fatalf("expected bob to not be on the list")
	}
}

func TestAddToAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	t.SetOriginCaller(alice)

	if err := a.AddToAuthList(bob); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	t.SetOriginCaller(bob)

	if err := a.AddToAuthList(bob); err == nil {
		t.Fatalf("Expected AddToAuth to error while bob called it, but it didn't")
	}
}

func TestDeleteFromList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	t.SetOriginCaller(alice)

	if err := a.AddToAuthList(bob); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if err := a.AddToAuthList(charlie); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	t.SetOriginCaller(bob)

	// Try an unauthorized deletion
	if err := a.DeleteFromAuthList(alice); err == nil {
		t.Fatalf("Expected DelFromAuth to error with %v", err)
	}

	t.SetOriginCaller(alice)

	if err := a.DeleteFromAuthList(charlie); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAssertOnList(t *testing.T) {
	t.SetOriginCaller(alice)

	a := NewAuthorizableWithAddress(alice)

	t.SetOriginCaller(bob)

	uassert.PanicsWithMessage(t, ErrNotInAuthList.Error(), func() {
		a.AssertOnAuthList()
	})
}
