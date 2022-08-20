package grc777

import (
	"std"
	"testing"
)

type dummyImpl struct{}

// FIXME: this should fail.
var _ IGRC777 = (*dummyImpl)(nil)

func TestInterface(t *testing.T) {
	var dummy IGRC777 = &dummyImpl{}
}

func (impl *dummyImpl) Name() string                           { panic("not implemented") }
func (impl *dummyImpl) Symbol() string                         { panic("not implemented") }
func (impl *dummyImpl) Decimals() uint                         { panic("not implemented") }
func (impl *dummyImpl) Granularity() (granularity uint64)      { panic("not implemented") }
func (impl *dummyImpl) TotalSupply() (supply uint64)           { panic("not implemented") }
func (impl *dummyImpl) BalanceOf(address std.Address) uint64   { panic("not implemented") }
func (impl *dummyImpl) Burn(amount uint64, data []byte)        { panic("not implemented") }
func (impl *dummyImpl) AuthorizeOperator(operator std.Address) { panic("not implemented") }
func (impl *dummyImpl) RevokeOperator(operators std.Address)   { panic("not implemented") }
func (impl *dummyImpl) DefaultOperators() []std.Address        { panic("not implemented") }
func (impl *dummyImpl) Send(recipient std.Address, amount uint64, data []byte) {
	panic("not implemented")
}

func (impl *dummyImpl) IsOperatorFor(operator, tokenHolder std.Address) bool {
	panic("not implemented")
}

func (impl *dummyImpl) OperatorSend(sender, recipient std.Address, amount uint64, data, operatorData []byte) {
	panic("not implemented")
}

func (impl *dummyImpl) OperatorBurn(account std.Address, amount uint64, data, operatorData []byte) {
	panic("not implemented")
}
