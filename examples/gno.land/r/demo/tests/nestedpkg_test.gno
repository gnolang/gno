package tests

import (
	"std"
	"testing"
)

func TestNestedPkg(t *testing.T) {
	// direct child
	cur := "gno.land/r/demo/tests/foo"
	testing.SetRealm(std.NewCodeRealm(cur))
	if !cross(IsCallerSubPath)() {
		t.Errorf(cur + " should be a sub path")
	}
	if cross(IsCallerParentPath)() {
		t.Errorf(cur + " should not be a parent path")
	}
	if !cross(HasCallerSameNamespace)() {
		t.Errorf(cur + " should be from the same namespace")
	}

	// grand-grand-child
	cur = "gno.land/r/demo/tests/foo/bar/baz"
	testing.SetRealm(std.NewCodeRealm(cur))
	if !cross(IsCallerSubPath)() {
		t.Errorf(cur + " should be a sub path")
	}
	if cross(IsCallerParentPath)() {
		t.Errorf(cur + " should not be a parent path")
	}
	if !cross(HasCallerSameNamespace)() {
		t.Errorf(cur + " should be from the same namespace")
	}

	// direct parent
	cur = "gno.land/r/demo"
	testing.SetRealm(std.NewCodeRealm(cur))
	if cross(IsCallerSubPath)() {
		t.Errorf(cur + " should not be a sub path")
	}
	if !cross(IsCallerParentPath)() {
		t.Errorf(cur + " should be a parent path")
	}
	if !cross(HasCallerSameNamespace)() {
		t.Errorf(cur + " should be from the same namespace")
	}

	// fake parent (prefix)
	cur = "gno.land/r/dem"
	testing.SetRealm(std.NewCodeRealm(cur))
	if cross(IsCallerSubPath)() {
		t.Errorf(cur + " should not be a sub path")
	}
	if cross(IsCallerParentPath)() {
		t.Errorf(cur + " should not be a parent path")
	}
	if cross(HasCallerSameNamespace)() {
		t.Errorf(cur + " should not be from the same namespace")
	}

	// different namespace
	cur = "gno.land/r/foo"
	testing.SetRealm(std.NewCodeRealm(cur))
	if cross(IsCallerSubPath)() {
		t.Errorf(cur + " should not be a sub path")
	}
	if cross(IsCallerParentPath)() {
		t.Errorf(cur + " should not be a parent path")
	}
	if cross(HasCallerSameNamespace)() {
		t.Errorf(cur + " should not be from the same namespace")
	}
}
