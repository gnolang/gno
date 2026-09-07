package grc20

import (
	"chain"
	"chain/runtime"
	"math"
	"math/overflow"
	"strconv"

	"gno.land/p/nt/ufmt/v0"
)

// NewToken creates a Token whose origRealm is bound to the calling realm.
// rlm must be the caller's own captured cur (asserted via rlm.IsCurrent()),
// and rlm.PkgPath() — the calling realm itself — becomes the Token's
// origRealm.
//
// Because IsCurrent runtime-validates that rlm came from the live
// crossing frame, origRealm is unforgeable: an external realm cannot
// fabricate a Token claiming to belong to a different package.
//
// A Token has two names, and every event carries both. Token.TokenPath() is
// rlmPath.symbol, readable and unforgeable — it is built from rlm.PkgPath()
// under the assertion above — but two tokens one realm issues under one symbol
// share it. Token.ID() is the address derived from the token object's own VM
// object ID (preimage "objectid:<pkgid>:<newtime>"); the realm has no say in
// it, so it is what tells those two apart.
//
// An object ID is stamped the first time the object is persisted, which happens
// when a realm frame returns. So a fresh token has no ID yet: the NewToken event
// below carries an empty "id", as does any Mint/Transfer emitted before the
// creating call returns — unless the token is handed to another realm first
// (grc20reg.Register does exactly that), because that realm's finalize stamps it
// too. Once stamped the ID never changes. Binding an ID back to its realm is a
// lookup — TokenPath(), or the registry entry — since the address is a hash,
// not a path.
//
// If the Token should be discoverable, follow up with
// grc20reg.Register(cross(cur), Token, slug). That registry keys by
// rlmPath.symbol, which is exactly Token.TokenPath().
//
// Every successful call emits a NewToken event carrying both names. Because
// Token's fields are unexported, NewToken is the only way a Token can come into
// existence, so every ledger that will ever emit is announced here — subject to
// the window above: an announcement made before the token is persisted names
// the issuing realm and symbol, not yet the instance. Once the id is final,
// every Mint/Burn/Transfer/Approval carrying it refers to exactly one ledger.
func NewToken(name, symbol string, decimals int, rlm realm) (*Token, *PrivateLedger) {
	if !rlm.IsCurrent() {
		panic(ErrSpoofedRealm)
	}
	pkgPath := rlm.PkgPath()
	if pkgPath == "" {
		panic(ErrNotRealm)
	}
	if !validName(name) {
		panic(ErrInvalidName)
	}
	if !validSymbol(symbol) {
		panic(ErrInvalidSymbol)
	}
	if decimals < 0 || decimals > MaxDecimals {
		panic(ErrInvalidDecimals)
	}
	// origRealm drops any realm.Sub subpath: a token created while its realm
	// operates under a sub identity still belongs to the host realm. guardHome
	// resolves the invoking host the same way, so storing the raw path here
	// would pin the token to that sub and lock its own realm out for good.
	origRealm, _, _ := chain.SplitPkgSubPath(pkgPath)
	ledger := &PrivateLedger{}
	token := &Token{
		tokenPath: pkgPath + "." + symbol,
		name:      name,
		symbol:    symbol,
		decimals:  decimals,
		ledger:    ledger,
		origRealm: origRealm,
	}
	ledger.token = token

	chain.Emit(
		NewTokenEvent,
		"id", token.ID(),
		"token", token.tokenPath,
		"name", name,
		"symbol", symbol,
		"decimals", strconv.Itoa(decimals),
	)

	return token, ledger
}

// validName reports whether name is a valid display name: non-empty,
// within MaxNameLen, and contains no control characters (any rune
// below 0x20 or 0x7f). Permits Unicode letters, digits, punctuation,
// and spaces — name is purely a display field.
func validName(name string) bool {
	if name == "" || len(name) > MaxNameLen {
		return false
	}
	for _, c := range name {
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	return true
}

// validSymbol reports whether s is valid slug-compatible metadata: non-empty,
// within MaxSymbolLen, and consists only of [A-Za-z0-9_-].
func validSymbol(s string) bool {
	if s == "" || len(s) > MaxSymbolLen {
		return false
	}
	for _, c := range s {
		if !isAlnum(c) && c != '_' && c != '-' {
			return false
		}
	}
	return true
}

func isAlnum(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
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

// ID returns the address derived from the token object's own VM object ID. It
// identifies this ledger and no other, but it is "" until the token has been
// persisted — see NewToken.
func (tok *Token) ID() string {
	return runtime.ObjectID(tok)
}

// TokenPath returns the token path of the token.
// It is the full path of the token, including the realm and symbol, and is the
// key grc20reg registers it under. Unlike ID() it is available immediately, and
// it is shared by every token one realm issues under the same symbol.
func (tok *Token) TokenPath() string {
	return tok.tokenPath
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
		"id", led.token.ID(),
		"token", led.token.TokenPath(),
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
		"id", led.token.ID(),
		"token", led.token.TokenPath(),
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
		"id", led.token.ID(),
		"token", led.token.TokenPath(),
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
		"id", led.token.ID(),
		"token", led.token.TokenPath(),
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
	balance := led.balances.Get(addr.String())
	if balance == nil {
		return 0
	}
	return balance.(int64)
}

// allowance returns the allowance of the specified owner and spender.
func (led PrivateLedger) allowance(owner, spender address) int64 {
	allowance := led.allowances.Get(allowanceKey(owner, spender))
	if allowance == nil {
		return 0
	}
	return allowance.(int64)
}

// allowanceKey returns the key for the allowance of the specified owner and spender.
func allowanceKey(owner, spender address) string {
	return owner.String() + ":" + spender.String()
}
