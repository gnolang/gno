package grc20

import (
	"std"
)

// token implements the Token interface.
//
// It is generated with Banker.Token().
// It can safely be explosed publicly.
type token struct {
	banker *Banker
}

// var _ Token = (*token)(nil)
func (t *token) GetName() string     { return t.banker.name }
func (t *token) GetSymbol() string   { return t.banker.symbol }
func (t *token) GetDecimals() uint   { return t.banker.decimals }
func (t *token) TotalSupply() uint64 { return t.banker.totalSupply }

func (t *token) BalanceOf(owner std.Address) uint64 {
	return t.banker.BalanceOf(owner)
}

func (t *token) Transfer(to std.Address, amount uint64) error {
	caller := std.PrevRealm().Addr()
	return t.banker.Transfer(caller, to, amount)
}

func (t *token) Allowance(owner, spender std.Address) uint64 {
	return t.banker.Allowance(owner, spender)
}

func (t *token) Approve(spender std.Address, amount uint64) error {
	caller := std.PrevRealm().Addr()
	return t.banker.Approve(caller, spender, amount)
}

func (t *token) TransferFrom(from, to std.Address, amount uint64) error {
	spender := std.PrevRealm().Addr()
	if err := t.banker.SpendAllowance(from, spender, amount); err != nil {
		return err
	}
	return t.banker.Transfer(from, to, amount)
}
