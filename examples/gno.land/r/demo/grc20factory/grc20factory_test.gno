package grc20factory

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestReadOnlyPublicMethods(t *testing.T) {

	admin := testutils.TestAddress("admin")     // g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh
	manfred := testutils.TestAddress("manfred") // g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
	unknown := testutils.TestAddress("unknown") // // valid but never used

	NewTokenWithAdmin("Foo", "FOO", 4, 10000000000, admin)
	NewTokenWithAdmin("Bar", "BAR", 4, 10000000, admin)
	mustTokenBySymbol("FOO").token.Mint(manfred, 100000000)

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
			if tc.fn() != tc.balance {
				t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.balance)
			}
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
			if tc.fn() != tc.balance {
				t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.balance)
			}
		}
	}
}
