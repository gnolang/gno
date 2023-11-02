package std

import (
	"testing"
)

func TestValid(t *testing.T) {
	type test struct {
		inputAddress Address
		expected     bool
	}

	testCases := []test{
		{inputAddress: "g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv", expected: true},
		{inputAddress: "g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa", expected: true},
		{inputAddress: "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq", expected: true},
		{inputAddress: "g14da4n9hcynyzz83q607uu8keuh9hwlv42ra6fa", expected: true},
		{inputAddress: "", expected: false},
		{inputAddress: "000000000000", expected: false},
		{inputAddress: "0000000000000000000000000000000000000000000000000000000000000000000000", expected: false},
	}

	for _, tc := range testCases {
		result := tc.inputAddress.IsValid()

		if result != tc.expected {
			t.Fatalf("Expected: %t, got: %t", tc.expected, result)
		}
	}
}
