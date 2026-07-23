package tdd

import "testing"

func TestAdd(cur realm, t *testing.T) {
	tests := []struct {
		name     string
		a, b     int64
		expected int64
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -2, -3, -5},
		{"mixed signs", -2, 3, 1},
		{"zero", 0, 5, 5},
	}

	for _, tt := range tests {
		result := Add(cross(cur), tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("%s: Add(%d, %d) = %d; want %d",
				tt.name, tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestSubtract(cur realm, t *testing.T) {
	result := Subtract(cross(cur), 5, 3)
	if result != 2 {
		t.Errorf("Subtract(5, 3) = %d; want 2", result)
	}
}

func TestMultiply(cur realm, t *testing.T) {
	result := Multiply(cross(cur), 4, 5)
	if result != 20 {
		t.Errorf("Multiply(4, 5) = %d; want 20", result)
	}
}

func TestDivide(cur realm, t *testing.T) {
	result := Divide(cross(cur), 10, 2)
	if result != 5 {
		t.Errorf("Divide(10, 2) = %d; want 5", result)
	}
}

func TestGetLastResult(cur realm, t *testing.T) {
	Add(cross(cur), 7, 8)
	if GetLastResult() != 15 {
		t.Errorf("GetLastResult() = %d; want 15", GetLastResult())
	}

	Multiply(cross(cur), 3, 4)
	if GetLastResult() != 12 {
		t.Errorf("GetLastResult() = %d; want 12", GetLastResult())
	}
}
