package faucet

import (
	"fmt"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/faucet"
)

func TestPackage(t *testing.T) {
	var (
		adminaddr      = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
		controlleraddr = testutils.TestAddress("controller")
		test1addr      = testutils.TestAddress("test1")
	)

	// by default, balance is empty, and as a user I cannot call Transfer, or Admin commands.
	std.TestSetOrigSend(std.Coins{{"ugnot", 2000000000}}, nil)
	assertBalance(t, test1addr, 0)
	assertErr(t, faucet.Transfer(test1addr))
	assertErr(t, faucet.AdminSetControllerAddr(controlleraddr))
	std.TestSetOrigCaller(controlleraddr)
	assertErr(t, faucet.Transfer(test1addr))

	// as an admin, set the controller to controlleraddr.
	std.TestSetOrigCaller(adminaddr)
	assertNoErr(t, faucet.AdminSetControllerAddr(controlleraddr))

	// now, send some tokens as controller.
	std.TestSetOrigCaller(controlleraddr)
	assertNoErr(t, faucet.Transfer(test1addr))
	assertBalance(t, test1addr, 1000000)
	assertNoErr(t, faucet.Transfer(test1addr))
	assertBalance(t, test1addr, 2000000)
}

func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Logf("info: got err: %v", err)
		t.Errorf("expected an error, got nil.")
	}
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
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
