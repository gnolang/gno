package faucet

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/p/nt/testutils"
)

func TestPackage(t *testing.T) {
	var (
		adminaddr        = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
		faucetaddr       = chain.PackageAddress("gno.land/r/gnoland/faucet")
		controlleraddr1  = testutils.TestAddress("controller1")
		controlleraddr2  = testutils.TestAddress("controller2")
		controlleraddr3  = testutils.TestAddress("controller3")
		controlleraddr4  = testutils.TestAddress("controller4")
		controlleraddr5  = testutils.TestAddress("controller5")
		controlleraddr6  = testutils.TestAddress("controller6")
		controlleraddr7  = testutils.TestAddress("controller7")
		controlleraddr8  = testutils.TestAddress("controller8")
		controlleraddr9  = testutils.TestAddress("controller9")
		controlleraddr10 = testutils.TestAddress("controller10")
		controlleraddr11 = testutils.TestAddress("controller11")

		test1addr = testutils.TestAddress("test1")
	)
	// deposit 1000gnot to faucet contract
	testing.IssueCoins(faucetaddr, chain.Coins{{"ugnot", 1_000_000_000}})
	assertBalance(t, faucetaddr, 1_000_000_000)

	// by default, balance is empty, and as a user I cannot call Transfer, or Admin commands.
	assertBalance(t, test1addr, 0)
	testing.SetRealm(testing.NewUserRealm(test1addr))
	assertErr(t, Transfer(cross, test1addr, 1_000_000))

	assertErr(t, AdminAddController(cross, controlleraddr1))
	testing.SetRealm(testing.NewUserRealm(controlleraddr1))
	assertErr(t, Transfer(cross, test1addr, 1_000_000))

	// as an admin, add the controller to contract and deposit more 2000gnot to contract
	testing.SetRealm(testing.NewUserRealm(adminaddr))
	assertNoErr(t, AdminAddController(cross, controlleraddr1))
	assertBalance(t, faucetaddr, 1_000_000_000)

	// now, send some tokens as controller.
	testing.SetRealm(testing.NewUserRealm(controlleraddr1))
	assertNoErr(t, Transfer(cross, test1addr, 1_000_000))
	assertBalance(t, test1addr, 1_000_000)
	assertNoErr(t, Transfer(cross, test1addr, 1_000_000))
	assertBalance(t, test1addr, 2_000_000)
	assertBalance(t, faucetaddr, 998_000_000)

	// remove controller
	// as an admin, remove controller
	testing.SetRealm(testing.NewUserRealm(adminaddr))
	assertNoErr(t, AdminRemoveController(cross, controlleraddr1))
	testing.SetRealm(testing.NewUserRealm(controlleraddr1))
	assertErr(t, Transfer(cross, test1addr, 1_000_000))

	// duplicate controller
	testing.SetRealm(testing.NewUserRealm(adminaddr))
	assertNoErr(t, AdminAddController(cross, controlleraddr1))
	assertErr(t, AdminAddController(cross, controlleraddr1))
	// add more than more than allowed controllers
	assertNoErr(t, AdminAddController(cross, controlleraddr2))
	assertNoErr(t, AdminAddController(cross, controlleraddr3))
	assertNoErr(t, AdminAddController(cross, controlleraddr4))
	assertNoErr(t, AdminAddController(cross, controlleraddr5))
	assertNoErr(t, AdminAddController(cross, controlleraddr6))
	assertNoErr(t, AdminAddController(cross, controlleraddr7))
	assertNoErr(t, AdminAddController(cross, controlleraddr8))
	assertNoErr(t, AdminAddController(cross, controlleraddr9))
	assertNoErr(t, AdminAddController(cross, controlleraddr10))
	assertErr(t, AdminAddController(cross, controlleraddr11))

	// send more than per transfer limit
	testing.SetRealm(testing.NewUserRealm(adminaddr))
	assertNoErr(t, AdminSetTransferLimit(cross, 300_000_000))
	testing.SetRealm(testing.NewUserRealm(controlleraddr1))
	assertErr(t, Transfer(cross, test1addr, 301_000_000))

	// block transefer from the address not on the controllers list.
	testing.SetRealm(testing.NewUserRealm(controlleraddr11))
	assertErr(t, Transfer(cross, test1addr, 1_000_000))
}

func assertErr(t *testing.T, err string) {
	t.Helper()

	if err == "" {
		t.Logf("info: got err: %v", err)
		t.Errorf("expected an error, got nil.")
	}
}

func assertNoErr(t *testing.T, err string) {
	t.Helper()
	if err != "" {
		t.Errorf("got err: %v.", err)
	}
}

func assertBalance(t *testing.T, addr address, expectedBal int64) {
	t.Helper()
	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	coins := banker_.GetCoins(addr)
	got := coins.AmountOf("ugnot")

	if expectedBal != got {
		t.Errorf("invalid balance: expected %d, got %d.", expectedBal, got)
	}
}
