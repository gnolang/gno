package testutils

import "std"

func TestOrigSend(to std.Address, coins std.Coins, spent std.Coins) {
	// issue coins
	std.TestIssueCoins(to, coins)
	// set context variable
	std.TestSetOrigSend(coins, spent)
}
