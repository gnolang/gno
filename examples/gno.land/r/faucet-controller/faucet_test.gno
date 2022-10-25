package faucet

import (
	"fmt"
	"std"
	"testing"

	"gno.land/p/testutils"
	faucet "gno.land/r/faucet-controller" // Actual Transfer
	FA "gno.land/r/faucet-admin"
)

func TestPackage(t *testing.T) {
	var (
		faucetAdminAddr = FA.GetAdminAddress()
		faucetControllerAddr = FA.GetControllerAddress()
		faucetPause	= FA.GetInPause()
		faucetMsg = FA.GetMessage()
		_tmpAdminAddr	= testutils.TestAddress("just tmp admin")
		_tmpControllerAddr = testutils.TestAddress("just tmp controller")
		_tmpUserAddr1 = testutils.TestAddress("just tmp user 01")
		_tmpUserAddr2 = testutils.TestAddress("just tmp user 02")
		_tmpUserAddr3 = testutils.TestAddress("just tmp user 03")

		// faucetAddr = std.TestDerivePkgAddr("gno.land/r/faucet-controller")		
	)

	// deposit 20,000gnot to facuet 
	std.TestSetOrigSend(std.Coins{{"ugnot", 20000000000}}, nil)
	// std.TestSetOrigPkgAddr(faucetAddr)
	// assertBalance(t, faucetAddr, 20000000000)
	
	// as a normal user
	// balance should empty
	// can't call transfer(controller command) or admin commands
	std.TestSetOrigCaller(_tmpUserAddr1)
	assertBalance(t, _tmpUserAddr1, 0)
	assertErr(t, faucet.Transfer(_tmpUserAddr1))
	assertErr(t, FA.SetAdminInPause(true))

	// as a controll user
	// transfer available, but not admin commands
	std.TestSetOrigCaller(faucetControllerAddr)
	assertBalance(t, _tmpUserAddr1, 0)
	assertBalance(t, _tmpUserAddr2, 0)
	
	assertNoErr(t, faucet.Transfer(_tmpUserAddr1))
	assertBalance(t, _tmpUserAddr1, 10000000)

	assertNoErr(t, faucet.Transfer(_tmpUserAddr2))
	assertBalance(t, _tmpUserAddr2, 10000000)

	// as an admin user
	// admin commads available, but not transfer
	std.TestSetOrigCaller(faucetAdminAddr)
	assertErr(t, faucet.Transfer(_tmpUserAddr1))

	// change controller address
	assertNoErr(t, FA.SetAdminControllerAddr(_tmpControllerAddr)) 
	// test old controller 
	std.TestSetOrigCaller(faucetControllerAddr)
	assertErr(t, faucet.Transfer(_tmpUserAddr1))

	// as an new controller, test faucet rules
	std.TestSetOrigCaller(_tmpControllerAddr)

	assertNoErr(t, faucet.Transfer(_tmpUserAddr3))  // 1st faucet gives you 10gnot
	assertBalance(t, _tmpUserAddr3, 10000000) // + 10,000,000

	assertNoErr(t, faucet.Transfer(_tmpUserAddr3)) // since 2nd faucet gives you 1gnot
	assertBalance(t, _tmpUserAddr3, 11000000) // + 1,000,000
}
 
// check response
func shouldEqual(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()

	if got != expected {
		t.Errorf("expected %v(%T), got %v(%T)", expected, expected, got, got)
	}
}

func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error, but got nil.")
		// t.Logf("info: got err: %v", err)
		// t.Errorf("expected an error, but got nil.")
	}
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		println("expected no error, but got err:", err)
		// t.Errorf("expected no error, but got err: ", err)
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