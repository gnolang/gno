package minidex

import (
	"std"
	"testing"

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
		step string,
		eatb, ebtb, edtb uint64,
		eafb, ebfb, edfb uint64,
	) {
		exp := ufmt.Sprintf("atb=%d btb=%d dtb=%d afb=%d bfb=%d dfb=%d", eatb, ebtb, edtb, eafb, ebfb, edfb)
		var (
			gatb = test20.Token.BalanceOf(alice)
			gbtb = test20.Token.BalanceOf(bob)
			gdtb = test20.Token.BalanceOf(dexAddr)
			gafb = foo20.Token.BalanceOf(alice)
			gbfb = foo20.Token.BalanceOf(bob)
			gdfb = foo20.Token.BalanceOf(dexAddr)
		)
		got := ufmt.Sprintf("atb=%d btb=%d dtb=%d afb=%d bfb=%d dfb=%d", gatb, gbtb, gdtb, gafb, gbfb, gdfb)
		uassert.Equal(t, exp, got, step)
	}
	checkAllowances := func(
		step string,
		eabt, eadt, ebat, ebdt, edat, edbt uint64,
		eabf, eadf, ebaf, ebdf, edaf, edbf uint64,
	) {
		exp := ufmt.Sprintf("abt=%d adt=%d bat=%d bdt=%d dat=%d dbt=%d abf=%d adf=%d baf=%d bdf=%d daf=%d dbf=%d",
			eabt, eadt, ebat, ebdt, edat, edbt, eabf, eadf, ebaf, ebdf, edaf, edbf,
		)
		var (
			gabt = test20.Token.Allowance(alice, bob)
			gadt = test20.Token.Allowance(alice, dexAddr)
			gbat = test20.Token.Allowance(bob, alice)
			gbdt = test20.Token.Allowance(bob, dexAddr)
			gdat = test20.Token.Allowance(dexAddr, alice)
			gdbt = test20.Token.Allowance(dexAddr, bob)
			gabf = foo20.Token.Allowance(alice, bob)
			gadf = foo20.Token.Allowance(alice, dexAddr)
			gbaf = foo20.Token.Allowance(bob, alice)
			gbdf = foo20.Token.Allowance(bob, dexAddr)
			gdaf = foo20.Token.Allowance(dexAddr, alice)
			gdbf = foo20.Token.Allowance(dexAddr, bob)
		)
		got := ufmt.Sprintf("abt=%d adt=%d bat=%d bdt=%d dat=%d dbt=%d abf=%d adf=%d baf=%d bdf=%d daf=%d dbf=%d",
			gabt, gadt, gbat, gbdt, gdat, gdbt, gabf, gadf, gbaf, gbdf, gdaf, gdbf,
		)
		uassert.Equal(t, exp, got, step)
	}

	std.TestSetOrigCaller(alice)
	std.TestSetRealm(std.NewUserRealm(alice))
	checkBalances("step1", 0, 0, 0, 0, 0, 0)

	test20.AdminLedger.Mint(alice, 1000)
	checkBalances("step2", 1000, 0, 0, 0, 0, 0)
	checkAllowances("step3", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	println("ALICE", alice)
	println("DEXADDR", dexAddr)
	test20.Token.RealmTeller().Approve(dexAddr, 100)
	checkBalances("step4", 1000, 0, 0, 0, 0, 0)
	checkAllowances("step5", 0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	uassert.Equal(t, dex.Size(), 0, "A")

	uassert.Equal(t, dex.PlaceOrder(test20.Token, foo20.Token, 100, false), 0, "B")
	checkBalances("step6", 900, 0, 100, 0, 0, 0)
	checkAllowances("step7", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	uassert.Equal(t, dex.Size(), 1, "C")
	uassert.Equal(t, dex.Render(""), `Orders:
Sell: Test20 -> Foo 100

`)

	std.TestSetOrigCaller(bob)
	std.TestSetRealm(std.NewUserRealm(bob))
	foo20.Faucet()
	checkBalances("step8", 900, 0, 100, 0, 10000000, 0)
	checkAllowances("step9", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	foo20.Token.RealmTeller().Approve(dexAddr, 2000)
	checkBalances("step10", 900, 0, 100, 0, 10000000, 0)
	checkAllowances("step11", 0, 0, 0, 0, 0, 0, 0, 0, 0, 2000, 0, 0)
	uassert.Equal(t, dex.Size(), 1, "D")

	uassert.Equal(t, dex.PlaceOrder(foo20.Token, test20.Token, 500, true), 0, "E") // XXX: should be 1?
	checkBalances("step12", 900, 0, 100, 0, 9999500, 500)
	checkAllowances("step13", 0, 0, 0, 0, 0, 0, 0, 0, 0, 1500, 0, 0)
	uassert.Equal(t, dex.Size(), 2, "F") // XXX: should be 1?
	uassert.Equal(t, dex.Render(""), `Orders:
Sell: Test20 -> Foo 100

Buy: Foo -> Test20 500

`, "G")
}
