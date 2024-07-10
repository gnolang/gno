package grc20

import (
	"std"
	"strconv"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

func NewBank(name, symbol string, decimals uint) *Bank {
	if name == "" {
		panic("name should not be empty")
	}
	if symbol == "" {
		panic("symbol should not be empty")
	}
	// XXX additional checks (length, characters, limits, etc)

	return &Bank{
		name:     name,
		symbol:   symbol,
		decimals: decimals,
		owner:    ownable.New(),
	}
}

func (b Bank) GetName() string     { return b.name }
func (b Bank) GetSymbol() string   { return b.symbol }
func (b Bank) GetDecimals() uint   { return b.decimals }
func (b Bank) TotalSupply() uint64 { return b.totalSupply_ }
func (b Bank) KnownAccounts() int  { return b.balances.Size() }

func (b *Bank) Mint(address std.Address, amount uint64) error {
	// panic("AdminBanker() can only be called by the Bank owner")

	if !address.IsValid() {
		return ErrInvalidAddress
	}

	// TODO: check for overflow

	b.totalSupply_ += amount
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

func (b *Bank) Burn(address std.Address, amount uint64) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}
	// TODO: check for overflow

	currentBalance := b.BalanceOf(address)
	if currentBalance < amount {
		return ErrInsufficientBalance
	}

	b.totalSupply_ -= amount
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

func (b Bank) BalanceOf(address std.Address) uint64 {
	balance, found := b.balances.Get(address.String())
	if !found {
		return 0
	}
	return balance.(uint64)
}

func (b *Bank) SpendAllowance(owner, spender std.Address, amount uint64) error {
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

func (b *Bank) Transfer(from, to std.Address, amount uint64) error {
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

func (b *Bank) TransferFrom(spender, from, to std.Address, amount uint64) error {
	if err := b.SpendAllowance(from, spender, amount); err != nil {
		return err
	}
	return b.Transfer(from, to, amount)
}

func (b *Bank) Allowance(owner, spender std.Address) uint64 {
	allowance, found := b.allowances.Get(allowanceKey(owner, spender))
	if !found {
		return 0
	}
	return allowance.(uint64)
}

func (b *Bank) Approve(owner, spender std.Address, amount uint64) error {
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
	str += ufmt.Sprintf("* **Total supply**: %d\n", b.totalSupply_)
	str += ufmt.Sprintf("* **Known accounts**: %d\n", b.KnownAccounts())
	return str
}

func allowanceKey(owner, spender std.Address) string {
	return owner.String() + ":" + spender.String()
}

func AccountSlugAddr(addr std.Address, slug string) std.Address {
	if slug == "" {
		return addr
	}
	key := addr.String() + "/" + slug
	return std.DerivePkgAddr(key) // temporarily using this helper
}
