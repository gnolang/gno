package grc20

import (
	"std"
	"testing"
)

func TestAdminTokenImpl(t *testing.T) {
	dummy := NewAdminToken("Dummy", "DUMMY", 4)
	if dummy == nil {
		t.Errorf("should not be nil")
	}
}

func TestAllowance(t *testing.T) {
	owner := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	spender := std.Address("g1us8428u2a5satrlxzagsqa5m6vmuze027sxc8x")
	dest := std.Address("g1us8428m6a5satrlxzagsqa5m6vmuze02tyelwj")

	token := NewAdminToken("Dummy", "DUMMY", 6)
	assertE(t, token.Mint(owner, 100000000))
	assertE(t, token.Approve(owner, spender, 5000000))

	err := token.TransferFrom(spender, owner, dest, 10000000)
	assert(t, err != nil, "should not be able to transfer more than approved")

	tests := []struct {
		spend uint64
		exp   uint64
	}{
		{3, 4999997},
		{999997, 4000000},
		{4000000, 0},
	}

	for _, tt := range tests {
		b0, _ := token.BalanceOf(dest)
		assertE(t, token.TransferFrom(spender, owner, dest, tt.spend))
		a, _ := token.Allowance(owner, spender)
		assert(t, a == tt.exp, "allowance exp: %d, got %d", tt.exp, a)

		b, _ := token.BalanceOf(dest)
		expB := b0+tt.spend
		assert(t, b == expB, "balance exp: %d, got %d", expB, b)
	}

	err = token.TransferFrom(spender, owner, dest, 1)
	assert(t, err != nil, "no allowance")

	key := allowanceKey(owner, spender)
	assert(t, !token.allowances.Has(key), "allowance should be removed")
}

func assertE(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func assert(t *testing.T, cond bool, format string, args ...interface{}) {
	if !cond {
		t.Fatalf(format, args...)
	}
}