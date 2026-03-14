package gnolang

import (
	"testing"
)

// TestConstantValuesDeepFill verifies that TypedValue.DeepFill
// correctly handles constant value types (StringValue, BigintValue, BigdecValue)
// by not calling their DeepFill methods (which panic).
func TestConstantValuesDeepFill(t *testing.T) {
	// Test that TypedValue.DeepFill skips constant values
	t.Run("StringValue through TypedValue", func(t *testing.T) {
		sv := StringValue("test")
		tv := TypedValue{T: StringType, V: sv}
		// This should NOT panic because TypedValue.DeepFill skips constant values
		tv.DeepFill(nil)
		if result, ok := tv.V.(StringValue); !ok || result != sv {
			t.Errorf("StringValue was modified: expected %v, got %v", sv, result)
		}
	})

	t.Run("BigintValue through TypedValue", func(t *testing.T) {
		biv := BigintValue{}
		tv := TypedValue{T: UntypedBigintType, V: biv}
		// This should NOT panic because TypedValue.DeepFill skips constant values
		tv.DeepFill(nil)
		if _, ok := tv.V.(BigintValue); !ok {
			t.Errorf("BigintValue type was changed")
		}
	})

	t.Run("BigdecValue through TypedValue", func(t *testing.T) {
		bdv := BigdecValue{}
		tv := TypedValue{T: UntypedBigdecType, V: bdv}
		// This should NOT panic because TypedValue.DeepFill skips constant values
		tv.DeepFill(nil)
		if _, ok := tv.V.(BigdecValue); !ok {
			t.Errorf("BigdecValue type was changed")
		}
	})
}
