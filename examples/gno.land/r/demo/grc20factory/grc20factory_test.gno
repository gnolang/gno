package foo20

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	admin := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	manfred := std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	unknown := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // valid but never used.
	NewTokenWithAdmin("Foo", "FOO", 4, 10000000000, admin)
	NewTokenWithAdmin("Bar", "BAR", 4, 10000000, admin)
	mustTokenBySymbol("FOO").adminToken.Mint(manfred, 100000000)

	type test struct {
		name    string
		balance uint64
		fn      func() uint64
	}

	// check balances #1.
	{
		tests := []test{
			{"TotalSupply", 10100000000, func() uint64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", 10000000000, func() uint64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(manfred)", 100000000, func() uint64 { return BalanceOf("FOO", manfred) }},
			{"Allowance(admin, manfred)", 0, func() uint64 { return Allowance("FOO", admin, manfred) }},
			{"BalanceOf(unknown)", 0, func() uint64 { return BalanceOf("FOO", unknown) }},
		}
		for _, tc := range tests {
			uassert.Equal(t, tc.balance, tc.fn(), "balance does not match")
		}
	}
	return

	// unknown uses the faucet.
	std.TestSetOrigCaller(unknown)
	Faucet("FOO")

	// check balances #2.
	{
		tests := []test{
			{"TotalSupply", 10110000000, func() uint64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", 10000000000, func() uint64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(manfred)", 100000000, func() uint64 { return BalanceOf("FOO", manfred) }},
			{"Allowance(admin, manfred)", 0, func() uint64 { return Allowance("FOO", admin, manfred) }},
			{"BalanceOf(unknown)", 10000000, func() uint64 { return BalanceOf("FOO", unknown) }},
		}
		for _, tc := range tests {
			uassert.Equal(t, tc.balance, tc.fn(), "balance does not match")
		}
	}
}
