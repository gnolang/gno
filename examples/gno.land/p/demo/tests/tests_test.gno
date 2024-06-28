package tests_test

import (
	"std"
	"testing"

	"gno.land/p/demo/tests"
	"gno.land/p/demo/testutils"
)

var World = "WORLD"

func TestGetHelloWorld(t *testing.T) {
	// tests.World is 'world'
	s := "hello " + tests.World + World
	const want = "hello worldWORLD"
	if s != want {
		t.Errorf("got %q want %q", s, want)
	}
}

func TestSetRealm_UserRealm(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetRealm(std.NewUserRealm(alice))

	var got std.Address
	proxyRealm(func() {
		got = tests.GetPrevRealm().Addr()
	})

	if alice != got {
		t.Fatalf("expected %s, got %s", alice, got)
	}
}

func TestSetRealm_CodeRealm(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/users"))
	expected := "g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s" // derived from the path above

	var got std.Address
	proxyRealm(func() {
		got = tests.GetPrevRealm().Addr()
	})

	if expected != got.String() {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func proxyRealm(f func()) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/proxy"))
	f()
}
