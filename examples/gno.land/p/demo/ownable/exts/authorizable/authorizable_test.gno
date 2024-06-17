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

//func TestAddToAuthList(t *testing.T) {
//	a := NewAuthorizable()
//
//	if err := a.AddToAuthList(bob); err != nil {
//		t.Fatalf("Expected AddToAuth to not error, but it did")
//	}
//
//	std.TestSetRealm(std.NewUserRealm(bob))
//
//	if err := a.AddToAuthList(charlie); err == nil {
//		t.Fatalf("Expected AddToAuth to error, did not happen")
//	}
//}
