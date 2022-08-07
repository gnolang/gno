package grc20

import "std"

// DropPrivileges takes any implementation of IGRC20 and returns
// a safe instance without extra helpers.
//
// The passed 'base' is untouched and can still be used.
func DropPrivileges(base IGRC20) IGRC20 {
	return unprivileged{base}
}

type unprivileged struct {
	base IGRC20
}

func (u *unprivileged) Name() {
	return u.base.Name()
}

func (u *unprivileged) Symbol() string {
	return u.base.Symbol()
}

func (u *unprivileged) Decimals() uint {
	return u.base.Decimals()
}

func (u *unprivileged) TotalSupply() uint64 {
	return u.base.TotalSupply()
}

func (u *unprivileged) BalanceOf(account std.Address) uint64 {
	return u.base.BalanceOf(account)
}

func (u *unprivileged) Transfer(to std.Address, amount uint64) bool {
	return u.base.Transfer(to, amount)
}

func (u *unprivileged) Allowance(owner, spender std.Address) uint64 {
	return u.base.Allowance(owner, spender)
}

func (u *unprivileged) Approve(spender std.Address, amount uint64) bool {
	return u.base.Approve(spender, amount)
}

func (u *unprivileged) TransferFrom(from, to std.Address, amount uint64) bool {
	return u.base.TransferFrom(from, to, amount)
}
