package example

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"positive numbers", 2, 3, 5},
		{"with zero", 0, 5, 5},
		{"negative numbers", -1, 5, 0}, // tests the if condition
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	// Only testing one case - partial coverage
	result := Multiply(2, 3)
	if result != 6 {
		t.Errorf("Multiply(2, 3) = %d, want 6", result)
	}
}