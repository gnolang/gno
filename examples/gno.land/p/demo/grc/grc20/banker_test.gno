package grc20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestBankImpl(t *testing.T) {
	bank := NewBank("Dummy", "DUMMY", 4)
	urequire.False(t, bank == nil, "dummy should not be nil")
}

func TestPrevRealmBankerImpl(t *testing.T) {
	bank := NewBank("Dummy", "DUMMY", 4)
	banker := PrevRealmBanker(bank)
	urequire.False(t, banker == nil)
	var _ GRC20 = banker
}

func TestBank(t *testing.T) {
	var (
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
		carl  = testutils.TestAddress("carl")
	)

	b := NewBank("Dummy", "DUMMY", 6)

	checkBalances := func(aliceEB, bobEB, carlEB uint64) {
		t.Helper()
		exp := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceEB, bobEB, carlEB)
		aliceGB := b.BalanceOf(alice)
		bobGB := b.BalanceOf(bob)
		carlGB := b.BalanceOf(carl)
		got := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceGB, bobGB, carlGB)
		uassert.Equal(t, got, exp, "invalid balances")
	}
	checkAllowances := func(abEB, acEB, baEB, bcEB, caEB, cbEB uint64) {
		t.Helper()
		exp := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abEB, acEB, baEB, bcEB, caEB, cbEB)
		abGB := b.Allowance(alice, bob)
		acGB := b.Allowance(alice, carl)
		baGB := b.Allowance(bob, alice)
		bcGB := b.Allowance(bob, carl)
		caGB := b.Allowance(carl, alice)
		cbGB := b.Allowance(carl, bob)
		got := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abGB, acGB, baGB, bcGB, caGB, cbGB)
		uassert.Equal(t, got, exp, "invalid allowances")
	}

	checkBalances(0, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	urequire.NoError(t, b.Mint(alice, 1000))
	urequire.NoError(t, b.Mint(alice, 100))
	checkBalances(1100, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	urequire.NoError(t, b.approve(alice, bob, 99999999))
	checkBalances(1100, 0, 0)
	checkAllowances(99999999, 0, 0, 0, 0, 0)

	urequire.NoError(t, b.approve(alice, bob, 400))
	checkBalances(1100, 0, 0)
	checkAllowances(400, 0, 0, 0, 0, 0)

	urequire.Error(t, b.transferFrom(alice, bob, carl, 100000000))
	checkBalances(1100, 0, 0)
	checkAllowances(400, 0, 0, 0, 0, 0)

	urequire.NoError(t, b.transferFrom(alice, bob, carl, 100))
	checkBalances(1000, 0, 100)
	checkAllowances(300, 0, 0, 0, 0, 0)

	urequire.Error(t, b.spendAllowance(alice, bob, 2000000))
	checkBalances(1000, 0, 100)
	checkAllowances(300, 0, 0, 0, 0, 0)

	urequire.NoError(t, b.spendAllowance(alice, bob, 100))
	checkBalances(1000, 0, 100)
	checkAllowances(200, 0, 0, 0, 0, 0)
}

func TestPrevRealmBanker(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")

	bank := NewBank("Dummy", "DUMMY", 6)
	banker := PrevRealmBanker(bank)

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

	urequire.NoError(t, bank.Mint(alice, 1000))
	checkBalances(1000, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	std.TestSetOrigCaller(alice)
	urequire.Error(t, bank.Mint(alice, 1000))
	urequire.NoError(t, banker.Approve(bob, 600))
	checkBalances(1000, 0, 0)
	checkAllowances(600, 0, 0, 0, 0, 0)

	std.TestSetOrigCaller(bob)
	urequire.Error(t, banker.TransferFrom(alice, carl, 700))
	checkBalances(1000, 0, 0)
	checkAllowances(600, 0, 0, 0, 0, 0)
	urequire.NoError(t, banker.TransferFrom(alice, carl, 400))
	checkBalances(600, 0, 400)
	checkAllowances(200, 0, 0, 0, 0, 0)
}
