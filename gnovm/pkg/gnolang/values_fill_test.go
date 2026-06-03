package gnolang

import (
	"testing"
)

// TestConstantValuesDeepFill verifies that TypedValue.DeepFill leaves the leaf
// constant value types (StringValue, BigintValue, BigdecValue) unchanged, since
// they hold no references to resolve.
func TestConstantValuesDeepFill(t *testing.T) {
	t.Run("StringValue through TypedValue", func(t *testing.T) {
		sv := StringValue("test")
		tv := TypedValue{T: StringType, V: sv}
		tv.DeepFill(nil)
		if result, ok := tv.V.(StringValue); !ok || result != sv {
			t.Errorf("StringValue was modified: expected %v, got %v", sv, result)
		}
	})

	t.Run("BigintValue through TypedValue", func(t *testing.T) {
		biv := BigintValue{}
		tv := TypedValue{T: UntypedBigintType, V: biv}
		tv.DeepFill(nil)
		if _, ok := tv.V.(BigintValue); !ok {
			t.Errorf("BigintValue type was changed")
		}
	})

	t.Run("BigdecValue through TypedValue", func(t *testing.T) {
		bdv := BigdecValue{}
		tv := TypedValue{T: UntypedBigdecType, V: bdv}
		tv.DeepFill(nil)
		if _, ok := tv.V.(BigdecValue); !ok {
			t.Errorf("BigdecValue type was changed")
		}
	})
}
