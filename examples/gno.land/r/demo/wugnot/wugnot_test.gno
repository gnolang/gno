package wugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/users"
)

var (
	alice           = testutils.TestAddress("alice") // g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh
	addrOrNameAlice = users.AddressOrName(alice)
)

func TestDeposit(t *testing.T) {
	std.TestSetOrigCaller(alice)
	std.TestIssueCoins(alice, std.NewCoins(std.NewCoin("ugnot", 1_000_000)))

	wrapAmount := uint64(100)

	balanceBefore := BalanceOf(addrOrNameAlice)

	if balanceBefore != 0 {
		t.Fatalf("Expected balance to be 0, got %d", balanceBefore)
	}

	// Send 100 ugnots to wrap
	std.TestSetOrigSend(std.NewCoins(std.NewCoin("ugnot", int64(wrapAmount))), nil)
	Deposit()

	balanceAfter := BalanceOf(addrOrNameAlice)

	if balanceAfter != uint64(wrapAmount) {
		t.Fatalf("Expected balance to be %d, got %d", balanceBefore+wrapAmount, balanceAfter)
	}
}

func TestWithdraw(t *testing.T) {
	// Deposit from previous test is carried over
	unwrapAmount := uint64(100)

	balanceBefore := BalanceOf(addrOrNameAlice)

	if balanceBefore != unwrapAmount {
		t.Fatalf("Expected balance to be %d, got %d", unwrapAmount, balanceBefore)
	}

	// Send 100 wugnots to unwrap
	Withdraw(unwrapAmount)

	balanceAfter := BalanceOf(addrOrNameAlice)

	if balanceAfter != 0 {
		t.Fatalf("Expected balance to be %d, got %d", balanceBefore-unwrapAmount, balanceAfter)
	}
}
