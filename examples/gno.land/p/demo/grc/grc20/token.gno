package grc20

import (
	"math/overflow"
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
)

// NewToken creates a new Token.
// It returns a pointer to the Token and a pointer to the Ledger.
// Expected usage: Token, admin := NewToken("Dummy", "DUMMY", 4)
func NewToken(name, symbol string, decimals uint) (*Token, *PrivateLedger) {
	if name == "" {
		panic("name should not be empty")
	}
	if symbol == "" {
		panic("symbol should not be empty")
	}

	if len(name) > 15 {
		panic("name should not exceed 15 characters")
	}
	if len(symbol) > 10 {
		panic("symbol should not exceed 10 characters")
	}
	if !isValidName(name) {
		panic("name should be alphanumeric")
	}
	if !isValidSymbol(symbol) {
		panic("symbol should be uppercase alphabetic")
	}

	ledger := &PrivateLedger{}
	token := &Token{
		name:     name,
		symbol:   symbol,
		decimals: decimals,
		ledger:   ledger,
	}
	ledger.token = token
	return token, ledger
}

// GetName returns the name of the token.
func (tok Token) GetName() string { return tok.name }

// GetSymbol returns the symbol of the token.
func (tok Token) GetSymbol() string { return tok.symbol }

// GetDecimals returns the number of decimals used to get the token's precision.
func (tok Token) GetDecimals() uint { return tok.decimals }

// TotalSupply returns the total supply of the token.
func (tok Token) TotalSupply() uint64 { return tok.ledger.totalSupply }

// KnownAccounts returns the number of known accounts in the bank.
func (tok Token) KnownAccounts() int { return tok.ledger.balances.Size() }

// BalanceOf returns the balance of the specified address.
func (tok Token) BalanceOf(address std.Address) uint64 {
	return tok.ledger.balanceOf(address)
}

// Allowance returns the allowance of the specified owner and spender.
func (tok Token) Allowance(owner, spender std.Address) uint64 {
	return tok.ledger.allowance(owner, spender)
}

func (tok *Token) RenderHome() string {
	str := ""
	str += ufmt.Sprintf("# %s ($%s)\n\n", tok.name, tok.symbol)
	str += ufmt.Sprintf("* **Decimals**: %d\n", tok.decimals)
	str += ufmt.Sprintf("* **Total supply**: %d\n", tok.ledger.totalSupply)
	str += ufmt.Sprintf("* **Known accounts**: %d\n", tok.KnownAccounts())
	return str
}

// SpendAllowance decreases the allowance of the specified owner and spender.
func (led *PrivateLedger) SpendAllowance(owner, spender std.Address, amount uint64) error {
	if !owner.IsValid() {
		return ErrInvalidAddress
	}
	if !spender.IsValid() {
		return ErrInvalidAddress
	}

	currentAllowance := led.allowance(owner, spender)
	if currentAllowance < amount {
		return ErrInsufficientAllowance
	}

	key := allowanceKey(owner, spender)
	newAllowance := currentAllowance - amount

	if newAllowance == 0 {
		led.allowances.Remove(key)
	} else {
		led.allowances.Set(key, newAllowance)
	}

	return nil
}

// Transfer transfers tokens from the specified from address to the specified to address.
func (led *PrivateLedger) Transfer(from, to std.Address, amount uint64) error {
	if !from.IsValid() {
		return ErrInvalidAddress
	}
	if !to.IsValid() {
		return ErrInvalidAddress
	}
	if from == to {
		return ErrCannotTransferToSelf
	}

	var (
		toBalance   = led.balanceOf(to)
		fromBalance = led.balanceOf(from)
	)

	if fromBalance < amount {
		return ErrInsufficientBalance
	}

	var (
		newToBalance   = toBalance + amount
		newFromBalance = fromBalance - amount
	)

	led.balances.Set(string(to), newToBalance)
	led.balances.Set(string(from), newFromBalance)

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
func (led *PrivateLedger) TransferFrom(owner, spender, to std.Address, amount uint64) error {
	if led.balanceOf(owner) < amount {
		return ErrInsufficientBalance
	}
	if err := led.SpendAllowance(owner, spender, amount); err != nil {
		return err
	}
	// XXX: since we don't "panic", we should take care of rollbacking spendAllowance if transfer fails.
	return led.Transfer(owner, to, amount)
}

// Approve sets the allowance of the specified owner and spender.
func (led *PrivateLedger) Approve(owner, spender std.Address, amount uint64) error {
	if !owner.IsValid() || !spender.IsValid() {
		return ErrInvalidAddress
	}

	led.allowances.Set(allowanceKey(owner, spender), amount)

	std.Emit(
		ApprovalEvent,
		"owner", string(owner),
		"spender", string(spender),
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

// Mint increases the total supply of the token and adds the specified amount to the specified address.
func (led *PrivateLedger) Mint(address std.Address, amount uint64) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}

	// XXX: math/overflow is not supporting uint64.
	// This checks prevents overflow but makes the totalSupply limited to a uint63.
	sum, ok := overflow.Add64(int64(led.totalSupply), int64(amount))
	if !ok {
		return ErrOverflow
	}

	led.totalSupply = uint64(sum)
	currentBalance := led.balanceOf(address)
	newBalance := currentBalance + amount

	led.balances.Set(string(address), newBalance)

	std.Emit(
		TransferEvent,
		"from", "",
		"to", string(address),
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

// Burn decreases the total supply of the token and subtracts the specified amount from the specified address.
func (led *PrivateLedger) Burn(address std.Address, amount uint64) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}

	currentBalance := led.balanceOf(address)
	if currentBalance < amount {
		return ErrInsufficientBalance
	}

	led.totalSupply -= amount
	newBalance := currentBalance - amount

	led.balances.Set(string(address), newBalance)

	std.Emit(
		TransferEvent,
		"from", string(address),
		"to", "",
		"value", strconv.Itoa(int(amount)),
	)

	return nil
}

// balanceOf returns the balance of the specified address.
func (led PrivateLedger) balanceOf(address std.Address) uint64 {
	balance, found := led.balances.Get(address.String())
	if !found {
		return 0
	}
	return balance.(uint64)
}

// allowance returns the allowance of the specified owner and spender.
func (led PrivateLedger) allowance(owner, spender std.Address) uint64 {
	allowance, found := led.allowances.Get(allowanceKey(owner, spender))
	if !found {
		return 0
	}
	return allowance.(uint64)
}

// allowanceKey returns the key for the allowance of the specified owner and spender.
func allowanceKey(owner, spender std.Address) string {
	return owner.String() + ":" + spender.String()
}

// Helper function to validate name, strictly alphanumeric
func isValidName(name string) bool {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// Helper function to validate symbol, strictly uppercase alphabetic
func isValidSymbol(symbol string) bool {
	for _, c := range symbol {
		if !(c >= 'A' && c <= 'Z') {
			return false
		}
	}
	return true
}
