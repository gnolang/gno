package flippando

import (
	"math/rand"
	"sort"
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
)

func TestGenerateRandomNumbers(t *testing.T) {
	tests := []struct {
		num, start, end int
		valid           bool
	}{
		{5, 1, 10, true},
		{-1, 1, 10, false},
		{5, 10, 1, true},
		{11, 1, 10, false},
	}

	for _, test := range tests {
		result := generateRandomNumbers(test.num, test.start, test.end)

		if test.valid && result == nil {
			t.Errorf("Expected valid result for input (%d, %d, %d), got nil", test.num, test.start, test.end)
			continue
		}

		if !test.valid && result != nil {
			t.Errorf("Expected nil for input (%d, %d, %d), got: %v", test.num, test.start, test.end, result)
			continue
		}

		if test.valid {
			// Check if values are unique
			values := make(map[int]bool)
			for _, val := range result {
				if values[val] {
					t.Errorf("Duplicate value found: %d", val)
				}
				values[val] = true

				// Check if the value is within range
				if val < test.start || val > test.end {
					t.Errorf("Value out of range: %d not in [%d, %d]", val, test.start, test.end)
				}
			}

			// Check if we got the correct number of values
			if len(result) != test.num {
				t.Errorf("Expected %d values, got %d", test.num, len(result))
			}
		}
	}
}

// Optional: Test the randomness (this is more of a sanity check)
func TestRandomness(t *testing.T) {
	results := make(map[int]int)
	trials := 1000
	for i := 0; i < trials; i++ {
		numbers := generateRandomNumbers(1, 1, 10)
		results[numbers[0]]++
	}

	// Expect about even distribution
	for num, count := range results {
		if count < (trials/10)/2 || count > (trials/10)*2 { // This is just a basic check
			t.Errorf("Number %d appears %d times, which might indicate lack of randomness", num, count)
		}
	}
}

// Helper function to compare two slices for equality
func sliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
