package params

import (
	"testing"
)

// Testing this package is limited because it only contains an `std.Set` method
// without a corresponding `std.Get` method. For comprehensive testing, refer to
// the tests located in the r/gov/dao/ directory, specifically in one of the
// propX_filetest.gno files.

func TestNewStringPropRequest(cur realm, t *testing.T) {
	pr := NewSysParamStringPropRequest(cur, "foo", "bar", "baz", "qux")
	if pr.Title() == "" {
		t.Errorf("executor shouldn't be nil")
	}
}

func TestNewSetHaltRequest(cur realm, t *testing.T) {
	pr := NewSetHaltRequest(cur, 100_000, "chain/gnoland1.1")
	if pr.Title() == "" {
		t.Errorf("proposal title shouldn't be empty")
	}
}

func TestNewSetHaltRequestNoVersion(cur realm, t *testing.T) {
	pr := NewSetHaltRequest(cur, 100_000, "")
	if pr.Title() == "" {
		t.Errorf("proposal title shouldn't be empty")
	}
}

func TestNewSetHaltRequestCancel(cur realm, t *testing.T) {
	pr := NewSetHaltRequest(cur, 0, "")
	if pr.Title() == "" {
		t.Errorf("proposal title shouldn't be empty")
	}
}
