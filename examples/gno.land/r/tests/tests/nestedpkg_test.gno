package tests

import (
	"std"
	"testing"
)

func TestNestedPkg(t *testing.T) {
	// direct child
	cur := "gno.land/r/tests/tests/foo"
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
	cur = "gno.land/r/tests/tests/foo/bar/baz"
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

	// direct parent
	cur = "gno.land/r/tests"
	testing.SetRealm(std.NewCodeRealm(cur))
	if IsCallerSubPath(cross) {
		t.Errorf(cur + " should not be a sub path")
	}
	if !IsCallerParentPath(cross) {
		t.Errorf(cur + " should be a parent path")
	}
	if !HasCallerSameNamespace(cross) {
		t.Errorf(cur + " should be from the same namespace")
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
