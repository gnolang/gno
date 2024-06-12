package nestedpkg

import (
	"std"
	"testing"

	"gno.land/r/demo/tests"
)

func TestPackage(t *testing.T) {
	// direct child
	cur := "gno.land/r/demo/tests/foo"
	std.TestSetRealm(std.NewCodeRealm(cur))
	if !tests.IsCallerSubPath() {
		t.Errorf(cur + " should be a sub path")
	}
	if tests.IsCallerParentPath() {
		t.Errorf(cur + " should not be a parent path")
	}
	if !tests.HasCallerSameNamespace() {
		t.Errorf(cur + " should be from the same namespace")
	}

	// grand-grand-child
	cur = "gno.land/r/demo/tests/foo/bar/baz"
	std.TestSetRealm(std.NewCodeRealm(cur))
	if !tests.IsCallerSubPath() {
		t.Errorf(cur + " should be a sub path")
	}
	if tests.IsCallerParentPath() {
		t.Errorf(cur + " should not be a parent path")
	}
	if !tests.HasCallerSameNamespace() {
		t.Errorf(cur + " should be from the same namespace")
	}

	// direct parent
	cur = "gno.land/r/demo"
	std.TestSetRealm(std.NewCodeRealm(cur))
	if tests.IsCallerSubPath() {
		t.Errorf(cur + " should not be a sub path")
	}
	if !tests.IsCallerParentPath() {
		t.Errorf(cur + " should be a parent path")
	}
	if !tests.HasCallerSameNamespace() {
		t.Errorf(cur + " should be from the same namespace")
	}

	// fake parent (prefix)
	cur = "gno.land/r/dem"
	std.TestSetRealm(std.NewCodeRealm(cur))
	if tests.IsCallerSubPath() {
		t.Errorf(cur + " should not be a sub path")
	}
	if tests.IsCallerParentPath() {
		t.Errorf(cur + " should not be a parent path")
	}
	if tests.HasCallerSameNamespace() {
		t.Errorf(cur + " should not be from the same namespace")
	}

	// different namespace
	cur = "gno.land/r/foo"
	std.TestSetRealm(std.NewCodeRealm(cur))
	if tests.IsCallerSubPath() {
		t.Errorf(cur + " should not be a sub path")
	}
	if tests.IsCallerParentPath() {
		t.Errorf(cur + " should not be a parent path")
	}
	if tests.HasCallerSameNamespace() {
		t.Errorf(cur + " should not be from the same namespace")
	}
}
