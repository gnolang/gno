package foo20

import (
	"std"
	"strings"

	"gno.land/p/avl"
	"gno.land/p/grc/grc20"
	"gno.land/p/ufmt"
)

type Token struct {
	// TODO: use big.Int or a custom uint256 instead of uint64?

	grc20.GRC20 // implements the GRC20 interface

	name        string
	symbol      string
	decimals    uint
	totalSupply uint64

	balances   *avl.MutTree // std.Address(owner) -> uint64
	allowances *avl.MutTree // string(owner+":"+spender) -> uint64
}

func newToken(name, symbol string, decimals uint) *Token {
	// FIXME: check for limits

	return &Token{
		name:     name,
		symbol:   symbol,
		decimals: decimals,

		balances:   avl.NewMutTree(),
		allowances: avl.NewMutTree(),
	}
}

var foo *Token
var admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // TODO: helper to change admin

const zeroAddress = std.Address("")

func init() {
	foo = newToken("Foo", "FOO", 4)
	foo.mint("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", 1000000*10000) // @administrator (1M)
	foo.mint("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq", 10000*10000)   // @manfred (10k)
}

// GRC20 implementation.
//

// TODO: create a reusable interface with optional hooks.
// TODO: simplify the API and try to use events when available.
// TODO: useful Render() method.
// TODO: add a lot of unit tests, really a lot.

func (t *Token) TotalSupply() uint64 {
	return t.totalSupply
}

func (t *Token) BalanceOf(address std.Address) uint64 {
	return t.balanceOf(address)
}

func (t *Token) Transfer(to std.Address, amount uint64) {
	caller := std.GetCallerAt(3)
	t.transfer(caller, to, amount)
}

func (t *Token) Allowance(owner, spender std.Address) uint64 {
	return t.allowance(owner, spender)
}

func (t *Token) Approve(spender std.Address, amount uint64) {
	owner := std.GetCallerAt(3)
	t.approve(owner, spender, amount)
}

func (t *Token) TransferFrom(from, to std.Address, amount uint64) {
	spender := std.GetCallerAt(3)
	t.spendAllowance(from, spender, amount)
	t.transfer(from, to, amount)
}

// Administration helpers implementation.
//

func (t *Token) Mint(address std.Address, amount uint64) {
	caller := std.GetCallerAt(3)
	assertIsAdmin(caller)
	t.mint(address, amount)
}

func (t *Token) Burn(address std.Address, amount uint64) {
	caller := std.GetCallerAt(3)
	assertIsAdmin(caller)
	t.burn(address, amount)
}

// private helpers
//

func (t *Token) mint(address std.Address, amount uint64) {
	checkIsValidAddress(address)
	// TODO: check for overflow

	t.totalSupply += amount
	currentBalance := t.balanceOf(address)
	newBalance := currentBalance + amount

	t.balances.Set(string(address), newBalance)

	event := grc20.TransferEvent{zeroAddress, address, amount}
	emit(&event)
}

func (t *Token) burn(address std.Address, amount uint64) {
	checkIsValidAddress(address)
	// TODO: check for overflow

	currentBalance := t.balanceOf(address)
	if currentBalance < amount {
		panic("insufficient balance")
	}

	t.totalSupply -= amount
	newBalance := currentBalance - amount

	t.balances.Set(string(address), newBalance)

	event := grc20.TransferEvent{address, zeroAddress, amount}
	emit(&event)
}

func (t *Token) balanceOf(address std.Address) uint64 {
	checkIsValidAddress(address)

	balance, found := t.balances.Get(address.String())
	if !found {
		return 0
	}
	return balance.(uint64)
}

func (t *Token) spendAllowance(owner, spender std.Address, amount uint64) {
	checkIsValidAddress(owner)
	checkIsValidAddress(spender)

	currentAllowance := t.allowance(owner, spender)
	if currentAllowance < amount {
		panic("insufficient allowance")
	}
}

func (t *Token) transfer(from, to std.Address, amount uint64) {
	checkIsValidAddress(from)
	checkIsValidAddress(to)

	if from == to {
		panic("cannot send transfer to self")
	}

	toBalance := t.balanceOf(to)
	fromBalance := t.balanceOf(from)

	if fromBalance < amount {
		panic("insufficient balance")
	}

	newToBalance := toBalance + amount
	newFromBalance := fromBalance - amount

	t.balances.Set(string(to), newToBalance)
	t.balances.Set(string(from), newFromBalance)

	event := grc20.TransferEvent{from, to, amount}
	emit(&event)
}

func (t *Token) allowance(owner, spender std.Address) uint64 {
	checkIsValidAddress(owner)
	checkIsValidAddress(spender)

	key := owner.String() + ":" + spender.String()

	allowance, found := t.allowances.Get(key)
	if !found {
		return 0
	}

	return allowance.(uint64)
}

func (t *Token) approve(owner, spender std.Address, amount uint64) {
	checkIsValidAddress(owner)
	checkIsValidAddress(spender)

	key := owner.String() + ":" + spender.String()
	t.allowances.Set(key, amount)

	event := grc20.ApprovalEvent{owner, spender, amount}
	emit(&event)
}

func checkIsValidAddress(addr std.Address) {
	if addr.String() == "" {
		panic("invalid address")
	}
}

func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}

// method proxies as public functions.
//

// getters.

func TotalSupply() uint64                         { return foo.TotalSupply() }
func BalanceOf(address std.Address) uint64        { return foo.BalanceOf(address) }
func Allowance(owner, spender std.Address) uint64 { return foo.Allowance(owner, spender) }

// setters.

func Transfer(to std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	foo.Transfer(to, amount)
}

func Approve(spender std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	foo.Approve(spender, amount)
}

func TransferFrom(from, to std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	foo.TransferFrom(from, to, amount)
}

// administration.

func Mint(address std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	foo.Mint(address, amount)
}

func Burn(address std.Address, amount uint64) {
	//std.AssertOriginCall() // FIXME: inconsistent
	foo.Burn(address, amount)
}

// placeholders.
//

func emit(event interface{}) {
	// TODO: should we do something there?
	// noop
}

// render.
//

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		str := ""
		str += ufmt.Sprintf("# %s ($%s)\n\n", foo.name, foo.symbol)
		str += ufmt.Sprintf("* **Decimals**: %d\n", foo.decimals)
		str += ufmt.Sprintf("* **Total supply**: %d\n", foo.totalSupply)
		str += ufmt.Sprintf("* **Known accounts**: %d\n", foo.balances.Size())
		return str
	case c == 2 && parts[0] == "balance":
		addr := std.Address(parts[1])
		balance := foo.BalanceOf(addr)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}
