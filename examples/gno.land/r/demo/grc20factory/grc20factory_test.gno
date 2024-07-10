package foo20

import (
	"std"
	"testing"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	admin := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	manfred := std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	unknown := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // valid but never used.
	NewWithAdmin("Foo", "FOO", 4, 10_000*1_000_000, 0, admin)
	NewWithAdmin("Bar", "BAR", 4, 10_000*1_000, 0, admin)
	mustGetInstance("FOO").bank.Mint(manfred, 100_000_000)

	type test struct {
		name    string
		balance uint64
		fn      func() uint64
	}

	// check balances #1.
	{
		tests := []test{
			{"TotalSupply", 10_100_000_000, func() uint64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", 10_000_000_000, func() uint64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(manfred)", 100_000_000, func() uint64 { return BalanceOf("FOO", manfred) }},
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
			{"TotalSupply", 10_110_000_000, func() uint64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", 10_000_000_000, func() uint64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(manfred)", 100_000_000, func() uint64 { return BalanceOf("FOO", manfred) }},
			{"Allowance(admin, manfred)", 0, func() uint64 { return Allowance("FOO", admin, manfred) }},
			{"BalanceOf(unknown)", 10_000_000, func() uint64 { return BalanceOf("FOO", unknown) }},
		}
		for _, tc := range tests {
			if tc.fn() != tc.balance {
				t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.balance)
			}
		}
	}
}
