package grc20

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

type Token struct {
	name        string
	symbol      string
	totalSupply uint64
	balances    *avl.Tree // std.Address > uint64 balance
	allowances  *avl.Tree // "OwnerAddress:SpenderAddress" -> uint64 allowance
}

var _ IGRC20 = (*Token)(nil)

const emptyAddress = std.Address("")

func (t Token) Name() (name string) {
	//TODO implement me
	panic("implement me")
}

func (t Token) Symbol() (symbol string) {
	//TODO implement me
	panic("implement me")
}

func (t Token) Decimals() (decimals uint8) {
	//TODO implement me
	panic("implement me")
}

func (t Token) TotalSupply() (totalSupply uint64) {
	//TODO implement me
	panic("implement me")
}

func (t Token) BalanceOf(owner std.Address) (balance uint64) {
	//TODO implement me
	panic("implement me")
}

func (t Token) Transfer(to std.Address, value uint64) (success bool) {
	//TODO implement me
	panic("implement me")
}

func (t Token) TransferFrom(from, to std.Address, value uint64) (success bool) {
	//TODO implement me
	panic("implement me")
}

func (t Token) Approve(spender std.Address, value uint64) (success bool) {
	//TODO implement me
	panic("implement me")
}

func (t Token) Allowance(owner, spender std.Address) (remaining uint64) {
	//TODO implement me
	panic("implement me")
}
