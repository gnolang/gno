package grc20

import (
	"std"

	"gno.land/p/demo/grc/exts"
)

// IMustGRC20 is like IGRC20 but without returned errors.
//
// It will either panic or silently ignore invalid usages,
// depending on the method.
type IMustGRC20 interface {
	exts.TokenMetadata
	TotalSupply() uint64
	BalanceOf(account std.Address) uint64
	Transfer(to std.Address, amount uint64)
	Allowance(owner, spender std.Address) uint64
	Approve(spender std.Address, amount uint64)
	TransferFrom(from, to std.Address, amount uint64)
}
