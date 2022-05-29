package foo20

import (
	"testing"

	"gno.land/r/users"
)

/*
func TestReadOnlyPublicMethods(t *testing.T) {
	var admin = users.AddressOrName("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	var manfred = users.AddressOrName("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	var unknown = users.AddressOrName("g1u0000000000000000000000000000000000000")

	type test struct {
		name    string
		balance uint64
		fn      func() uint64
	}

	tests := []test{
		{
			"TotalSupply",
			10100000000,
			func() uint64 { return TotalSupply() },
		},
		{
			"BalanceOf(admin)",
			10000000000,
			func() uint64 { return BalanceOf(admin) },
		},
		{
			"BalanceOf(manfred)",
			100000000,
			func() uint64 { return BalanceOf(manfred) },
		},
		{
			"BalanceOf(unknown)",
			0,
			func() uint64 { return Allowance(admin, manfred) },
		},
	}

	for _, tc := range tests {
		if tc.fn() != tc.balance {
			t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.balance)
		}
	}
}
*/
