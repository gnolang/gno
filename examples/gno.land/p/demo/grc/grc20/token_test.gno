package grc20

import (
	"std"
	"testing"
)

func TestUserTokenImpl(t *testing.T) {
	dummyAdmin := NewAdminToken("Dummy", "DUMMY", 4)
	dummyUser := dummyAdmin.GRC20()
	_ = dummyUser
}

func TestUserApprove(t *testing.T) {
	owner := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	spender := std.Address("g1us8428u2a5satrlxzagsqa5m6vmuze027sxc8x")
	dest := std.Address("g1us8428m6a5satrlxzagsqa5m6vmuze02tyelwj")

	dummyAdmin := NewAdminToken("Dummy", "DUMMY", 6)

	// Set owner as the original caller
	std.TestSetOrigCaller(owner)
	// Mint 100000000 tokens for owner
	assertE(t, dummyAdmin.Mint(owner, 100000000))

	dummyUser := dummyAdmin.GRC20()
	// Approve spender to spend 5000000 tokens
	assertE(t, dummyUser.Approve(spender, 5000000))

	// Set spender as the original caller
	std.TestSetOrigCaller(spender)
	// Try to transfer 10000000 tokens from owner to dest, should fail because it exceeds allowance
	err := dummyUser.TransferFrom(owner, dest, 10000000)
	assert(t, err != nil, "should not be able to transfer more than approved")

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
		b0, _ := dummyUser.BalanceOf(dest)
		// Perform transfer from owner to dest
		assertE(t, dummyUser.TransferFrom(owner, dest, tt.spend))
		a, _ := dummyUser.Allowance(owner, spender)
		// Check if allowance equals expected value
		assert(t, a == tt.exp, "allowance exp: %d,got %d", tt.exp, a)

		// Get dest current balance
		b, _ := dummyUser.BalanceOf(dest)
		// Calculate expected balance ,should be initial balance plus transfer amount
		expB := b0 + tt.spend
		// Check if balance equals expected value
		assert(t, b == expB, "balance exp: %d,got %d", expB, b)
	}

	// Try to transfer one token from owner to dest ,should fail because no allowance left
	err = dummyUser.TransferFrom(owner, dest, 1)
	assert(t, err != nil, "no allowance")
}
