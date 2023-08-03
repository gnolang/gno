package grc20

import "std"

func Mustify(original IGRC20) IMustGRC20 {
	return &mustGRC20{original: original}
}

type mustGRC20 struct {
	original IGRC20
}

func (t *mustGRC20) GetName() string     { return t.original.GetName() }
func (t *mustGRC20) GetSymbol() string   { return t.original.GetSymbol() }
func (t *mustGRC20) GetDecimals() uint   { return t.original.GetDecimals() }
func (t *mustGRC20) TotalSupply() uint64 { return t.original.TotalSupply() }

func (t *mustGRC20) BalanceOf(owner std.Address) uint64 {
	balance, err := t.original.BalanceOf(owner)
	if err != nil {
		return 0
	}
	return balance
}

func (t *mustGRC20) Transfer(to std.Address, amount uint64) {
	err := t.original.Transfer(to, amount)
	if err != nil {
		panic(err)
	}
}

func (t *mustGRC20) Allowance(owner, spender std.Address) uint64 {
	allowance, err := t.original.Allowance(owner, spender)
	if err != nil {
		return 0
	}
	return allowance
}

func (t *mustGRC20) Approve(spender std.Address, amount uint64) {
	err := t.original.Approve(spender, amount)
	if err != nil {
		panic(err)
	}
}

func (t *mustGRC20) TransferFrom(from, to std.Address, amount uint64) {
	err := t.original.TransferFrom(from, to, amount)
	if err != nil {
		panic(err)
	}
}
