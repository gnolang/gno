package registry

import (
	"std"
	"testing"

	"gno.land/r/x/grc20_dynamic_call/bar"
	"gno.land/r/x/grc20_dynamic_call/baz"
	"gno.land/r/x/grc20_dynamic_call/foo"

	_ "gno.land/r/x/grc20_dynamic_call/wrapper"

	"gno.land/p/demo/testutils"

	"gno.land/r/demo/users"
)

var (
	cRealmAddr std.Address

	transferTo std.Address

	transferFromFrom std.Address
	transferFromTo   std.Address
)

func init() {
	cRealmAddr = std.CurrentRealm().Addr()

	transferTo = testutils.TestAddress("transferTo")

	transferFromFrom = transferTo // use transferTo as transferFromFrom
	transferFromTo = testutils.TestAddress("transferFromTo")

	std.TestSetOrigCaller(cRealmAddr)
	foo.Faucet()
	bar.Faucet()
	baz.Faucet()
}

func TestTransferByNameFOO(t *testing.T) {
	fooBalance := BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/foo", transferTo)
	if fooBalance != 0 {
		t.Fatal("transferTo should have 0 FOO but", fooBalance)
	}

	TransferByInterfaceName("gno.land/r/x/grc20_dynamic_call/foo", transferTo, 12345)

	fooBalance = BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/foo", transferTo)
	if fooBalance != 12345 {
		t.Fatal("transferTo should have 12345 FOO, but", fooBalance)
	}
}

func TestTransferByNameBAR(t *testing.T) {
	barBalance := BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/bar", transferTo)
	if barBalance != 0 {
		t.Fatal("transferTo should have 0 BAR but", barBalance)
	}

	TransferByInterfaceName("gno.land/r/x/grc20_dynamic_call/bar", transferTo, 12345)

	barBalance = BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/bar", transferTo)
	if barBalance != 12345 {
		t.Fatal("transferTo should have 12345 BAR, but", barBalance)
	}
}

func TestTransferByNameBAZ(t *testing.T) {
	bazBalance := BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/baz", transferTo)
	if bazBalance != 0 {
		t.Fatal("transferTo should have 0 BAZ but", bazBalance)
	}

	TransferByInterfaceName("gno.land/r/x/grc20_dynamic_call/baz", transferTo, 12345)

	bazBalance = BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/baz", transferTo)
	if bazBalance != 12345 {
		t.Fatal("transferTo should have 12345 BAZ, but", bazBalance)
	}
}

func TestTransferFromByNameFOO(t *testing.T) {
	fooBalanceTo := BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/foo", transferFromTo)
	if fooBalanceTo != 0 {
		t.Fatal("transferFromTo should have 0 FOO, but", fooBalanceTo)
	}

	std.TestSetOrigCaller(transferFromFrom)
	foo.Approve(users.AddressOrName(cRealmAddr), 5)
	TransferFromByInterfaceName("gno.land/r/x/grc20_dynamic_call/foo", transferFromFrom, transferFromTo, 3)

	fooBalanceTo = BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/foo", transferFromTo)
	if fooBalanceTo != 3 {
		t.Fatal("transferFromTo should have 3 FOO, but", fooBalanceTo)
	}

	fooBalanceFrom := BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/foo", transferFromFrom)
	if fooBalanceFrom != 12342 {
		t.Fatal("transferFromFrom should have 12342 FOO, but", fooBalanceFrom)
	}
}

func TestTransferFromByNameBAR(t *testing.T) {
	barBalanceTo := BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/bar", transferFromTo)
	if barBalanceTo != 0 {
		t.Fatal("transferFromTo should have 0 BAR, but", barBalanceTo)
	}

	std.TestSetOrigCaller(transferFromFrom)
	bar.Approve(users.AddressOrName(cRealmAddr), 5)
	TransferFromByInterfaceName("gno.land/r/x/grc20_dynamic_call/bar", transferFromFrom, transferFromTo, 3)

	barBalanceTo = BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/bar", transferFromTo)
	if barBalanceTo != 3 {
		t.Fatal("transferFromTo should have 3 BAR, but", barBalanceTo)
	}

	barBalanceFrom := BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/bar", transferFromFrom)
	if barBalanceFrom != 12342 {
		t.Fatal("transferFromFrom should have 12342 BAR, but", barBalanceFrom)
	}
}

func TestTransferFromByNameBAZ(t *testing.T) {
	bazBalanceTo := BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/baz", transferFromTo)
	if bazBalanceTo != 0 {
		t.Fatal("transferFromTo should have 0 BAZ, but", bazBalanceTo)
	}

	std.TestSetOrigCaller(transferFromFrom)
	baz.Approve(users.AddressOrName(cRealmAddr), 5)
	TransferFromByInterfaceName("gno.land/r/x/grc20_dynamic_call/baz", transferFromFrom, transferFromTo, 3)

	bazBalanceTo = BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/baz", transferFromTo)
	if bazBalanceTo != 3 {
		t.Fatal("transferFromTo should have 3 BAZ, but", bazBalanceTo)
	}

	bazBalanceFrom := BalanceOfByInterfaceName("gno.land/r/x/grc20_dynamic_call/baz", transferFromFrom)
	if bazBalanceFrom != 12342 {
		t.Fatal("transferFromFrom should have 12342 BAZ, but", bazBalanceFrom)
	}
}

func TestUnregisterUnauthorized(t *testing.T) {
	shouldPanic(t, func() { UnregisterGRC20Interface("gno.land/r/x/grc20_dynamic_call/foo") })
}

func TestUnregisterAuthorized(t *testing.T) {
	if len(registered) != 3 {
		t.Fatal("should have 3 registered interfaces but", len(registered))
	}

	std.TestSetOrigCaller("g1sqt92sa06ugh8nlt98kyghw83qy84paf4csyh6")
	UnregisterGRC20Interface("gno.land/r/x/grc20_dynamic_call/foo")

	if len(registered) != 2 {
		t.Fatal("should have 2 registered interfaces")
	}
}

func shouldPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	f()
}
