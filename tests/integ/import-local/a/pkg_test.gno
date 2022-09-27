package pkga

import "testing"

func TestPkgA(t *testing.T) {
	got := Double(42)
	expected := 84
	if got != expected {
		t.Errorf("invalid answer, got %d, want %d", got, expected)
	}
}
