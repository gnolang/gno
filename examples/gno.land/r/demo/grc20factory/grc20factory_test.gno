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
	NewWithAdmin("Foo", "FOO", 4, 10_000*1_000_000, 0, admin)
	NewWithAdmin("Bar", "BAR", 4, 10_000*1_000, 0, admin)
	mustGetInstance("FOO").banker.Mint(manfred, 100_000_000)

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
			{"TotalSupply", 10_110_000_000, func() uint64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", 10_000_000_000, func() uint64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(manfred)", 100_000_000, func() uint64 { return BalanceOf("FOO", manfred) }},
			{"Allowance(admin, manfred)", 0, func() uint64 { return Allowance("FOO", admin, manfred) }},
			{"BalanceOf(unknown)", 10_000_000, func() uint64 { return BalanceOf("FOO", unknown) }},
		}
		for _, tc := range tests {
			uassert.Equal(t, tc.balance, tc.fn(), "balance does not match")
		}
	}

	// check Render
	renderTests := []struct {
		path     string
		expected string
	}{
		{"", "Token: Foo, Admin: " + admin.String() + "\nToken: Bar, Admin: " + admin.String() + "\n"},
		{"FOO", mustGetInstance("FOO").banker.RenderHome()},
		{"FOO/balance/g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", "0"},
		{"unknown", "404\n"},
	}

	for _, tc := range renderTests {
		result := Render(tc.path)
		uassert.Equal(t, tc.expected, result, "render output does not match for path %s", tc.path)
	}
}
