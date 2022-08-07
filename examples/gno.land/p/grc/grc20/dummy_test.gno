package grc20

import (
	"std"
	"testing"
)

type dummyImpl struct{}

// FIXME: this should fail.
var _ IGRC20 = (*dummyImpl)(nil)

func TestInterface(t *testing.T) {
	var dummy IGRC20 = &dummyImpl{}
}

func (impl *dummyImpl) Name() string                                    { panic("not implemented") }
func (impl *dummyImpl) Symbol() string                                  { panic("not implemented") }
func (impl *dummyImpl) Decimals() uint                                  { panic("not implemented") }
func (impl *dummyImpl) TotalSupply() uint64                             { panic("not implemented") }
func (impl *dummyImpl) BalanceOf(account std.Address) uint64            { panic("not implemented") }
func (impl *dummyImpl) Transfer(to std.Address, amount uint64) bool     { panic("not implemented") }
func (impl *dummyImpl) Allowance(owner, spender std.Address) uint64     { panic("not implemented") }
func (impl *dummyImpl) Approve(spender std.Address, amount uint64) bool { panic("not implemented") }
func (impl *dummyImpl) TransferFrom(from, to std.Address, amount uint64) bool {
	panic("not implemented")
}
