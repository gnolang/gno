package grc20reg

import (
	"std"

	"gno.land/p/demo/grc/grc20"
)

// FIXME; delete this file.
// we currently have a gnovm limitation:
//
//	panic: cannot modify external-realm or non-realm object
type tmpproxy struct {
	getName      func() string
	getSymbol    func() string
	getDecimals  func() uint
	totalSupply  func() uint64
	balanceOf    func(std.Address) (uint64, error)
	transfer     func(std.Address, uint64) error
	allowance    func(owner, spender std.Address) (uint64, error)
	approve      func(std.Address, uint64) error
	transferFrom func(from, to std.Address, amount uint64) error
}

func Proxify(input grc20.IGRC20) *tmpproxy {
	return &tmpproxy{
		getName:      input.GetName,
		getSymbol:    input.GetSymbol,
		getDecimals:  input.GetDecimals,
		totalSupply:  input.TotalSupply,
		balanceOf:    input.BalanceOf,
		transfer:     input.Transfer,
		allowance:    input.Allowance,
		approve:      input.Approve,
		transferFrom: input.TransferFrom,
	}
}

var _ grc20.IGRC20 = (*tmpproxy)(nil)

func (impl *tmpproxy) GetName() string                               { return impl.getName() }
func (impl *tmpproxy) GetSymbol() string                             { return impl.getSymbol() }
func (impl *tmpproxy) GetDecimals() uint                             { return impl.getDecimals() }
func (impl *tmpproxy) TotalSupply() uint64                           { return impl.totalSupply() }
func (impl *tmpproxy) BalanceOf(account std.Address) (uint64, error) { return impl.balanceOf(account) }
func (impl *tmpproxy) Transfer(to std.Address, amount uint64) error  { return impl.transfer(to, amount) }
func (impl *tmpproxy) Allowance(owner, spender std.Address) (uint64, error) {
	return impl.allowance(owner, spender)
}

func (impl *tmpproxy) Approve(spender std.Address, amount uint64) error {
	return impl.approve(spender, amount)
}

func (impl *tmpproxy) TransferFrom(from, to std.Address, amount uint64) error {
	return impl.transferFrom(from, to, amount)
}
