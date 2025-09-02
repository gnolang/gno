package tests

import (
	"std"
	"testing"
)

func TestNestedPkg(t *testing.T) {
	// direct child
	cur := "gno.land/r/tests/foo"
	testing.SetRealm(std.NewCodeRealm(cur))
	if !IsCallerSubPath(cross) {
		t.Errorf(cur + " should be a sub path")
	}
	if IsCallerParentPath(cross) {
		t.Errorf(cur + " should not be a parent path")
	}
	if !HasCallerSameNamespace(cross) {
		t.Errorf(cur + " should be from the same namespace")
	}

	// grand-grand-child
	cur = "gno.land/r/tests/foo/bar/baz"
	testing.SetRealm(std.NewCodeRealm(cur))
	if !IsCallerSubPath(cross) {
		t.Errorf(cur + " should be a sub path")
	}
	if IsCallerParentPath(cross) {
		t.Errorf(cur + " should not be a parent path")
	}
	if !HasCallerSameNamespace(cross) {
		t.Errorf(cur + " should be from the same namespace")
	}

	// NOTE: After moving from gno.land/r/tests/tests to gno.land/r/tests,
	// the "direct parent" test case no longer makes sense as we are now testing
	// the same path as the current package itself.
	// Skipping this test case.

	// direct parent (was previously fake parent)
	cur = "gno.land/r/test" // without the 's' at the end
	testing.SetRealm(std.NewCodeRealm(cur))
	if IsCallerSubPath(cross) {
		t.Errorf(cur + " should not be a sub path")
	}
	if IsCallerParentPath(cross) {
		t.Errorf(cur + " should not be a parent path")
	}
	if HasCallerSameNamespace(cross) {
		t.Errorf(cur + " should not be from the same namespace")
	}

	// fake parent (prefix)
	cur = "gno.land/r/dem"
	testing.SetRealm(std.NewCodeRealm(cur))
	if IsCallerSubPath(cross) {
		t.Errorf(cur + " should not be a sub path")
	}
	if IsCallerParentPath(cross) {
		t.Errorf(cur + " should not be a parent path")
	}
	if HasCallerSameNamespace(cross) {
		t.Errorf(cur + " should not be from the same namespace")
	}

	// different namespace
	cur = "gno.land/r/foo"
	testing.SetRealm(std.NewCodeRealm(cur))
	if IsCallerSubPath(cross) {
		t.Errorf(cur + " should not be a sub path")
	}
	if IsCallerParentPath(cross) {
		t.Errorf(cur + " should not be a parent path")
	}
	if HasCallerSameNamespace(cross) {
		t.Errorf(cur + " should not be from the same namespace")
	}
}
