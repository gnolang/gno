package grc20

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestBankerImpl(t *testing.T) {
	dummy := NewBanker("Dummy", "DUMMY", 4)
	urequire.False(t, dummy == nil, "dummy should not be nil")
}

func TestAllowance(t *testing.T) {
	var (
		owner   = testutils.TestAddress("owner")
		spender = testutils.TestAddress("spender")
		dest    = testutils.TestAddress("dest")
	)

	b := NewBanker("Dummy", "DUMMY", 6)
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
