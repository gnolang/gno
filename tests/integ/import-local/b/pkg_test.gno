package pkga

import "testing"

func TestPkgB(t *testing.T) {
	got := Quadruple(42)
	expected := 168
	if got != expected {
		t.Errorf("invalid answer, got %d, want %d", got, expected)
	}
}
