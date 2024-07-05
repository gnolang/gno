package bar20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestPackage(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // XXX: should not need this

	balance, _ := Token.BalanceOf(alice)
	expected := uint64(0)
	if balance != expected {
		t.Errorf("balance should be %d, got %d", expected, balance)
	}

	ret := Faucet()
	if ret != "OK" {
		t.Errorf("faucet should be OK, got %s", ret)
	}

	balance, _ = Token.BalanceOf(alice)
	expected = uint64(1_000_000)
	if balance != expected {
		t.Errorf("balance should be %d, got %d", expected, balance)
	}
}
