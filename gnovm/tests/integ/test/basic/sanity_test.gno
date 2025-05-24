// Basic testing file for testing the `gno test` command

package basic

import (
	"math"
	"testing"
)

func TestBasic(t *testing.T) {
	tc := []struct {
		name        string
		valToSquare float64
	}{
		{
			// This test should pass
			name:        "greater than one",
			valToSquare: 5.0,
		},
		{
			// This test should fail
			name:        "less than one",
			valToSquare: 0.5,
		},
		// {
		// 	// This test should loop infinitely
		// 	name:        "exactly one",
		// 	valToSquare: 1.0,
		// },
	}
	for _, testCase := range tc {
		t.Run(testCase.name, func(t *testing.T) {
			lastVal := testCase.valToSquare
			for math.Abs(testCase.valToSquare - lastVal) < 0.0000001 {
				lastVal = Square(lastVal)
			}
			// Sometimes true assertion: So we can test pass, fail and loop
			if lastVal <= testCase.valToSquare {
				t.Errorf("Got lastVal = %f, want > %f", lastVal, testCase.valToSquare)
			}
		})
	}
}
