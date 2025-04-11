package grc777

import (
	"chain"
	"testing"
)

type dummyImpl struct{}

// FIXME: this should fail.
var _ IGRC777 = (*dummyImpl)(nil)

func TestInterface(t *testing.T) {
	var _ IGRC777 = &dummyImpl{}
}

func (impl *dummyImpl) GetName() string                          { panic("not implemented") }
func (impl *dummyImpl) GetSymbol() string                        { panic("not implemented") }
func (impl *dummyImpl) GetDecimals() uint                        { panic("not implemented") }
func (impl *dummyImpl) Granularity() (granularity uint64)        { panic("not implemented") }
func (impl *dummyImpl) TotalSupply() (supply uint64)             { panic("not implemented") }
func (impl *dummyImpl) BalanceOf(address chain.Address) uint64   { panic("not implemented") }
func (impl *dummyImpl) Burn(amount uint64, data []byte)          { panic("not implemented") }
func (impl *dummyImpl) AuthorizeOperator(operator chain.Address) { panic("not implemented") }
func (impl *dummyImpl) RevokeOperator(operators chain.Address)   { panic("not implemented") }
func (impl *dummyImpl) DefaultOperators() []chain.Address        { panic("not implemented") }
func (impl *dummyImpl) Send(recipient chain.Address, amount uint64, data []byte) {
	panic("not implemented")
}

func (impl *dummyImpl) IsOperatorFor(operator, tokenHolder chain.Address) bool {
	panic("not implemented")
}

func (impl *dummyImpl) OperatorSend(sender, recipient chain.Address, amount uint64, data, operatorData []byte) {
	panic("not implemented")
}

func (impl *dummyImpl) OperatorBurn(account chain.Address, amount uint64, data, operatorData []byte) {
	panic("not implemented")
}
