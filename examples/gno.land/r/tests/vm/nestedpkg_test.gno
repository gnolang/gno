package vm

import (
	"testing"
)

func TestNestedPkg(cur realm, t *testing.T) {
	// direct child
	curPath := "gno.land/r/tests/vm/foo"
	testing.SetRealm(testing.NewCodeRealm(curPath))
	if !IsCallerSubPath(cross(cur)) {
		t.Errorf(curPath + " should be a sub path")
	}
	if IsCallerParentPath(cross(cur)) {
		t.Errorf(curPath + " should not be a parent path")
	}
	if !HasCallerSameNamespace(cross(cur)) {
		t.Errorf(curPath + " should be from the same namespace")
	}

	// grand-grand-child
	curPath = "gno.land/r/tests/vm/foo/bar/baz"
	testing.SetRealm(testing.NewCodeRealm(curPath))
	if !IsCallerSubPath(cross(cur)) {
		t.Errorf(curPath + " should be a sub path")
	}
	if IsCallerParentPath(cross(cur)) {
		t.Errorf(curPath + " should not be a parent path")
	}
	if !HasCallerSameNamespace(cross(cur)) {
		t.Errorf(curPath + " should be from the same namespace")
	}

	// NOTE: This is now back in the gno.land/r/tests/vm structure,
	// so the direct parent test case is valid again.

	// direct parent (was previously fake parent)
	curPath = "gno.land/r/test" // without the 's' at the end
	testing.SetRealm(testing.NewCodeRealm(curPath))
	if IsCallerSubPath(cross(cur)) {
		t.Errorf(curPath + " should not be a sub path")
	}
	if IsCallerParentPath(cross(cur)) {
		t.Errorf(curPath + " should not be a parent path")
	}
	if HasCallerSameNamespace(cross(cur)) {
		t.Errorf(curPath + " should not be from the same namespace")
	}

	// fake parent (prefix)
	curPath = "gno.land/r/dem"
	testing.SetRealm(testing.NewCodeRealm(curPath))
	if IsCallerSubPath(cross(cur)) {
		t.Errorf(curPath + " should not be a sub path")
	}
	if IsCallerParentPath(cross(cur)) {
		t.Errorf(curPath + " should not be a parent path")
	}
	if HasCallerSameNamespace(cross(cur)) {
		t.Errorf(curPath + " should not be from the same namespace")
	}

	// different namespace
	curPath = "gno.land/r/foo"
	testing.SetRealm(testing.NewCodeRealm(curPath))
	if IsCallerSubPath(cross(cur)) {
		t.Errorf(curPath + " should not be a sub path")
	}
	if IsCallerParentPath(cross(cur)) {
		t.Errorf(curPath + " should not be a parent path")
	}
	if HasCallerSameNamespace(cross(cur)) {
		t.Errorf(curPath + " should not be from the same namespace")
	}
}
