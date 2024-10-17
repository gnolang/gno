package grc20

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
)

// NewBank creates a new Bank.
// It returns a pointer to the Bank and a pointer to the AdminBanker.
// Expected usage: Bank, admin := NewBank("Dummy", "DUMMY", 4)
func NewBank(name, symbol string, decimals uint) (*Bank, *AdminBanker) {
	if name == "" {
		panic("name should not be empty")
	}
	if symbol == "" {
		panic("symbol should not be empty")
	}
	// XXX additional checks (length, characters, limits, etc)

	bank := &Bank{
		name:     name,
		symbol:   symbol,
		decimals: decimals,
	}
	adm := &AdminBanker{}
	bank.adm = adm
	return bank, adm
}

// GetName returns the name of the token.
func (b Bank) GetName() string { return b.name }

// GetSymbol returns the symbol of the token.
func (b Bank) GetSymbol() string { return b.symbol }

// GetDecimals returns the number of decimals used to get the token's precision.
func (b Bank) GetDecimals() uint { return b.decimals }

// TotalSupply returns the total supply of the token.
func (b Bank) TotalSupply() uint64 { return b.adm.totalSupply }

// KnownAccounts returns the number of known accounts in the bank.
func (b Bank) KnownAccounts() int { return b.adm.balances.Size() }

// BalanceOf returns the balance of the specified address.
func (b Bank) BalanceOf(address std.Address) uint64 {
	return b.adm.balanceOf(address)
}

// Allowance returns the allowance of the specified owner and spender.
func (b Bank) Allowance(owner, spender std.Address) uint64 {
	return b.adm.allowance(owner, spender)
}

func (b *Bank) RenderHome() string {
	str := ""
	str += ufmt.Sprintf("# %s ($%s)\n\n", b.name, b.symbol)
	str += ufmt.Sprintf("* **Decimals**: %d\n", b.decimals)
	str += ufmt.Sprintf("* **Total supply**: %d\n", b.adm.totalSupply)
	str += ufmt.Sprintf("* **Known accounts**: %d\n", b.KnownAccounts())
	return str
}

// SpendAllowance decreases the allowance of the specified owner and spender.
func (b *AdminBanker) SpendAllowance(owner, spender std.Address, amount uint64) error {
	if !owner.IsValid() {
		return ErrInvalidAddress
	}
	if !spender.IsValid() {
		return ErrInvalidAddress
	}

	currentAllowance := b.allowance(owner, spender)
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

// Transfer transfers tokens from the specified from address to the specified to address.
func (b *AdminBanker) Transfer(from, to std.Address, amount uint64) error {
	if !from.IsValid() {
		return ErrInvalidAddress
	}
	if !to.IsValid() {
		return ErrInvalidAddress
	}
	if from == to {
		return ErrCannotTransferToSelf
	}

	toBalance := b.balanceOf(to)
	fromBalance := b.balanceOf(from)

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

// TransferFrom transfers tokens from the specified owner to the specified to address.
// It first checks if the owner has sufficient balance and then decreases the allowance.
func (b *AdminBanker) TransferFrom(owner, spender, to std.Address, amount uint64) error {
	if b.balanceOf(owner) < amount {
		return ErrInsufficientBalance
	}
	if err := b.SpendAllowance(owner, spender, amount); err != nil {
		return err
	}
	// XXX: since we don't "panic", we should take care of rollbacking spendAllowance if transfer fails.
	return b.Transfer(owner, to, amount)
}

// Approve sets the allowance of the specified owner and spender.
func (b *AdminBanker) Approve(owner, spender std.Address, amount uint64) error {
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

// Mint increases the total supply of the token and adds the specified amount to the specified address.
func (a *AdminBanker) Mint(address std.Address, amount uint64) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}

	// XXX check for overflow

	a.totalSupply += amount
	currentBalance := a.balanceOf(address)
	newBalance := currentBalance + amount

	a.balances.Set(string(address), newBalance)

	std.Emit(
		TransferEvent,
		"from", "",
		"to", string(address),
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

// Burn decreases the total supply of the token and subtracts the specified amount from the specified address.
func (a *AdminBanker) Burn(address std.Address, amount uint64) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}
	// XXX check for overflow

	currentBalance := a.balanceOf(address)
	if currentBalance < amount {
		return ErrInsufficientBalance
	}

	a.totalSupply -= amount
	newBalance := currentBalance - amount

	a.balances.Set(string(address), newBalance)

	std.Emit(
		TransferEvent,
		"from", string(address),
		"to", "",
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

// balanceOf returns the balance of the specified address.
func (b AdminBanker) balanceOf(address std.Address) uint64 {
	balance, found := b.balances.Get(address.String())
	if !found {
		return 0
	}
	return balance.(uint64)
}

// allowance returns the allowance of the specified owner and spender.
func (b AdminBanker) allowance(owner, spender std.Address) uint64 {
	allowance, found := b.allowances.Get(allowanceKey(owner, spender))
	if !found {
		return 0
	}
	return allowance.(uint64)
}

// allowanceKey returns the key for the allowance of the specified owner and spender.
func allowanceKey(owner, spender std.Address) string {
	return owner.String() + ":" + spender.String()
}
