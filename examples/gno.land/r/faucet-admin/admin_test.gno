package admin

import (
	"fmt"
	"std"
	"testing"

	"gno.land/p/testutils"
	FA "gno.land/r/faucet-admin"
)

func TestPackage(t *testing.T) {
	var (
		adminAddr = std.Address("g1qsme7hy3whlhwjf2lfq93szez7jveqzj5gwexj") 
		controllerAddr = std.Address("g1tzgxd2e59hywmqtg0pd3nljff39j36sxu2flpg")
		newAdminAddr = testutils.TestAddress("new admin addr")
		newControllerAddr = testutils.TestAddress("new controller addr")
		_tmp1 = testutils.TestAddress("tmp 1")
		_tmp2 = testutils.TestAddress("tmp 2")
	)

	// test for getter in /r/faucet-admin
	shouldEqual(t, AdminAddress(), "g1qsme7hy3whlhwjf2lfq93szez7jveqzj5gwexj")
	shouldEqual(t, ControllerAddress(), "g1tzgxd2e59hywmqtg0pd3nljff39j36sxu2flpg")
	shouldEqual(t, InPause(), false)
	shouldEqual(t, Message(), "# Let's Go for Gno Faucet by Onbloc & b1t\n\n")

	// TODO (struct compare fix)
	// why is it failing ?_?
	// shouldEqual(t, GetPerTransferSend(_tmp1), std.Coins{std.Coin{"ugnot", 1000000}})

	
	// normal address
	assertErr(t, FA.SetAdminInPause(true))
	assertErr(t, FA.SetAdminMessage("test"))
	assertErr(t, FA.SetAdminPerTransferSend(std.Coins{std.Coin{"ugnot", 100000}})) 
	assertErr(t, FA.SetAdminAdminAddr(newAdminAddr))
	assertErr(t, FA.SetAdminControllerAddr(newControllerAddr))

	// controller address
	std.TestSetOrigCaller(controllerAddr)
	assertErr(t, FA.SetAdminInPause(true))
	assertErr(t, FA.SetAdminMessage("test"))
	assertErr(t, FA.SetAdminPerTransferSend(std.Coins{std.Coin{"ugnot", 200000}})) 
	assertErr(t, FA.SetAdminAdminAddr(newAdminAddr))
	assertErr(t, FA.SetAdminControllerAddr(newControllerAddr))

	// admin address
	std.TestSetOrigCaller(adminAddr)
	assertNoErr(t, FA.SetAdminInPause(true))
	assertNoErr(t, FA.SetAdminMessage("test"))
	assertNoErr(t, FA.SetAdminPerTransferSend(std.Coins{std.Coin{"ugnot", 300000}}))
	
	// set new admin address
	assertNoErr(t, FA.SetAdminAdminAddr(newAdminAddr))
	// test old admin address
	assertErr(t, FA.SetAdminControllerAddr(newControllerAddr))

	// new admin address
	std.TestSetOrigCaller(newAdminAddr)
	assertNoErr(t, FA.SetAdminInPause(true))
	assertNoErr(t, FA.SetAdminMessage("test"))
	assertNoErr(t, FA.SetAdminPerTransferSend(std.Coins{std.Coin{"ugnot", 400000}}))
	assertNoErr(t, FA.SetAdminControllerAddr(newControllerAddr))
}


func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		// t.Logf("expected an error, but got nil.")
		t.Errorf("expected an error, but got nil.")
	}
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		// t.Logf("expected no error, but got err: %v", err.Error())
		t.Errorf("expected no error, but got err: %v", err.Error())
	}
}

// TODO
// struct compare failing
func shouldEqual(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()

	if got != expected {
		t.Errorf("expected %v(type: %T), got %v(type: %T)", expected, expected, got, got)
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