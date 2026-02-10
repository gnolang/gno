package foo20

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	admin := testutils.TestAddress("admin")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")

	type test struct {
		name    string
		balance int64
		fn      func() int64
	}

	checkBalances := func(step string, totSup, balAdm, balBob, allowAdmBob, balCarl int64) {
		tests := []test{
			{"TotalSupply", totSup, func() int64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", balAdm, func() int64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(bob)", balBob, func() int64 { return BalanceOf("FOO", bob) }},
			{"Allowance(admin, bob)", allowAdmBob, func() int64 { return Allowance("FOO", admin, bob) }},
			{"BalanceOf(carl)", balCarl, func() int64 { return BalanceOf("FOO", carl) }},
		}
		for _, tc := range tests {
			reason := ufmt.Sprintf("%s.%s - %s", step, tc.name, "balances do not match")
			uassert.Equal(t, tc.balance, tc.fn(), reason)
		}
	}

	// admin creates FOO and BAR.
	testing.SetOriginCaller(admin)
	NewWithAdmin(cross, "Foo", "FOO", 3, 1_111_111_000, 5_555, admin)
	NewWithAdmin(cross, "Bar", "BAR", 3, 2_222_000, 6_666, admin)
	checkBalances("step1", 1_111_111_000, 1_111_111_000, 0, 0, 0)

	// admin mints to bob.
	mustGetInstance("FOO").ledger.Mint(bob, 333_333_000)
	checkBalances("step2", 1_444_444_000, 1_111_111_000, 333_333_000, 0, 0)

	// carl uses the faucet.
	testing.SetOriginCaller(carl)
	Faucet(cross, "FOO")
	checkBalances("step3", 1_444_449_555, 1_111_111_000, 333_333_000, 0, 5_555)

	// admin gives to bob some allowance.
	testing.SetOriginCaller(admin)
	Approve(cross, "FOO", bob, 1_000_000)
	checkBalances("step4", 1_444_449_555, 1_111_111_000, 333_333_000, 1_000_000, 5_555)

	// bob uses a part of the allowance.
	testing.SetOriginCaller(bob)
	TransferFrom(cross, "FOO", admin, carl, 400_000)
	checkBalances("step5", 1_444_449_555, 1_110_711_000, 333_333_000, 600_000, 405_555)

	// bob uses a part of the allowance.
	testing.SetOriginCaller(bob)
	TransferFrom(cross, "FOO", admin, carl, 600_000)
	checkBalances("step6", 1_444_449_555, 1_110_111_000, 333_333_000, 0, 1_005_555)
}
