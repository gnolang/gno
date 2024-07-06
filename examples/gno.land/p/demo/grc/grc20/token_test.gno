package grc20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestUserTokenImpl(t *testing.T) {
	bank := NewBanker("Dummy", "DUMMY", 4)
	tok := bank.Token()
	_ = tok
}

func TestUserApprove(t *testing.T) {
	owner := testutils.TestAddress("owner")
	spender := testutils.TestAddress("spender")
	dest := testutils.TestAddress("dest")

	bank := NewBanker("Dummy", "DUMMY", 6)
	tok := bank.Token()

	// Set owner as the original caller
	std.TestSetOrigCaller(owner)
	// Mint 100000000 tokens for owner
	urequire.NoError(t, bank.Mint(owner, 100000000))

	// Approve spender to spend 5000000 tokens
	urequire.NoError(t, tok.Approve(spender, 5000000))

	// Set spender as the original caller
	std.TestSetOrigCaller(spender)
	// Try to transfer 10000000 tokens from owner to dest, should fail because it exceeds allowance
	urequire.Error(t,
		tok.TransferFrom(owner, dest, 10000000),
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
		b0 := tok.BalanceOf(dest)
		// Perform transfer from owner to dest
		urequire.NoError(t, tok.TransferFrom(owner, dest, tt.spend))
		a := tok.Allowance(owner, spender)
		// Check if allowance equals expected value
		urequire.True(t, a == tt.exp, ufmt.Sprintf("allowance exp: %d,got %d", tt.exp, a))

		// Get dest current balance
		b := tok.BalanceOf(dest)
		// Calculate expected balance ,should be initial balance plus transfer amount
		expB := b0 + tt.spend
		// Check if balance equals expected value
		urequire.True(t, b == expB, ufmt.Sprintf("balance exp: %d,got %d", expB, b))
	}

	// Try to transfer one token from owner to dest ,should fail because no allowance left
	urequire.Error(t, tok.TransferFrom(owner, dest, 1), ErrInsufficientAllowance.Error(), "no allowance")
}
