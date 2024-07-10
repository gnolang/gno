package foo20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	std.TestSetOrigPkgAddr("gno.land/r/demo/grc20factory")
	admin := testutils.TestAddress("admin")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")

	std.TestSetOrigCaller(admin)
	std.TestSetRealm(std.NewUserRealm(admin))
	NewWithAdmin("Foo", "FOO", 4, 10_000*1_000_000, 10_000, admin)
	NewWithAdmin("Bar", "BAR", 4, 10_000*1_000, 10_000, admin)
	mustGetInstance("FOO").bank.Mint(bob, 100_000_000)

	type test struct {
		name    string
		balance uint64
		fn      func() uint64
	}

	// check balances #1.
	{
		tests := []test{
			{"TotalSupply", 10_000_000_000, func() uint64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", 10_000_000_000, func() uint64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(bob)", 0, func() uint64 { return BalanceOf("FOO", bob) }},
			{"Allowance(admin, bob)", 0, func() uint64 { return Allowance("FOO", admin, bob) }},
			{"BalanceOf(carl)", 0, func() uint64 { return BalanceOf("FOO", carl) }},
		}
		for _, tc := range tests {
			if tc.fn() != tc.balance {
				t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.balance)
			}
		}
	}

	// carl uses the faucet.
	// std.TestSetOrigCaller(carl)
	// std.TestSetRealm(std.NewUserRealm(carl))
	Faucet("FOO")
	return

	// check balances #2.
	{
		tests := []test{
			{"TotalSupply", 10_000_010_000, func() uint64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", 10_000_000_000, func() uint64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(bob)", 0, func() uint64 { return BalanceOf("FOO", bob) }},
			{"Allowance(admin, bob)", 0, func() uint64 { return Allowance("FOO", admin, bob) }},
			{"BalanceOf(carl)", 10_000, func() uint64 { return BalanceOf("FOO", carl) }},
		}
		for _, tc := range tests {
			if tc.fn() != tc.balance {
				t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.balance)
			}
		}
	}
}
