package grc777

import (
	"testing"
)

type dummyImpl struct{}

// FIXME: this should fail.
var _ IGRC777 = (*dummyImpl)(nil)

func TestInterface(t *testing.T) {
	var _ IGRC777 = &dummyImpl{}
}

func (impl *dummyImpl) GetName() string                     { panic("not implemented") }
func (impl *dummyImpl) GetSymbol() string                   { panic("not implemented") }
func (impl *dummyImpl) GetDecimals() int                    { panic("not implemented") }
func (impl *dummyImpl) Granularity() (granularity int64)    { panic("not implemented") }
func (impl *dummyImpl) TotalSupply() (supply int64)         { panic("not implemented") }
func (impl *dummyImpl) BalanceOf(address_XXX address) int64 { panic("not implemented") }
func (impl *dummyImpl) Burn(amount int64, data []byte)      { panic("not implemented") }
func (impl *dummyImpl) AuthorizeOperator(operator address)  { panic("not implemented") }
func (impl *dummyImpl) RevokeOperator(operators address)    { panic("not implemented") }
func (impl *dummyImpl) DefaultOperators() []address         { panic("not implemented") }
func (impl *dummyImpl) Send(recipient address, amount int64, data []byte) {
	panic("not implemented")
}

func (impl *dummyImpl) IsOperatorFor(operator, tokenHolder address) bool {
	panic("not implemented")
}

func (impl *dummyImpl) OperatorSend(sender, recipient address, amount int64, data, operatorData []byte) {
	panic("not implemented")
}

func (impl *dummyImpl) OperatorBurn(account address, amount int64, data, operatorData []byte) {
	panic("not implemented")
}
