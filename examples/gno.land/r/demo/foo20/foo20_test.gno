package foo20

import (
	pusers "gno.land/p/demo/users"
	"gno.land/r/demo/users"
	"std"
	"testing"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	admin := pusers.AddressOrName("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
	unknown := pusers.AddressOrName("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // valid but never used.

	std.TestSetOrigCaller(users.Resolve(admin))
	EnableFaucet()

	type test struct {
		name    string
		balance uint64
		fn      func() uint64
	}

	// check balances #1.
	{
		tests := []test{
			{"TotalSupply", 1_000_000 * 10000, func() uint64 { return TotalSupply() }},
			{"BalanceOf(admin)", 10000000000, func() uint64 { return BalanceOf(admin) }},
			{"BalanceOf(unknown)", 0, func() uint64 { return BalanceOf(unknown) }},
		}
		for _, tc := range tests {
			if tc.fn() != tc.balance {
				t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.balance)
			}
		}
	}

	// unknown uses the faucet.
	std.TestSetOrigCaller(users.Resolve(unknown))
	Faucet()

	// check balances #2.
	{
		tests := []test{
			{"TotalSupply", 10010000000, func() uint64 { return TotalSupply() }},
			{"BalanceOf(admin)", 10000000000, func() uint64 { return BalanceOf(admin) }},
			{"BalanceOf(unknown)", 10000000, func() uint64 { return BalanceOf(unknown) }},
		}
		for _, tc := range tests {
			if tc.fn() != tc.balance {
				t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.balance)
			}
		}
	}
}

func TestErrConditions(t *testing.T) {
	admin := std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
	empty := std.Address("")

	type test struct {
		name string
		msg  string
		fn   func()
	}

	std.TestSetOrigCaller(admin)
	{
		tests := []test{
			{"Approve(empty, 1))", "invalid address", func() { Approve(pusers.AddressOrName(empty), 1) }},
		}
		for _, tc := range tests {
			shouldPanicWithMsg(t, tc.fn, tc.msg)
		}
	}
}

func shouldPanicWithMsg(t *testing.T, f func(), expectedMsg string) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected a panic")
		}
	}()
	f()
}
