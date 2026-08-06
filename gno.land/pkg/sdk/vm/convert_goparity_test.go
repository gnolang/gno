package vm

import (
	"math"
	"testing"
)

// TestConvertFloatMatchesGo locks the float call-argument conversion to Go's
// semantics. A "-0.0"/"-0" argument folds to +0, matching Go, where the source
// literal -0.0 is a constant that folds to +0. NaN and Inf are accepted, not
// rejected: Go takes them as float arguments and the VM produces them too. This
// guards against a future change dropping the -0 fold or rejecting NaN/Inf.
func TestConvertFloatMatchesGo(t *testing.T) {
	t.Parallel()

	// Signbit cases: assert the exact zero produced, so -0 is caught.
	zeros := []struct {
		in   string
		want float64 // Go value for the same source
	}{
		{"0", 0},
		{"-0", 0},   // Go folds the -0.0 literal to +0
		{"-0.0", 0}, // same
	}
	for _, prec := range []int{32, 64} {
		for _, tc := range zeros {
			got := convertFloat(tc.in, prec)
			if math.Float64bits(got) != math.Float64bits(tc.want) {
				t.Errorf("convertFloat(%q, %d) bits = %#x, want %#x (+0)",
					tc.in, prec, math.Float64bits(got), math.Float64bits(tc.want))
			}
			if math.Signbit(got) {
				t.Errorf("convertFloat(%q, %d) has sign bit set, want +0", tc.in, prec)
			}
		}
	}

	// NaN and Inf must be accepted, never rejected.
	if got := convertFloat("NaN", 64); !math.IsNaN(got) {
		t.Errorf("convertFloat(\"NaN\", 64) = %v, want NaN", got)
	}
	if got := convertFloat("Inf", 64); !math.IsInf(got, 1) {
		t.Errorf("convertFloat(\"Inf\", 64) = %v, want +Inf", got)
	}
	if got := convertFloat("-Inf", 64); !math.IsInf(got, -1) {
		t.Errorf("convertFloat(\"-Inf\", 64) = %v, want -Inf", got)
	}
}
