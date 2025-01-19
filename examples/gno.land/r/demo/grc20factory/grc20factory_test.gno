package foo20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	std.TestSetOriginPkgAddress("gno.land/r/demo/grc20factory")
	admin := testutils.TestAddress("admin")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")

	type test struct {
		name    string
		balance uint64
		fn      func() uint64
	}

	checkBalances := func(step string, totSup, balAdm, balBob, allowAdmBob, balCarl uint64) {
		tests := []test{
			{"TotalSupply", totSup, func() uint64 { return TotalSupply("FOO") }},
			{"BalanceOf(admin)", balAdm, func() uint64 { return BalanceOf("FOO", admin) }},
			{"BalanceOf(bob)", balBob, func() uint64 { return BalanceOf("FOO", bob) }},
			{"Allowance(admin, bob)", allowAdmBob, func() uint64 { return Allowance("FOO", admin, bob) }},
			{"BalanceOf(carl)", balCarl, func() uint64 { return BalanceOf("FOO", carl) }},
		}
		for _, tc := range tests {
			reason := ufmt.Sprintf("%s.%s - %s", step, tc.name, "balances do not match")
			uassert.Equal(t, tc.balance, tc.fn(), reason)
		}
	}

	// admin creates FOO and BAR.
	std.TestSetOriginCaller(admin)
	std.TestSetRealm(std.NewUserRealm(admin))
	NewWithAdmin("Foo", "FOO", 3, 1_111_111_000, 5_555, admin)
	NewWithAdmin("Bar", "BAR", 3, 2_222_000, 6_666, admin)
	checkBalances("step1", 1_111_111_000, 1_111_111_000, 0, 0, 0)

	// admin mints to bob.
	mustGetInstance("FOO").ledger.Mint(bob, 333_333_000)
	checkBalances("step2", 1_444_444_000, 1_111_111_000, 333_333_000, 0, 0)

	// carl uses the faucet.
	std.TestSetOriginCaller(carl)
	std.TestSetRealm(std.NewUserRealm(carl))
	Faucet("FOO")
	checkBalances("step3", 1_444_449_555, 1_111_111_000, 333_333_000, 0, 5_555)

	// admin gives to bob some allowance.
	std.TestSetOriginCaller(admin)
	std.TestSetRealm(std.NewUserRealm(admin))
	Approve("FOO", bob, 1_000_000)
	checkBalances("step4", 1_444_449_555, 1_111_111_000, 333_333_000, 1_000_000, 5_555)

	// bob uses a part of the allowance.
	std.TestSetOriginCaller(bob)
	std.TestSetRealm(std.NewUserRealm(bob))
	TransferFrom("FOO", admin, carl, 400_000)
	checkBalances("step5", 1_444_449_555, 1_110_711_000, 333_333_000, 600_000, 405_555)

	// bob uses a part of the allowance.
	std.TestSetOriginCaller(bob)
	std.TestSetRealm(std.NewUserRealm(bob))
	TransferFrom("FOO", admin, carl, 600_000)
	checkBalances("step6", 1_444_449_555, 1_110_111_000, 333_333_000, 0, 1_005_555)
}
