package grc20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

const (
	name       = "MyToken"
	symbol     = "MT"
	decimals   = 6
	mintAmount = uint64(100)
)

var (
	alice = testutils.TestAddress("alice") // g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh
	bob   = testutils.TestAddress("bob")   // g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
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

	token.Mint(alice, mintAmount)

	bal := token.BalanceOf(alice)

	if bal != mintAmount {
		t.Fatalf("Expected %d tokens for alice's balance, got %d", mintAmount, bal)
	}

	ts := token.TotalSupply()
	if ts != mintAmount {
		t.Fatalf("Expected %d in total supply, got %d", mintAmount, ts)
	}
}

func TestToken_MintInvalid(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Mint should have panicked")
		}
	}()

	// Try to mint to invalid address
	token.Mint(emptyAddress, mintAmount)
}

func TestToken_Burn(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)
	token.Mint(alice, mintAmount)

	bal := token.BalanceOf(alice)

	if bal != mintAmount {
		t.Fatalf("Expected %d tokens for alice's balance, got %d", mintAmount, bal)
	}

	// Burn whole balance
	token.Burn(alice, bal)

	bal = token.BalanceOf(alice)

	if bal != 0 {
		t.Fatalf("Expected %d tokens for alice's balance, got %d", 0, bal)
	}

	ts := token.TotalSupply()
	if ts != 0 {
		t.Fatalf("Expected %d in total supply, got %d", 0, ts)
	}
}

func TestToken_BurnInvalid(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)
	token.Mint(alice, mintAmount)

	bal := token.BalanceOf(alice)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Burn should have panicked")
		}
	}()

	// Try to burn more than in balance
	token.Burn(alice, bal+1)
}

func TestToken_Transfer(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)
	token.Mint(alice, mintAmount)

	std.TestSetOrigCaller(alice)

	bobBalance := token.BalanceOf(bob)

	if bobBalance != 0 {
		t.Fatalf("Expected %d tokens for bob's balance, got %d", 0, bobBalance)
	}

	token.Transfer(bob, mintAmount/2)

	bobBalance = token.BalanceOf(bob)
	aliceBalance := token.BalanceOf(alice)

	if bobBalance != mintAmount/2 {
		t.Fatalf("Expected %d tokens for alice's balance, got %d", mintAmount/2, bobBalance)
	}

	if aliceBalance != mintAmount/2 {
		t.Fatalf("Expected %d tokens for alice's balance, got %d", mintAmount/2, aliceBalance)
	}
}

func TestToken_TransferInvalid(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)
	token.Mint(alice, mintAmount)

	std.TestSetOrigCaller(alice)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Transfer should have panicked")
		}
	}()

	// Try to transfer more than in balance
	token.Transfer(bob, mintAmount+1)
}

func TestToken_Approve(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)
	token.Mint(alice, mintAmount)

	std.TestSetOrigCaller(alice)

	token.Approve(bob, mintAmount)
	allowance := token.Allowance(alice, bob)

	if allowance != mintAmount {
		t.Fatalf("Expected %d tokens in allowance, got %d", mintAmount, allowance)
	}
}

func TestToken_TransferFrom(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)
	token.Mint(alice, mintAmount)

	std.TestSetOrigCaller(alice)

	token.Approve(bob, mintAmount)

	std.TestSetOrigCaller(bob)

	token.TransferFrom(alice, bob, mintAmount)

	allowance := token.Allowance(alice, bob)
	if allowance != 0 {
		t.Fatalf("Expected %d tokens in allowance after spending, got %d", mintAmount, allowance)
	}
}

func TestToken_TransferFromInvalid(t *testing.T) {
	token := NewGRC20Token(name, symbol, decimals)
	token.Mint(alice, mintAmount)

	std.TestSetOrigCaller(alice)

	token.Approve(bob, mintAmount)

	std.TestSetOrigCaller(bob)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("TransferFrom should have panicked")
		}
	}()

	// Try to transfer out more than in allowance
	token.TransferFrom(alice, bob, mintAmount+1)
}
