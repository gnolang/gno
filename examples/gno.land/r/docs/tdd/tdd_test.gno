package tdd

import "testing"

func TestAdd(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; want %d",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	result := Subtract(5, 3)
	if result != 2 {
		t.Errorf("Subtract(5, 3) = %d; want 2", result)
	}
}

func TestMultiply(t *testing.T) {
	result := Multiply(4, 5)
	if result != 20 {
		t.Errorf("Multiply(4, 5) = %d; want 20", result)
	}
}

func TestDivide(t *testing.T) {
	result := Divide(10, 2)
	if result != 5 {
		t.Errorf("Divide(10, 2) = %d; want 5", result)
	}
}

func TestDivideByZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Divide(10, 0) should panic")
		}
	}()
	Divide(10, 0)
}

func TestGetLastResult(t *testing.T) {
	Add(7, 8)
	if GetLastResult() != 15 {
		t.Errorf("GetLastResult() = %d; want 15", GetLastResult())
	}

	Multiply(3, 4)
	if GetLastResult() != 12 {
		t.Errorf("GetLastResult() = %d; want 12", GetLastResult())
	}
}
