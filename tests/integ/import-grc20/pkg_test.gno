package importgrc20

import "testing"

func TestPackage(t *testing.T) {
	got := BalanceOf("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	expected := 10000000000
	if got != expected {
		t.Errorf("invalid balance, got %d, want %d", got, expected)
	}
}
