// Package grc20 is a port of the Ethereum ERC20 standard for fungible tokens.
// It follows the interface defined in igrc20.gno, which is a by-spec implementation of the ERC20 standard.
// This package uses panics for reverting in specific cases, as it is defined by the specification.
// To use it, import it in your realm and instantiate a new top-level variable that will represent your token.
// Make sure that this variable is not exported, as it would otherwise be a security issue, since exported functions on
// exported top-level variables are callable.
// For example usage, check out r/demo/foo20 & r/demo/wugnot.

package grc20

import (
	"math"
	"math/bits"
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type Token struct {
	name        string
	symbol      string
	totalSupply uint64
	decimals    uint8
	balances    *avl.Tree // std.Address -> uint64 balance
	allowances  *avl.Tree // "OwnerAddress:SpenderAddress" -> uint64 allowance
}

// Make sure Token implements IGRC20 interface
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
	from := std.PrevRealm().Addr()
	t.transfer(from, to, value)

	return true
}

func (t *Token) TransferFrom(from, to std.Address, value uint64) (success bool) {
	spender := std.PrevRealm().Addr()
	t.spendAllowance(from, spender, value)
	t.transfer(from, to, value)

	return true
}

func (t *Token) Approve(spender std.Address, value uint64) (success bool) {
	caller := std.PrevRealm().Addr()
	t.approve(caller, spender, value, true)

	return true
}

func (t *Token) Allowance(owner, spender std.Address) (remaining uint64) {
	rawAllowance, found := t.allowances.Get(spenderKey(owner, spender))
	if !found {
		return 0
	}

	return rawAllowance.(uint64)
}

func (t *Token) Mint(account std.Address, amount uint64) {
	mustBeValid(account)
	t.update(emptyAddress, account, amount)
}

func (t *Token) Burn(account std.Address, amount uint64) {
	mustBeValid(account)
	t.update(account, emptyAddress, amount)
}

// Helpers
func (t *Token) update(from, to std.Address, value uint64) {
	// If new tokens are minted, check for overflow
	if from == emptyAddress {
		sum, carry := bits.Add64(t.totalSupply, value, 0)
		if carry != 0 {
			err := ufmt.Sprintf("GRC20($%s): uint64 overflow: %d + %d", t.symbol, t.totalSupply, value)
			panic(err)
		}

		t.totalSupply = sum
	} else {
		// Deduct `value` from `from`
		rawFromBalance, found := t.balances.Get(from.String())
		if !found {
			err := ufmt.Sprintf("GRC20($%s): Address %s not found", t.symbol, from)
			panic(err)
		}

		fromBalance := rawFromBalance.(uint64)

		if fromBalance < value {
			err := ufmt.Sprintf("GRC20($%s): Insufficient balance %s, %s, %s", t.symbol, fromBalance, from, value)
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
		// Valid transfer
		rawToBalance, found := t.balances.Get(to.String())
		if !found {
			// Initialize the balance the first time the address receives tokens
			t.balances.Set(to.String(), value)
		} else {
			toBalance := rawToBalance.(uint64)
			// Overflow not possible: balance + value is at most totalSupply, which we know fits into a uint64
			t.balances.Set(to.String(), toBalance+value)
		}
	}

	std.Emit(TransferEvent,
		"from", from.String(),
		"to", to.String(),
		"value", strconv.Itoa(int(value)),
	)
}

func (t *Token) transfer(from, to std.Address, value uint64) {
	mustBeValid(from)
	mustBeValid(to)
	t.update(from, to, value)
}

func (t *Token) approve(owner, spender std.Address, value uint64, emit bool) {
	mustBeValid(owner)
	mustBeValid(spender)

	t.allowances.Set(spenderKey(owner, spender), value)

	if emit {
		std.Emit(ApprovalEvent,
			"owner", owner.String(),
			"spender", spender.String(),
			"value", strconv.Itoa(int(value)),
		)
	}
}

// spendAllowance updates `owner`s allowance for `spender` based on spent `value`
// Does not update the allowance value in case of infinite allowance
// Panics if not enough allowance is available
// Does not emit an Approval event
func (t *Token) spendAllowance(owner, spender std.Address, value uint64) {
	currentAllowance := t.Allowance(owner, spender)

	if currentAllowance != math.MaxUint64 {
		if currentAllowance < value {
			err := ufmt.Sprintf("GRC20($%s): insufficient allowance: %s, %s, %d", t.symbol, spender.String(), currentAllowance, value)
			panic(err)
		}

		t.approve(owner, spender, currentAllowance-value, false)
	}
}

// Rendering

// RenderHome renders current values for the token instance
func (t *Token) RenderHome() string {
	str := ""
	str += ufmt.Sprintf("# %s ($%s)\n\n", t.name, t.symbol)
	str += ufmt.Sprintf("* **Decimals**: %d\n", t.decimals)
	str += ufmt.Sprintf("* **Total supply**: %d\n", t.totalSupply)
	str += ufmt.Sprintf("* **Known owners**: %d\n", t.balances.Size())

	return str
}

// Util
// mustBeValid checks if the given address is valid and panics otherwise
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
