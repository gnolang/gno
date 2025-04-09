package params

import (
	"testing"
)

// Testing this package is limited because it only contains an `std.Set` method
// without a corresponding `std.Get` method. For comprehensive testing, refer to
// the tests located in the r/gov/dao/ directory, specifically in one of the
// propX_filetest.gno files.

func TestNewStringPropRequest(t *testing.T) {
	pr := NewSysParamStringPropRequest("foo", "bar", "baz", "qux")
	if pr.Title() == "" {
		t.Errorf("executor shouldn't be nil")
	}
}
