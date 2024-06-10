package grc20

import (
	"gno.land/p/demo/testutils"
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

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
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

func TestToken_Mint(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)
	mintAmount := uint64(100)

	token.Mint(alice, mintAmount)

	bal := token.BalanceOf(alice)

	if bal != mintAmount {
		t.Fatalf("Expected %d tokens for alice's balance, got %d", mintAmount, bal)
	}
}
