package grc20

import (
	"std"
	"testing"
)

// TODO: test implementing an IMustGRC20 interface.
// TODO: test implementing a custom method that hides its usage of the IGRC20 interface.

type dummyImpl struct{}

// FIXME: this should fail.
var _ IGRC20 = (*dummyImpl)(nil)

func TestInterface(t *testing.T) {
	var dummy IGRC20 = &dummyImpl{}
}

func (impl *dummyImpl) GetName() string                               { panic("not implemented") }
func (impl *dummyImpl) GetSymbol() string                             { panic("not implemented") }
func (impl *dummyImpl) GetDecimals() uint                             { panic("not implemented") }
func (impl *dummyImpl) TotalSupply() uint64                           { panic("not implemented") }
func (impl *dummyImpl) BalanceOf(account std.Address) (uint64, error) { panic("not implemented") }
func (impl *dummyImpl) Transfer(to std.Address, amount uint64) error  { panic("not implemented") }
func (impl *dummyImpl) Allowance(owner, spender std.Address) (uint64, error) {
	panic("not implemented")
}
func (impl *dummyImpl) Approve(spender std.Address, amount uint64) error { panic("not implemented") }
func (impl *dummyImpl) TransferFrom(from, to std.Address, amount uint64) error {
	panic("not implemented")
}
