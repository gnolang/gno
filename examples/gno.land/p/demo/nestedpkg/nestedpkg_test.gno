package nestedpkg

import (
	"std"
	"testing"

	"gno.land/p/demo/nestedpkg"
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

	// grand-grand-child
	cur = "gno.land/r/demo/tests/foo/bar/baz"
	std.TestSetRealm(std.NewCodeRealm(cur))
	if !tests.IsCallerSubPath() {
		t.Errorf(cur + " should be a sub path")
	}
	if tests.IsCallerParentPath() {
		t.Errorf(cur + " should not be a parent path")
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

	// XXX: grand-grand-parent
}
