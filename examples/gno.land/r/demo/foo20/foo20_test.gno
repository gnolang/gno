package foo20

import (
	"std"
	"testing"

	pusers "gno.land/p/demo/users"
	"gno.land/r/demo/users"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	admin := pusers.AddressOrName("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	manfred := pusers.AddressOrName("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	unknown := pusers.AddressOrName("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // valid but never used.

	type test struct {
		name    string
		balance uint64
		fn      func() uint64
	}

	// check balances #1.
	{
		tests := []test{
			{"TotalSupply", 10100000000, func() uint64 { return TotalSupply() }},
			{"BalanceOf(admin)", 10000000000, func() uint64 { return BalanceOf(admin) }},
			{"BalanceOf(manfred)", 100000000, func() uint64 { return BalanceOf(manfred) }},
			{"Allowance(admin, manfred)", 0, func() uint64 { return Allowance(admin, manfred) }},
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
			{"TotalSupply", 10110000000, func() uint64 { return TotalSupply() }},
			{"BalanceOf(admin)", 10000000000, func() uint64 { return BalanceOf(admin) }},
			{"BalanceOf(manfred)", 100000000, func() uint64 { return BalanceOf(manfred) }},
			{"Allowance(admin, manfred)", 0, func() uint64 { return Allowance(admin, manfred) }},
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
	admin := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	empty := std.Address("")

	type test struct {
		name string
		msg  string
		fn   func()
	}

	std.TestSetOrigCaller(admin)
	{
		tests := []test{
			{"Transfer(admin, 1)", "cannot send transfer to self", func() { Transfer(pusers.AddressOrName(admin), 1) }},
			{"Approve(empty, 1))", "invalid address", func() { Approve(pusers.AddressOrName(empty), 1) }},
		}
		for _, tc := range tests {
			shouldPanicWithMsg(t, tc.fn, tc.msg)
		}
	}
}

func shouldPanicWithMsg(t *testing.T, f func(), msg string) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		} else {
			errMsg := error(r).Error()
			if errMsg != msg {
				t.Errorf("excepted panic(%v), got(%v)", msg, errMsg)
			}
		}
	}()
	f()
}
