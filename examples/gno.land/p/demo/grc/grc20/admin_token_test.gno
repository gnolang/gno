package grc20

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestAdminTokenImpl(t *testing.T) {
	dummy := NewAdminToken("Dummy", "DUMMY", 4)
	uassert.True(t, dummy != nil, "should not be nil")
}

func TestAllowance(t *testing.T) {
	owner := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	spender := std.Address("g1us8428u2a5satrlxzagsqa5m6vmuze027sxc8x")
	dest := std.Address("g1us8428m6a5satrlxzagsqa5m6vmuze02tyelwj")

	token := NewAdminToken("Dummy", "DUMMY", 6)
	uassert.NoError(t, token.Mint(owner, 100000000))
	uassert.NoError(t, token.Approve(owner, spender, 5000000))
	uassert.Error(t, token.TransferFrom(spender, owner, dest, 10000000), "should not be able to transfer more than approved")

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
		uassert.NoError(t, token.TransferFrom(spender, owner, dest, tt.spend))

		a, _ := token.Allowance(owner, spender)
		uassert.Equal(t, tt.exp, a)

		b, _ := token.BalanceOf(dest)
		expB := b0 + tt.spend
		uassert.Equal(t, expB, b)
	}

	err := token.TransferFrom(spender, owner, dest, 1)
	uassert.Error(t, err, "no allowance")

	key := allowanceKey(owner, spender)
	uassert.False(t, token.allowances.Has(key), "allowance should be removed")
}
