package registry

import (
	"std"

	"gno.land/r/demo/users"
)

const APPROVED_UNREGISTER_CALLER = "g1sqt92sa06ugh8nlt98kyghw83qy84paf4csyh6"

var registered = []GRC20Pair{}

type GRC20Interface interface {
	Transfer() func(to users.AddressOrName, amount uint64)
	TransferFrom() func(from, to users.AddressOrName, amount uint64)
	BalanceOf() func(owner users.AddressOrName) uint64
}

type GRC20Pair struct {
	pkgPath string
	igrc20  GRC20Interface
}

func findGRC20(pkgPath string) (int, bool) {
	for i, pair := range registered {
		if pair.pkgPath == pkgPath {
			return i, true
		}
	}

	return -1, false
}

func appendGRC20Interface(pkgPath string, igrc20 GRC20Interface) {
	registered = append(registered, GRC20Pair{pkgPath: pkgPath, igrc20: igrc20})
}

func removeGRC20Interface(pkgPath string) {
	i, found := findGRC20(pkgPath)
	if !found {
		return
	}

	registered = append(registered[:i], registered[i+1:]...)
}

func RegisterGRC20Interface(pkgPath string, igrc20 GRC20Interface) {
	_, found := findGRC20(pkgPath)
	if found {
		panic("GRC20 already registered")
	}

	appendGRC20Interface(pkgPath, igrc20)
}

func UnregisterGRC20Interface(pkgPath string) {
	// do not allow realm to unregister
	std.AssertOriginCall()
	caller := std.GetOrigCaller()

	if caller != APPROVED_UNREGISTER_CALLER {
		panic("unauthorized")
	}

	_, found := findGRC20(pkgPath)
	if found {
		removeGRC20Interface(pkgPath)
	}
}

func TransferByInterfaceName(pkgPath string, to string, amount uint64) bool {
	i, found := findGRC20(pkgPath)
	if !found {
		return false
	}

	registered[i].igrc20.Transfer()(users.AddressOrName(to), amount)

	return true
}

func TransferFromByInterfaceName(pkgPath string, from, to string, amount uint64) bool {
	i, found := findGRC20(pkgPath)
	if !found {
		return false
	}

	registered[i].igrc20.TransferFrom()(users.AddressOrName(from), users.AddressOrName(to), amount)

	return true
}

func BalanceOfByInterfaceName(pkgPath string, owner string) uint64 {
	i, found := findGRC20(pkgPath)
	if !found {
		return 0
	}

	balance := registered[i].igrc20.BalanceOf()(users.AddressOrName(owner))
	return balance
}
