package grc20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestBankImpl(t *testing.T) {
	dummy := NewBank("Dummy", "DUMMY", 4)
	urequire.False(t, dummy == nil, "dummy should not be nil")
}

func TestAllowance(t *testing.T) {
	var (
		owner   = testutils.TestAddress("owner")
		spender = testutils.TestAddress("spender")
		dest    = testutils.TestAddress("dest")
	)

	b := NewBank("Dummy", "DUMMY", 6)
	urequire.NoError(t, b.Mint(owner, 100000000))
	urequire.NoError(t, b.Approve(owner, spender, 5000000))
	urequire.Error(t, b.TransferFrom(spender, owner, dest, 10000000), ErrInsufficientAllowance.Error(), "should not be able to transfer more than approved")

	tests := []struct {
		spend uint64
		exp   uint64
	}{
		{3, 4999997},
		{999997, 4000000},
		{4000000, 0},
	}

	for _, tt := range tests {
		b0 := b.BalanceOf(dest)
		urequire.NoError(t, b.TransferFrom(spender, owner, dest, tt.spend))
		a := b.Allowance(owner, spender)
		urequire.Equal(t, a, tt.exp, ufmt.Sprintf("allowance exp: %d, got %d", tt.exp, a))
		b := b.BalanceOf(dest)
		expB := b0 + tt.spend
		urequire.Equal(t, b, expB, ufmt.Sprintf("balance exp: %d, got %d", expB, b))
	}

	urequire.Error(t, b.TransferFrom(spender, owner, dest, 1), "no allowance")
	key := allowanceKey(owner, spender)
	urequire.False(t, b.allowances.Has(key), "allowance should be removed")
	urequire.Equal(t, b.Allowance(owner, spender), uint64(0), "allowance should be 0")
}

func TestUserTokenImpl(t *testing.T) {
	bank := NewBank("Dummy", "DUMMY", 4)
	banker := PrevRealmBanker(bank)
	var _ GRC20 = banker
}

func TestUserApprove(t *testing.T) {
	owner := testutils.TestAddress("owner")
	spender := testutils.TestAddress("spender")
	dest := testutils.TestAddress("dest")

	bank := NewBank("Dummy", "DUMMY", 6)
	banker := PrevRealmBanker(bank)

	// Set owner as the original caller
	std.TestSetOrigCaller(owner)
	// Mint 100000000 tokens for owner
	urequire.NoError(t, bank.Mint(owner, 100000000))

	// Approve spender to spend 5000000 tokens
	urequire.NoError(t, banker.Approve(spender, 5000000))

	// Set spender as the original caller
	std.TestSetOrigCaller(spender)
	// Try to transfer 10000000 tokens from owner to dest, should fail because it exceeds allowance
	urequire.Error(t,
		banker.TransferFrom(owner, dest, 10000000),
		ErrInsufficientAllowance.Error(),
		"should not be able to transfer more than approved",
	)

	// Define a set of test data with spend amount and expected remaining allowance
	tests := []struct {
		spend uint64 // Spend amount
		exp   uint64 // Remaining allowance
	}{
		{3, 4999997},
		{999997, 4000000},
		{4000000, 0},
	}

	// perform transfer operation,and check if allowance and balance are correct
	for _, tt := range tests {
		b0 := banker.BalanceOf(dest)
		// Perform transfer from owner to dest
		urequire.NoError(t, banker.TransferFrom(owner, dest, tt.spend))
		a := banker.Allowance(owner, spender)
		// Check if allowance equals expected value
		urequire.True(t, a == tt.exp, ufmt.Sprintf("allowance exp: %d,got %d", tt.exp, a))

		// Get dest current balance
		b := banker.BalanceOf(dest)
		// Calculate expected balance ,should be initial balance plus transfer amount
		expB := b0 + tt.spend
		// Check if balance equals expected value
		urequire.True(t, b == expB, ufmt.Sprintf("balance exp: %d,got %d", expB, b))
	}

	// Try to transfer one token from owner to dest ,should fail because no allowance left
	urequire.Error(t, banker.TransferFrom(owner, dest, 1), ErrInsufficientAllowance.Error(), "no allowance")
}
