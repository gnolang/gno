package grc20

import (
	"std"

	"gno.land/p/demo/grc/exts"
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

type Token2 interface {
	exts.TokenMetadata

	// Returns the amount of tokens in existence.
	TotalSupply() uint64

	// Returns the amount of tokens owned by `account`.
	BalanceOf(account std.Address) uint64

	// Moves `amount` tokens from the caller's account to `to`.
	//
	// Returns an error if the operation failed.
	Transfer(to std.Address, amount uint64) error

	// Returns the remaining number of tokens that `spender` will be
	// allowed to spend on behalf of `owner` through {transferFrom}. This is
	// zero by default.
	//
	// This value changes when {approve} or {transferFrom} are called.
	Allowance(owner, spender std.Address) uint64

	// Sets `amount` as the allowance of `spender` over the caller's tokens.
	//
	// Returns an error if the operation failed.
	//
	// IMPORTANT: Beware that changing an allowance with this method brings the risk
	// that someone may use both the old and the new allowance by unfortunate
	// transaction ordering. One possible solution to mitigate this race
	// condition is to first reduce the spender's allowance to 0 and set the
	// desired value afterwards:
	// https://github.com/ethereum/EIPs/issues/20#issuecomment-263524729
	Approve(spender std.Address, amount uint64) error

	// Moves `amount` tokens from `from` to `to` using the
	// allowance mechanism. `amount` is then deducted from the caller's
	// allowance.
	//
	// Returns an error if the operation failed.
	TransferFrom(from, to std.Address, amount uint64) error
}
