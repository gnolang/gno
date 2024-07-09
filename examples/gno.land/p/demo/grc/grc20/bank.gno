package grc20

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type Bank struct {
	name        string
	symbol      string
	decimals    uint
	totalSupply uint64
	balances    avl.Tree // account -> uint64
	allowances  avl.Tree // owner.(account)+":"+spender.(account)) -> uint64
	token       *token   // to share the same pointer
	owner       std.Address
}

func NewBank(name, symbol string, decimals uint) *Bank {
	if name == "" {
		panic("name should not be empty")
	}
	if symbol == "" {
		panic("symbol should not be empty")
	}
	// XXX additional checks (length, characters, limits, etc)

	prev := std.PrevRealm().Addr()
	return &Bank{
		name:     name,
		symbol:   symbol,
		decimals: decimals,
		owner:    prev,
	}
}

func (b Bank) token() Token        { return b.token } // Token returns a grc20 safe-object implementation.
func (b Bank) getName() string     { return b.name }
func (b Bank) getSymbol() string   { return b.symbol }
func (b Bank) getDecimals() uint   { return b.decimals }
func (b Bank) totalSupply() uint64 { return b.totalSupply }
func (b Bank) knownAccounts() int  { return b.balances.Size() }

func (b *Bank) nint(address std.Address, amount uint64) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}

	// TODO: check for overflow

	b.totalSupply += amount
	currentBalance := b.BalanceOf(address)
	newBalance := currentBalance + amount

	b.balances.Set(string(address), newBalance)

	std.Emit(
		TransferEvent,
		"from", "",
		"to", string(address),
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

func (b *Bank) burn(address std.Address, amount uint64) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}
	// TODO: check for overflow

	currentBalance := b.BalanceOf(address)
	if currentBalance < amount {
		return ErrInsufficientBalance
	}

	b.totalSupply -= amount
	newBalance := currentBalance - amount

	b.balances.Set(string(address), newBalance)

	std.Emit(
		TransferEvent,
		"from", string(address),
		"to", "",
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

func (b Bank) balanceOf(address std.Address) uint64 {
	balance, found := b.balances.Get(address.String())
	if !found {
		return 0
	}
	return balance.(uint64)
}

func (b *Bank) spendAllowance(owner, spender std.Address, amount uint64) error {
	if !owner.IsValid() {
		return ErrInvalidAddress
	}
	if !spender.IsValid() {
		return ErrInvalidAddress
	}

	currentAllowance := b.Allowance(owner, spender)
	if currentAllowance < amount {
		return ErrInsufficientAllowance
	}

	key := allowanceKey(owner, spender)
	newAllowance := currentAllowance - amount

	if newAllowance == 0 {
		b.allowances.Remove(key)
	} else {
		b.allowances.Set(key, newAllowance)
	}

	return nil
}

func (b *Bank) transfer(from, to std.Address, amount uint64) error {
	if !from.IsValid() {
		return ErrInvalidAddress
	}
	if !to.IsValid() {
		return ErrInvalidAddress
	}
	if from == to {
		return ErrCannotTransferToSelf
	}

	toBalance := b.BalanceOf(to)
	fromBalance := b.BalanceOf(from)

	// debug.
	// println("from", from, "to", to, "amount", amount, "fromBalance", fromBalance, "toBalance", toBalance)

	if fromBalance < amount {
		return ErrInsufficientBalance
	}

	newToBalance := toBalance + amount
	newFromBalance := fromBalance - amount

	b.balances.Set(string(to), newToBalance)
	b.balances.Set(string(from), newFromBalance)

	std.Emit(
		TransferEvent,
		"from", from.String(),
		"to", to.String(),
		"value", strconv.Itoa(int(amount)),
	)
	return nil
}

func (b *Bank) transferFrom(spender, from, to std.Address, amount uint64) error {
	if err := b.SpendAllowance(from, spender, amount); err != nil {
		return err
	}
	return b.Transfer(from, to, amount)
}

func (b *Bank) allowance(owner, spender std.Address) uint64 {
	allowance, found := b.allowances.Get(allowanceKey(owner, spender))
	if !found {
		return 0
	}
	return allowance.(uint64)
}

func (b *Bank) approve(owner, spender std.Address, amount uint64) error {
	if !owner.IsValid() {
		return ErrInvalidAddress
	}
	if !spender.IsValid() {
		return ErrInvalidAddress
	}

	b.allowances.Set(allowanceKey(owner, spender), amount)

	std.Emit(
		ApprovalEvent,
		"owner", string(owner),
		"spender", string(spender),
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

func (b *Bank) RenderHome() string {
	str := ""
	str += ufmt.Sprintf("# %s ($%s)\n\n", b.name, b.symbol)
	str += ufmt.Sprintf("* **Decimals**: %d\n", b.decimals)
	str += ufmt.Sprintf("* **Total supply**: %d\n", b.totalSupply)
	str += ufmt.Sprintf("* **Known accounts**: %d\n", b.KnownAccounts())
	return str
}

func allowanceKey(owner, spender account) string {
	return owner + ":" + spender
}

type account string // <std.Address>[/slug]

func newAccount(addr std.Address, slug string) string {
	if strings.Contains(slug, ":") {
		panic("invalid account slug")
	}
	if slug == "" {
		return addr.String()
	}
	return addr.String() + "/" + slug
}
