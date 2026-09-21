package gnolang

import "testing"

// TestConstantValuesDeepFill verifies that TypedValue.DeepFill leaves the leaf
// constant value types (StringValue, BigintValue, BigdecValue) unchanged, since
// they hold no references to resolve.
func TestConstantValuesDeepFill(t *testing.T) {
	for _, v := range []Value{StringValue("test"), BigintValue{}, BigdecValue{}} {
		tv := TypedValue{V: v}
		tv.DeepFill(nil)
		if tv.V != v {
			t.Errorf("%T was modified by DeepFill: got %v, want %v", v, tv.V, v)
		}
	}
}
