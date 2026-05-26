package grc721

import (
	"chain"
	"math/overflow"
	"strconv"

	"gno.land/p/nt/ufmt/v0"
)

// NewToken creates a Token whose origRealm is bound to the calling realm.
// rlm must be the caller's own captured cur (asserted via rlm.IsCurrent()),
// and rlm.PkgPath() — the calling realm itself — becomes the Token's
// origRealm. Token.ID() returns origRealm + "." + symbol.
//
// Because IsCurrent runtime-validates that rlm came from the live
// crossing frame, origRealm is unforgeable: an external realm cannot
// fabricate a Token claiming to belong to a different package.
//
// Typical call from a realm's init(cur realm) or other crossing function:
//
//	Token, ledger := grc721.NewToken(0, cur, "FooNFT", "FNFT")
//
// Do not export the returned *PrivateLedger.
func NewToken(_ int, rlm realm, name, symbol string) (*Token, *PrivateLedger) {
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

	ledger := &PrivateLedger{}
	token := &Token{
		name:      name,
		symbol:    symbol,
		origRealm: pkgPath,
		ledger:    ledger,
	}
	ledger.token = token

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

// validSymbol reports whether symbol is a valid token symbol: non-empty,
// within MaxSymbolLen, and consists only of [A-Za-z0-9_-].
func validSymbol(symbol string) bool {
	if symbol == "" || len(symbol) > MaxSymbolLen {
		return false
	}

	for _, c := range symbol {
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

// TotalSupply returns the total number of tokens in the collection.
func (tok Token) TotalSupply() int64 { return int64(tok.ledger.owners.Size()) }

// KnownAccounts returns the number of known accounts in the ledger.
func (tok Token) KnownAccounts() int { return tok.ledger.balances.Size() }

// ID returns the Identifier of the token.
// It is composed of the original realm and the provided symbol.
func (tok *Token) ID() string {
	return tok.origRealm + "." + tok.symbol
}

// BalanceOf returns the number of tokens owned by addr.
func (tok Token) BalanceOf(addr address) (int64, error) {
	if err := isValidAddress(addr); err != nil {
		return 0, err
	}

	bal, found := tok.ledger.balances.Get(addr.String())
	if !found {
		return 0, nil
	}

	return bal.(int64), nil
}

// OwnerOf returns the address that currently owns tid, or
// ErrInvalidTokenId if the token does not exist.
func (tok Token) OwnerOf(tid TokenID) (address, error) {
	owner, found := tok.ledger.owners.Get(tid.String())
	if !found {
		return "", ErrInvalidTokenId
	}

	return owner.(address), nil
}

// GetApproved returns the address approved to act on tid, or
// the empty address with ErrTokenIdNotHasApproved when no approval is set.
func (tok Token) GetApproved(tid TokenID) (address, error) {
	addr, found := tok.ledger.tokenApprovals.Get(tid.String())
	if !found {
		return "", ErrTokenIdNotHasApproved
	}

	return addr.(address), nil
}

// IsApprovedForAll reports whether op is an approved operator for all of
// owner's tokens.
func (tok Token) IsApprovedForAll(owner, op address) bool {
	key := owner.String() + ":" + op.String()
	v, found := tok.ledger.operatorApprovals.Get(key)
	if !found {
		return false
	}

	return v.(bool)
}

func (tok Token) RenderHome() string {
	return ufmt.Sprintf(
		"# %s ($%s)\n\n* **Total supply**: %d\n* **Known accounts**: %d\n",
		tok.name, tok.symbol, tok.TotalSupply(), tok.KnownAccounts(),
	)
}

// Mint creates token tid owned by to.
//
// Validity invariants:
//   - to is a valid address
//   - tid does not already exist
func (led *PrivateLedger) Mint(to address, tid TokenID) error {
	if err := isValidAddress(to); err != nil {
		return err
	}

	if led.exists(tid) {
		return ErrTokenIdAlreadyExists
	}

	bal, err := led.balanceOf(to)
	if err != nil {
		return err
	}

	bal = overflow.Add64p(bal, 1)

	toStr := to.String()
	tidStr := tid.String()
	led.balances.Set(toStr, bal)
	led.owners.Set(tidStr, to)

	chain.Emit(
		MintEvent,
		"token", led.token.ID(),
		"to", toStr,
		"tokenId", tidStr,
	)
	return nil
}

// SafeMint currently aliases Mint: there is no registry-based receiver
// check yet (see https://github.com/gnolang/gno/issues/5546).
func (led *PrivateLedger) SafeMint(to address, tid TokenID) error {
	return led.Mint(to, tid)
}

// Burn destroys token tid.
//
// Validity invariants:
//   - tid exists
func (led *PrivateLedger) Burn(tid TokenID) error {
	owner, err := led.token.OwnerOf(tid)
	if err != nil {
		return err
	}

	tidStr := tid.String()
	led.tokenApprovals.Remove(tidStr)

	bal, err := led.balanceOf(owner)
	if err != nil {
		return err
	}
	bal = overflow.Sub64p(bal, 1)

	ownerStr := owner.String()
	if bal == 0 {
		led.balances.Remove(ownerStr)
	} else {
		led.balances.Set(ownerStr, bal)
	}

	led.owners.Remove(tidStr)

	chain.Emit(
		BurnEvent,
		"token", led.token.ID(),
		"from", ownerStr,
		"tokenId", tidStr,
	)
	return nil
}

// Transfer moves token tid from `from` to `to` without authorization
// checks. Possessing *PrivateLedger IS the authorization; caller-resolved
// transfers belong on the Teller layer.
//
// Validity invariants:
//   - from and to are valid addresses
//   - from == OwnerOf(tid)
//   - from != to
func (led *PrivateLedger) Transfer(from, to address, tid TokenID) error {
	if err := isValidAddress(from); err != nil {
		return err
	}

	if err := isValidAddress(to); err != nil {
		return err
	}

	if from == to {
		return ErrCannotTransferToSelf
	}

	owner, err := led.token.OwnerOf(tid)
	if err != nil {
		return err
	}

	if owner != from {
		return ErrTransferFromIncorrectOwner
	}

	tidStr := tid.String()
	led.tokenApprovals.Remove(tidStr)

	fromBal, err := led.balanceOf(from)
	if err != nil {
		return err
	}

	toBal, err := led.balanceOf(to)
	if err != nil {
		return err
	}

	fromBal = overflow.Sub64p(fromBal, 1)
	toBal = overflow.Add64p(toBal, 1)

	fromStr := from.String()
	toStr := to.String()
	if fromBal == 0 {
		led.balances.Remove(fromStr)
	} else {
		led.balances.Set(fromStr, fromBal)
	}

	led.balances.Set(toStr, toBal)
	led.owners.Set(tidStr, to)

	chain.Emit(
		TransferEvent,
		"token", led.token.ID(),
		"from", fromStr,
		"to", toStr,
		"tokenId", tidStr,
	)
	return nil
}

// Approve sets `to` as the approved address for tid. The owner is derived
// from OwnerOf(tid); for NFTs no separate owner argument is needed.
//
// Validity invariants:
//   - tid exists
//   - to is a valid address
//   - to != OwnerOf(tid)
func (led *PrivateLedger) Approve(to address, tid TokenID) error {
	if err := isValidAddress(to); err != nil {
		return err
	}

	owner, err := led.token.OwnerOf(tid)
	if err != nil {
		return err
	}

	if owner == to {
		return ErrApprovalToCurrentOwner
	}

	tidStr := tid.String()
	led.tokenApprovals.Set(tidStr, to)

	chain.Emit(
		ApprovalEvent,
		"token", led.token.ID(),
		"owner", owner.String(),
		"to", to.String(),
		"tokenId", tidStr,
	)
	return nil
}

// SetApprovalForAll grants or revokes op as operator over all of owner's
// tokens.
//
// Validity invariants:
//   - owner and op are valid addresses
//   - owner != op
func (led *PrivateLedger) SetApprovalForAll(owner, op address, approved bool) error {
	if err := isValidAddress(owner); err != nil {
		return err
	}

	if err := isValidAddress(op); err != nil {
		return err
	}

	if owner == op {
		return ErrSelfApproval
	}

	ownerStr := owner.String()
	opStr := op.String()
	key := ownerStr + ":" + opStr
	if approved {
		led.operatorApprovals.Set(key, true)
	} else {
		led.operatorApprovals.Remove(key)
	}

	chain.Emit(
		ApprovalForAllEvent,
		"token", led.token.ID(),
		"owner", ownerStr,
		"to", opStr,
		"approved", strconv.FormatBool(approved),
	)
	return nil
}

func (led *PrivateLedger) exists(tid TokenID) bool {
	_, found := led.owners.Get(tid.String())

	return found
}

func (led *PrivateLedger) balanceOf(addr address) (int64, error) {
	if err := isValidAddress(addr); err != nil {
		return 0, err
	}

	v, found := led.balances.Get(addr.String())
	if !found {
		return 0, nil
	}

	return v.(int64), nil
}

func isValidAddress(addr address) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}

	return nil
}
