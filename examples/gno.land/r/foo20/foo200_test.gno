package foo20

import (
	"testing"

	"gno.land/r/users"
)

func TestReadOnlyPublicMethods(t *testing.T) {
	var admin = users.AddressOrName("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	t.Logf("admin: %v", admin)

	if TotalSupply() != 10100000000 {
		t.Errorf("TotalSupply: want: %d have: %d", 10100000000, TotalSupply())
	}

	if BalanceOf(admin) == 10000000000 {
		t.Errorf("BalanceOf(admin): want: %d have: %d", 10000000000, BalanceOf(admin))
	}
}
