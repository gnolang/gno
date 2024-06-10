package grc20

import (
	"testing"
)

// Test Transfer
// Test TransferFrom
// Test Approve
// Test Allowance

const (
	name     = "MyToken"
	symbol   = "MT"
	decimals = 6
)

func TestNewGRC20Token(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)

	if token.Decimals() != decimals {
		t.Fatalf("Expected %d decimals, got %d", decimals, token.Decimals())
	}

	if token.Name() != name {
		t.Fatalf("Expected %s for name, got %s", name, token.Name())
	}

	if token.Symbol() != symbol {
		t.Fatalf("Expected %s for symbol, got %s", symbol, token.Symbol())
	}
}
