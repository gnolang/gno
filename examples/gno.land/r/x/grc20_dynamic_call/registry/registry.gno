package registry

import (
	"std"

	"gno.land/p/demo/ufmt"

	"gno.land/r/demo/users"
)

const (
	APPROVED_UNREGISTER_CALLER = "g1sqt92sa06ugh8nlt98kyghw83qy84paf4csyh6" // like admin
)

type GRC20Interface interface {
	Transfer() func(to users.AddressOrName, amount uint64)
	TransferFrom() func(from, to users.AddressOrName, amount uint64)
	BalanceOf() func(owner users.AddressOrName) uint64
}

var (
	registered = make(map[string]GRC20Interface)
)

func RegisterGRC20Interface(pkgPath string, igrc20 GRC20Interface) {
	_, found := registered[pkgPath]
	if found {
		panic(ufmt.Sprintf("GRC20(pkgPath:%s) already registered", pkgPath))
	}

	registered[pkgPath] = igrc20
}

func UnregisterGRC20Interface(pkgPath string) {
	// do not allow realm to unregister
	std.AssertOriginCall()

	caller := std.GetOrigCaller()
	if caller != APPROVED_UNREGISTER_CALLER {
		panic("unauthorized")
	}

	_, found := registered[pkgPath]
	if found {
		delete(registered, pkgPath)
	}
}

func BalanceOfByInterface(pkgPath string, owner std.Address) uint64 {
	igrc20, found := registered[pkgPath]
	if !found {
		return false
	}

	return igrc20.BalanceOf()(users.AddressOrName(owner))
}

func TransferByInterface(pkgPath string, to std.Address, amount uint64) bool {
	igrc20, found := registered[pkgPath]
	if !found {
		return false
	}

	// will panic if transfer fails
	igrc20.Transfer()(users.AddressOrName(to), amount)

	return true
}

func TransferFromByInterface(pkgPath string, from, to std.Address, amount uint64) bool {
	igrc20, found := registered[pkgPath]
	if !found {
		return false
	}

	// will panic if transferFrom fails
	igrc20.TransferFrom()(users.AddressOrName(from), users.AddressOrName(to), amount)

	return true
}
