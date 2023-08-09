package flippando

import (
	"math/rand"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestGenerateRandomNumbers(t *testing.T) {
	// Create a seed value for the random number generator to ensure deterministic testing.
	// The seed can be any arbitrary value.
	seed := 123

	// Override the random number generator with a deterministic one for testing.
	randIntn = func(n int) int {
		seed = (seed*1103515245 + 12345) & 0x7FFFFFFF
		return seed % n
	}

	// Test case 1: Positive test case
	randomValues := generateRandomNumbers(4, 1, 4)
	expectedValues := []int{1, 2, 3, 4} // Since we override the random number generator, we know the output.
	if !sliceEqual(randomValues, expectedValues) {
		t.Fatalf("Expected %v, but got %v", expectedValues, randomValues)
	}

	// Test case 3: Edge case with num exceeding the range size
	randomValues = generateRandomNumbers(10, 1, 4)
	if randomValues != nil {
		t.Fatalf("Expected nil for num exceeding range size, but got %v", randomValues)
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
