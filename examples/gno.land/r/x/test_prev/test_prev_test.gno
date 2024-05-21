package test_prev

import (
	"std"
	"testing"

	"gno.land/r/demo/foo20"
)

func TestDoSomeWithUserBalancePoor(t *testing.T) {
	std.TestSetOrigCaller("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	if DoSomeWithUserBalance() != "poor user" {
		t.Error("expected poor user")
	}
}

func TestDoSomeWithUserBalanceRichButPoor(t *testing.T) {
	std.TestSetOrigCaller("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	foo20.Faucet() // foo20 will mint tokens to this realm(std.PrevRealm) not user
	if DoSomeWithUserBalance() != "poor user" {
		t.Error("expected poor user")
	}
}

func TestDoSomeWithUserBalanceRich(t *testing.T) {
	std.TestSetPrevAddr("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	foo20.Faucet() // foo20 will mint tokens to this realm not user
	if DoSomeWithUserBalance() != "rich user" {
		t.Error("expected rich user")
	}
}
