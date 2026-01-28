package testcoverage

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d, want 5", result)
	}
}

func TestSubtract(t *testing.T) {
	result := Subtract(5, 3)
	if result != 2 {
		t.Errorf("Subtract(5, 3) = %d, want 2", result)
	}
}

func TestMultiply(t *testing.T) {
	result := Multiply(3, 4)
	if result != 12 {
		t.Errorf("Multiply(3, 4) = %d, want 12", result)
	}
}

func TestIsEven(t *testing.T) {
	if !IsEven(4) {
		t.Error("IsEven(4) should be true")
	}
	// Note: not testing odd case for coverage testing
}