package flippando

import (
	"std"
	"testing"

	"gno.land/p/demo/rand"
	"gno.land/p/demo/testutils"
)

func TestGenerateRandomNumbers(t *testing.T) {
	// Create an instance of your custom deterministic random number generator.
	randInstance := rand.New() // Use your custom Instance type here

	// Test case 1: Positive test case
	randomValues := generateRandomNumbers(4, 1, 4)
	expectedValues := []int{2, 3, 1, 4}
	if !sliceEqual(randomValues, expectedValues) {
		t.Fatalf("Expected %v, but got %v", expectedValues, randomValues)
	}

	// Test case 2: Edge case with num exceeding the range size
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
