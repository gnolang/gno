package minidex

import (
	"std"
	"testing"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/foo20"
	"gno.land/r/demo/tests/test20"
)

func TestPackage(t *testing.T) {
	// Setup
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	dex := NewDEX()
	dexAddr := std.DerivePkgAddr("gno.land/r/demo/minidex")

	checkBalances := func(
		eatb, ebtb, edtb uint64,
		eafb, ebfb, edfb uint64,
	) {
		exp := ufmt.Sprintf("atb=%d btb=%d dtb=%d afb=%d bfb=%d dfb=%d", eatb, ebtb, edtb, eafb, ebfb, edfb)
		var (
			gatb = test20.Bank.BalanceOf(alice)
			gbtb = test20.Bank.BalanceOf(bob)
			gdtb = test20.Bank.BalanceOf(dexAddr)
			gafb = foo20.Bank.BalanceOf(alice)
			gbfb = foo20.Bank.BalanceOf(bob)
			gdfb = foo20.Bank.BalanceOf(dexAddr)
		)
		got := ufmt.Sprintf("atb=%d btb=%d dtb=%d afb=%d bfb=%d dfb=%d", gatb, gbtb, gdtb, gafb, gbfb, gdfb)
		uassert.Equal(t, exp, got, "balances")
	}
	checkAllowances := func(
		eabt, eadt, ebat, ebdt, edat, edbt uint64,
		eabf, eadf, ebaf, ebdf, edaf, edbf uint64,
	) {
		exp := ufmt.Sprintf("abt=%d adt=%d bat=%d bdt=%d dat=%d dbt=%d abf=%d adf=%d baf=%d bdf=%d daf=%d dbf=%d",
			eabt, eadt, ebat, ebdt, edat, edbt, eabf, eadf, ebaf, ebdf, edaf, edbf,
		)
		var (
			gabt = test20.Bank.Allowance(alice, bob)
			gadt = test20.Bank.Allowance(alice, dexAddr)
			gbat = test20.Bank.Allowance(bob, alice)
			gbdt = test20.Bank.Allowance(bob, dexAddr)
			gdat = test20.Bank.Allowance(dexAddr, alice)
			gdbt = test20.Bank.Allowance(dexAddr, bob)
			gabf = foo20.Bank.Allowance(alice, bob)
			gadf = foo20.Bank.Allowance(alice, dexAddr)
			gbaf = foo20.Bank.Allowance(bob, alice)
			gbdf = foo20.Bank.Allowance(bob, dexAddr)
			gdaf = foo20.Bank.Allowance(dexAddr, alice)
			gdbf = foo20.Bank.Allowance(dexAddr, bob)
		)
		got := ufmt.Sprintf("abt=%d adt=%d bat=%d bdt=%d dat=%d dbt=%d abf=%d adf=%d baf=%d bdf=%d daf=%d dbf=%d",
			gabt, gadt, gbat, gbdt, gdat, gdbt, gabf, gadf, gbaf, gbdf, gdaf, gdbf,
		)
		uassert.Equal(t, exp, got, "allowances")
	}

	std.TestSetOrigCaller(alice)
	std.TestSetRealm(std.NewUserRealm(alice))
	checkBalances(0, 0, 0, 0, 0, 0)

	test20.Adm.Mint(alice, 1000)
	foo20.Faucet()
	checkBalances(1000, 0, 0, 10000000, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	grc20.PrevRealmBanker(test20.Bank).Approve(dexAddr, 100)
	checkBalances(1000, 0, 0, 10000000, 0, 0)
	checkAllowances(0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	uassert.Equal(t, dex.Size(), 0)

	dex.PlaceOrder(test20.Bank, foo20.Bank, 100, 2)
	checkBalances(900, 0, 100, 10000000, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	uassert.Equal(t, dex.Size(), 1)

	uassert.Equal(t, dex.MatchPairOrders("foo20/test20"), 0)
	checkBalances(900, 0, 100, 10000000, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	uassert.Equal(t, dex.Size(), 1)

	std.TestSetOrigCaller(bob)
	std.TestSetRealm(std.NewUserRealm(bob))
	grc20.PrevRealmBanker(foo20.Bank).Approve(dexAddr, 2000)
	checkBalances(900, 0, 100, 10000000, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0, 0, 0, 0, 2000, 0, 0)
	uassert.Equal(t, dex.Size(), 1)

	println("_______________")
	return
	dex.PlaceOrder(foo20.Bank, test20.Bank, 500, 2)
	checkBalances(900, 0, 100, 10000000, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	uassert.Equal(t, dex.Size(), 2)
}
