package vm

import (
	"testing"
)

func TestNestedPkg(t *testing.T) {
	// direct child
	cur := "gno.land/r/tests/vm/foo"
	testing.SetRealm(testing.NewCodeRealm(cur))
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
	cur = "gno.land/r/tests/vm/foo/bar/baz"
	testing.SetRealm(testing.NewCodeRealm(cur))
	if !IsCallerSubPath(cross) {
		t.Errorf(cur + " should be a sub path")
	}
	if IsCallerParentPath(cross) {
		t.Errorf(cur + " should not be a parent path")
	}
	if !HasCallerSameNamespace(cross) {
		t.Errorf(cur + " should be from the same namespace")
	}

	// NOTE: This is now back in the gno.land/r/tests/vm structure,
	// so the direct parent test case is valid again.

	// direct parent (was previously fake parent)
	cur = "gno.land/r/test" // without the 's' at the end
	testing.SetRealm(testing.NewCodeRealm(cur))
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
	testing.SetRealm(testing.NewCodeRealm(cur))
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
	testing.SetRealm(testing.NewCodeRealm(cur))
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
