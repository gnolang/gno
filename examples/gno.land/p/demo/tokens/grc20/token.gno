package grc20

import (
	"chain"
	"chain/runtime"
	"math"
	"math/overflow"
	"strconv"

	"gno.land/p/nt/ufmt/v0"
)

// NewToken creates a new Token.
// It returns a pointer to the Token and a pointer to the Ledger.
// Expected usage: Token, admin := NewToken("Dummy", "DUMMY", 4)
func NewToken(name, symbol string, decimals int) (*Token, *PrivateLedger) {
	if name == "" {
		panic("name should not be empty")
	}
	if symbol == "" {
		panic("symbol should not be empty")
	}
	// XXX additional checks (length, characters, limits, etc)

	ledger := &PrivateLedger{}
	token := &Token{
		name:      name,
		symbol:    symbol,
		decimals:  decimals,
		origRealm: runtime.CurrentRealm().PkgPath(),
		ledger:    ledger,
	}
	ledger.token = token
	return token, ledger
}

// GetName returns the name of the token.
func (tok Token) GetName() string { return tok.name }

// GetSymbol returns the symbol of the token.
func (tok Token) GetSymbol() string { return tok.symbol }

// GetDecimals returns the number of decimals used to get the token's precision.
func (tok Token) GetDecimals() int { return tok.decimals }

// TotalSupply returns the total supply of the token.
func (tok Token) TotalSupply() int64 { return tok.ledger.totalSupply }

// KnownAccounts returns the number of known accounts in the bank.
func (tok Token) KnownAccounts() int { return tok.ledger.balances.Size() }

// ID returns the Identifier of the token.
// It is composed of the original realm and the provided symbol.
func (tok *Token) ID() string {
	return tok.origRealm + "." + tok.symbol
}

// HasAddr checks if the specified address is a known account in the bank.
func (tok Token) HasAddr(addr address) bool {
	return tok.ledger.hasAddr(addr)
}

// BalanceOf returns the balance of the specified address.
func (tok Token) BalanceOf(addr address) int64 {
	return tok.ledger.balanceOf(addr)
}

// Allowance returns the allowance of the specified owner and spender.
func (tok Token) Allowance(owner, spender address) int64 {
	return tok.ledger.allowance(owner, spender)
}

func (tok Token) RenderHome() string {
	str := ""
	str += ufmt.Sprintf("# %s ($%s)\n\n", tok.name, tok.symbol)
	str += ufmt.Sprintf("* **Decimals**: %d\n", tok.decimals)
	str += ufmt.Sprintf("* **Total supply**: %d\n", tok.ledger.totalSupply)
	str += ufmt.Sprintf("* **Known accounts**: %d\n", tok.KnownAccounts())
	return str
}

// SpendAllowance decreases the allowance of the specified owner and spender.
func (led *PrivateLedger) SpendAllowance(owner, spender address, amount int64) error {
	if !owner.IsValid() || !spender.IsValid() {
		return ErrInvalidAddress
	}

	if amount < 0 {
		return ErrInvalidAmount
	}
	// do nothing
	if amount == 0 {
		return nil
	}

	currentAllowance := led.allowance(owner, spender)
	if currentAllowance < amount {
		return ErrInsufficientAllowance
	}

	key := allowanceKey(owner, spender)
	newAllowance := overflow.Sub64p(currentAllowance, amount)

	if newAllowance == 0 {
		led.allowances.Remove(key)
	} else {
		led.allowances.Set(key, newAllowance)
	}

	return nil
}

// Transfer transfers tokens from the specified from address to the specified to address.
func (led *PrivateLedger) Transfer(from, to address, amount int64) error {
	if !from.IsValid() {
		return ErrInvalidAddress
	}
	if !to.IsValid() {
		return ErrInvalidAddress
	}
	if from == to {
		return ErrCannotTransferToSelf
	}
	if amount < 0 {
		return ErrInvalidAmount
	}

	var (
		toBalance   = led.balanceOf(to)
		fromBalance = led.balanceOf(from)
	)

	if fromBalance < amount {
		return ErrInsufficientBalance
	}

	var (
		newToBalance   = overflow.Add64p(toBalance, amount)
		newFromBalance = overflow.Sub64p(fromBalance, amount)
	)

	led.balances.Set(string(to), newToBalance)

	if newFromBalance == 0 {
		led.balances.Remove(string(from))
	} else {
		led.balances.Set(string(from), newFromBalance)
	}

	chain.Emit(
		TransferEvent,
		"token", led.token.ID(),
		"from", from.String(),
		"to", to.String(),
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

// TransferFrom transfers tokens from the specified owner to the specified to address.
// It first checks if the owner has sufficient balance and then decreases the allowance.
func (led *PrivateLedger) TransferFrom(owner, spender, to address, amount int64) error {
	if amount < 0 {
		return ErrInvalidAmount
	}

	if !owner.IsValid() || !to.IsValid() {
		return ErrInvalidAddress
	}

	if led.balanceOf(owner) < amount {
		return ErrInsufficientBalance
	}

	// The check above guarantees that Transfer will succeed, ensuring
	// atomicity for the subsequent operations.
	if err := led.SpendAllowance(owner, spender, amount); err != nil {
		return err
	}

	if err := led.Transfer(owner, to, amount); err != nil {
		return err
	}

	return nil
}

// Approve sets the allowance of the specified owner and spender.
func (led *PrivateLedger) Approve(owner, spender address, amount int64) error {
	if !owner.IsValid() || !spender.IsValid() {
		return ErrInvalidAddress
	}
	if amount < 0 {
		return ErrInvalidAmount
	}

	led.allowances.Set(allowanceKey(owner, spender), amount)

	chain.Emit(
		ApprovalEvent,
		"token", led.token.ID(),
		"owner", string(owner),
		"spender", string(spender),
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

// Mint increases the total supply of the token and adds the specified amount to the specified address.
func (led *PrivateLedger) Mint(addr address, amount int64) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	if amount < 0 {
		return ErrInvalidAmount
	}

	// limit amount to MaxInt64 - totalSupply
	if amount > overflow.Sub64p(math.MaxInt64, led.totalSupply) {
		return ErrMintOverflow
	}

	led.totalSupply += amount
	currentBalance := led.balanceOf(addr)
	newBalance := overflow.Add64p(currentBalance, amount)

	led.balances.Set(string(addr), newBalance)

	chain.Emit(
		TransferEvent,
		"token", led.token.ID(),
		"from", "",
		"to", string(addr),
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

// Burn decreases the total supply of the token and subtracts the specified amount from the specified address.
func (led *PrivateLedger) Burn(addr address, amount int64) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	if amount < 0 {
		return ErrInvalidAmount
	}

	currentBalance := led.balanceOf(addr)
	if currentBalance < amount {
		return ErrInsufficientBalance
	}

	led.totalSupply = overflow.Sub64p(led.totalSupply, amount)
	newBalance := overflow.Sub64p(currentBalance, amount)

	if newBalance == 0 {
		led.balances.Remove(string(addr))
	} else {
		led.balances.Set(string(addr), newBalance)
	}

	chain.Emit(
		TransferEvent,
		"token", led.token.ID(),
		"from", string(addr),
		"to", "",
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

// hasAddr checks if the specified address is a known account in the ledger.
func (led PrivateLedger) hasAddr(addr address) bool {
	return led.balances.Has(addr.String())
}

// balanceOf returns the balance of the specified address.
func (led PrivateLedger) balanceOf(addr address) int64 {
	balance, found := led.balances.Get(addr.String())
	if !found {
		return 0
	}
	return balance.(int64)
}

// allowance returns the allowance of the specified owner and spender.
func (led PrivateLedger) allowance(owner, spender address) int64 {
	allowance, found := led.allowances.Get(allowanceKey(owner, spender))
	if !found {
		return 0
	}
	return allowance.(int64)
}

// allowanceKey returns the key for the allowance of the specified owner and spender.
func allowanceKey(owner, spender address) string {
	return owner.String() + ":" + spender.String()
}
