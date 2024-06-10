package grc20

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ufmt"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"strconv"
)

type Token struct {
	name        string
	symbol      string
	totalSupply uint64
	decimals    uint8
	balances    *avl.Tree // std.Address > uint64 balance
	allowances  *avl.Tree // "OwnerAddress:SpenderAddress" -> uint64 allowance
}

var _ IGRC20 = (*Token)(nil)

const (
	emptyAddress  = std.Address("")
	TransferEvent = "Transfer"
	ApprovalEvent = "Approval"
)

func NewGRC20Token(name, symbol string, decimals uint8) *Token {
	return &Token{
		name:       name,
		symbol:     symbol,
		decimals:   decimals,
		balances:   avl.NewTree(),
		allowances: avl.NewTree(),
	}
}

func (t *Token) Name() (name string) {
	return t.name
}

func (t *Token) Symbol() (symbol string) {
	return t.symbol
}

func (t *Token) Decimals() (decimals uint8) {
	return t.decimals
}

func (t *Token) TotalSupply() (totalSupply uint64) {
	return t.totalSupply
}

func (t *Token) BalanceOf(owner std.Address) (balance uint64) {
	mustBeValid(owner)

	b, found := t.balances.Get(owner.String())
	if !found {
		return 0
	}

	return b.(uint64)
}

func (t *Token) Transfer(to std.Address, value uint64) (success bool) {
	//TODO implement me
	panic("implement me")
}

func (t *Token) TransferFrom(from, to std.Address, value uint64) (success bool) {
	//TODO implement me
	panic("implement me")
}

func (t *Token) Approve(spender std.Address, value uint64) (success bool) {
	//TODO implement me
	panic("implement me")
}

func (t *Token) Allowance(owner, spender std.Address) (remaining uint64) {
	//TODO implement me
	panic("implement me")
}

// Helpers

func (t *Token) update(from, to std.Address, value uint64) {
	// If new tokens are minted, check for overflow
	if from == emptyAddress {
		t.totalSupply += value // FIXME: actual overflow check?
	} else {
		// Deduct `value` from `from`
		rawFromBalance, found := t.balances.Get(from.String())
		if !found {
			err := ufmt.Sprintf("GRC20: Address %s not found", from)
			panic(err)
		}

		fromBalance := rawFromBalance.(uint64)

		if fromBalance < value {
			err := ufmt.Sprintf("GRC20: Insufficient balance %s, %s, %s", fromBalance, from, value)
			panic(err)
		}

		// Overflow not possible: value <= fromBalance <= totalSupply
		t.balances.Set(from.String(), fromBalance-value)
	}

	// Check if coins are burned
	if to == emptyAddress {
		// Overflow not possible: value <= totalSupply or value <= fromBalance <= totalSupply
		t.totalSupply -= value
	} else {
		rawToBalance, found := t.balances.Get(to.String())
		if !found {
			err := ufmt.Sprintf("GRC20: Address %s not found", to)
			panic(err)
		}

		toBalance := rawToBalance.(uint64)
		// Overflow not possible: balance + value is at most totalSupply, which we know fits into a uint64
		t.balances.Set(to.String(), toBalance+value)
	}

	std.Emit(TransferEvent,
		"from", from.String(),
		"to", to.String(),
		"value", strconv.Itoa(int(value)),
	)
}

func mustBeValid(address std.Address) {
	if !address.IsValid() {
		err := ufmt.Sprintf("GRC20: invalid address %s", address)
		panic(err)
	}
}

// spenderKey is a helper to create the key for the allowances tree
func spenderKey(owner, spender std.Address) string {
	return owner.String() + ":" + spender.String()
}
