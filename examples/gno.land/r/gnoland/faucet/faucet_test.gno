package faucet

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/faucet"
)

func TestPackage(t *testing.T) {
	var (
		adminaddr        = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
		faucetaddr       = std.DerivePkgAddr("gno.land/r/gnoland/faucet")
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
	std.TestIssueCoins(faucetaddr, std.Coins{{"ugnot", 1000000000}})
	assertBalance(t, faucetaddr, 1200000000)

	// by default, balance is empty, and as a user I cannot call Transfer, or Admin commands.

	assertBalance(t, test1addr, 0)
	std.TestSetOrigCaller(test1addr)
	assertErr(t, faucet.Transfer(test1addr, 1000000))

	assertErr(t, faucet.AdminAddController(controlleraddr1))
	std.TestSetOrigCaller(controlleraddr1)
	assertErr(t, faucet.Transfer(test1addr, 1000000))

	// as an admin, add the controller to contract and deposit more 2000gnot to contract
	std.TestSetOrigCaller(adminaddr)
	assertNoErr(t, faucet.AdminAddController(controlleraddr1))
	assertBalance(t, faucetaddr, 1200000000)

	// now, send some tokens as controller.
	std.TestSetOrigCaller(controlleraddr1)
	assertNoErr(t, faucet.Transfer(test1addr, 1000000))
	assertBalance(t, test1addr, 1000000)
	assertNoErr(t, faucet.Transfer(test1addr, 1000000))
	assertBalance(t, test1addr, 2000000)
	assertBalance(t, faucetaddr, 1198000000)

	// remove controller
	// as an admin, remove controller
	std.TestSetOrigCaller(adminaddr)
	assertNoErr(t, faucet.AdminRemoveController(controlleraddr1))
	std.TestSetOrigCaller(controlleraddr1)
	assertErr(t, faucet.Transfer(test1addr, 1000000))

	// duplicate controller
	std.TestSetOrigCaller(adminaddr)
	assertNoErr(t, faucet.AdminAddController(controlleraddr1))
	assertErr(t, faucet.AdminAddController(controlleraddr1))
	// add more than more than allowed controllers
	assertNoErr(t, faucet.AdminAddController(controlleraddr2))
	assertNoErr(t, faucet.AdminAddController(controlleraddr3))
	assertNoErr(t, faucet.AdminAddController(controlleraddr4))
	assertNoErr(t, faucet.AdminAddController(controlleraddr5))
	assertNoErr(t, faucet.AdminAddController(controlleraddr6))
	assertNoErr(t, faucet.AdminAddController(controlleraddr7))
	assertNoErr(t, faucet.AdminAddController(controlleraddr8))
	assertNoErr(t, faucet.AdminAddController(controlleraddr9))
	assertNoErr(t, faucet.AdminAddController(controlleraddr10))
	assertErr(t, faucet.AdminAddController(controlleraddr11))

	// send more than per transfer limit
	std.TestSetOrigCaller(adminaddr)
	faucet.AdminSetTransferLimit(300000000)
	std.TestSetOrigCaller(controlleraddr1)
	assertErr(t, faucet.Transfer(test1addr, 301000000))

	// block transefer from the address not on the controllers list.
	std.TestSetOrigCaller(controlleraddr11)
	assertErr(t, faucet.Transfer(test1addr, 1000000))
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

func assertBalance(t *testing.T, addr std.Address, expectedBal int64) {
	t.Helper()
	banker := std.GetBanker(std.BankerTypeReadonly)
	coins := banker.GetCoins(addr)
	got := coins.AmountOf("ugnot")

	if expectedBal != got {
		t.Errorf("invalid balance: expected %d, got %d.", expectedBal, got)
	}
}
