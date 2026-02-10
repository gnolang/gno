package foo20

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	var (
		admin = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
	)

	type test struct {
		name    string
		balance int64
		fn      func() int64
	}

	// check balances #1.
	{
		tests := []test{
			{"TotalSupply", 10_000_000_000, func() int64 { return TotalSupply() }},
			{"BalanceOf(admin)", 10_000_000_000, func() int64 { return BalanceOf(admin) }},
			{"BalanceOf(alice)", 0, func() int64 { return BalanceOf(alice) }},
			{"Allowance(admin, alice)", 0, func() int64 { return Allowance(admin, alice) }},
			{"BalanceOf(bob)", 0, func() int64 { return BalanceOf(bob) }},
		}
		for _, tc := range tests {
			got := tc.fn()
			uassert.Equal(t, got, tc.balance)
		}
	}

	// bob uses the faucet.
	testing.SetOriginCaller(bob)
	Faucet(cross)

	// check balances #2.
	{
		tests := []test{
			{"TotalSupply", 10_010_000_000, func() int64 { return TotalSupply() }},
			{"BalanceOf(admin)", 10_000_000_000, func() int64 { return BalanceOf(admin) }},
			{"BalanceOf(alice)", 0, func() int64 { return BalanceOf(alice) }},
			{"Allowance(admin, alice)", 0, func() int64 { return Allowance(admin, alice) }},
			{"BalanceOf(bob)", 10_000_000, func() int64 { return BalanceOf(bob) }},
		}
		for _, tc := range tests {
			got := tc.fn()
			uassert.Equal(t, got, tc.balance)
		}
	}
}

func TestErrConditions(t *testing.T) {
	var (
		admin = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
		empty = address("")
	)

	type test struct {
		name    string
		msg     string
		isCross bool
		fn      func()
	}

	privateLedger.Mint(address(admin), 10000)
	{
		tests := []test{
			{"Transfer(admin, 1)", "cannot send transfer to self", false, func() {
				// XXX: should replace with: Transfer(admin, 1)
				// but there is currently a limitation in manipulating the frame stack and simulate
				// calling this package from an outside point of view.
				adminAddr := address(admin)
				if err := privateLedger.Transfer(adminAddr, adminAddr, 1); err != nil {
					panic(err)
				}
			}},
			{"Approve(empty, 1))", "invalid address", true, func() { Approve(cross, empty, 1) }},
		}
		for _, tc := range tests {
			if tc.isCross {
				uassert.AbortsWithMessage(t, tc.msg, tc.fn)
			} else {
				uassert.PanicsWithMessage(t, tc.msg, tc.fn)
			}
		}
	}
}

//func TestNewFoo20(t *testing.T) {
//	t.Run("invalid input", func(t *testing.T) {
//		testCases := []struct {
//			msg string
//			fn  func()
//		}{
//			// Test AbortsWithMessage
//	uassert.PanicsWithMessage", func() { NewFoo20("foo", "f", 0) }},
//			{"symbol cannot be empty", func() { NewFoo20("foo", "", 1) }},
//			{"name cannot be empty", func() { NewFoo20("", "f", 1) }},
//		}
//		for _, tc := range testCases {
//			uassert.AbortsWithMessage(t, tc.msg, tc.fn)
//		}
//	})
//	t.Run("transfer", func(t *testing.T) {
//		// ... existing code ...
//	})
//}
