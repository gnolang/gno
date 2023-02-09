package main

import (
	"fmt"
	"testing"
)

func TestRender(t *testing.T) {
	got := Render("subpath/subsubpath")
	expected := "hello hello:subpath/subsubpath"
	if got != expected {
		t.Errorf("invalid output: %q instead of %q", got, expected)
	}
}

func TestHello(t *testing.T) {
	got := Hello()
	expected := "world"

	if got != expected {
		t.Errorf("invalid output: %q instead of %q", got, expected)
	}
}

func TestSumPow2(t *testing.T) {
	tests := []struct {
		numberA  int
		numberB  int
		expected int
	}{
		{2, 3, 25},
		{2, 4, 36},
		{2, 5, 49},
		{2, 30, 1024},
	}

	for _, testCase := range tests {
		name := fmt.Sprintf("%d-%d", testCase.numberA, testCase.numberB)
		t.Run(name, func(t *testing.T) {
			got := SumPow2(testCase.numberA, testCase.numberB)
			expected := testCase.expected
			if got != expected {
				t.Errorf("failed: got %d, expected %d", got, expected)
			}
		})
	}
}
