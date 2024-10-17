package grc20

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestBankImpl(t *testing.T) {
	bank, _ := NewToken("Dummy", "DUMMY", 4)
	urequire.False(t, bank == nil, "dummy should not be nil")
}

func TestBank(t *testing.T) {
	var (
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
		carl  = testutils.TestAddress("carl")
	)

	bank, adm := NewToken("Dummy", "DUMMY", 6)

	checkBalances := func(aliceEB, bobEB, carlEB uint64) {
		t.Helper()
		exp := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceEB, bobEB, carlEB)
		aliceGB := bank.BalanceOf(alice)
		bobGB := bank.BalanceOf(bob)
		carlGB := bank.BalanceOf(carl)
		got := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceGB, bobGB, carlGB)
		uassert.Equal(t, got, exp, "invalid balances")
	}
	checkAllowances := func(abEB, acEB, baEB, bcEB, caEB, cbEB uint64) {
		t.Helper()
		exp := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abEB, acEB, baEB, bcEB, caEB, cbEB)
		abGB := bank.Allowance(alice, bob)
		acGB := bank.Allowance(alice, carl)
		baGB := bank.Allowance(bob, alice)
		bcGB := bank.Allowance(bob, carl)
		caGB := bank.Allowance(carl, alice)
		cbGB := bank.Allowance(carl, bob)
		got := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abGB, acGB, baGB, bcGB, caGB, cbGB)
		uassert.Equal(t, got, exp, "invalid allowances")
	}

	checkBalances(0, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.Mint(alice, 1000))
	urequire.NoError(t, adm.Mint(alice, 100))
	checkBalances(1100, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.Approve(alice, bob, 99999999))
	checkBalances(1100, 0, 0)
	checkAllowances(99999999, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.Approve(alice, bob, 400))
	checkBalances(1100, 0, 0)
	checkAllowances(400, 0, 0, 0, 0, 0)

	urequire.Error(t, adm.TransferFrom(alice, bob, carl, 100000000))
	checkBalances(1100, 0, 0)
	checkAllowances(400, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.TransferFrom(alice, bob, carl, 100))
	checkBalances(1000, 0, 100)
	checkAllowances(300, 0, 0, 0, 0, 0)

	urequire.Error(t, adm.SpendAllowance(alice, bob, 2000000))
	checkBalances(1000, 0, 100)
	checkAllowances(300, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.SpendAllowance(alice, bob, 100))
	checkBalances(1000, 0, 100)
	checkAllowances(200, 0, 0, 0, 0, 0)
}
