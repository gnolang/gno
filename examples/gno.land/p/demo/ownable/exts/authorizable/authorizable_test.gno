package authorizable

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	charlie = testutils.TestAddress("charlie")
)

func TestNew(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // TODO(bug): should not be needed

	a := NewAuthorizable()
	got := a.Owner()

	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func TestNewWithAddress(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)

	got := a.Owner()

	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func TestOnList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice)

	if err := a.OnAuthList(); err != nil {
		t.Fatalf("expected alice to be on the list")
	}
}

func TestNotOnList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOrigCaller(bob)

	if err := a.OnAuthList(); err == nil {
		t.Fatalf("expected bob to not be on the list")
	}
}

func TestAddToAuthList(t *testing.T) {
	a := NewAuthorizableWithAddress(alice)
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice)

	if err := a.AddToAuthList(bob); err != nil {
		println(err)
		t.Fatalf("Expected AddToAuth to not error while adding bob, but it did")
	}

	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOrigCaller(bob)

	if err := a.AddToAuthList(bob); err == nil {
		t.Fatalf("Expected AddToAuth to error while bob called it, but it didn't")
	}

	//if err := a.AddToAuthList(charlie); err == nil {
	//	t.Fatalf("Expected AddToAuth to error, did not happen")
	//}
}
