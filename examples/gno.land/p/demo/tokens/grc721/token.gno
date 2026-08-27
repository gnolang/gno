package grc721

import (
	"chain"
	"math/overflow"
	"strconv"

	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
)

// NewToken creates a core token and its ledger.
// rlm must be the caller's own captured cur (IsCurrent); its PkgPath becomes the unforgeable origRealm.
//
// Every successful call emits a NewToken event carrying the resulting Token.ID().
// Because Token's fields are unexported, NewToken is the only way a Token can come
// into existence, so this event makes token creation fully observable: an indexer
// that sees the same Token.ID() announced twice knows the realm built two
// independent ledgers behind one identifier, and that every later
// Mint/Burn/Transfer/Approval carrying that id is ambiguous. Such a realm is
// emitting untrustworthy events and should be flagged or ignored wholesale.
func NewToken(name, symbol string, id seqid.ID, rlm realm) (*Token, *PrivateLedger) {
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

	// origRealm is the host path, with any ":subpath" synthesized by realm.Sub
	// stripped. guardHome strips the invoking path the same way, so a token
	// created from a sub frame is not locked out of its own host realm, and a
	// host's sub-realms are not falsely rejected. Token.ID() keeps the raw
	// pkgPath, so identities and the NewToken event are unaffected.
	origRealm, _, _ := chain.SplitPkgSubPath(pkgPath)

	ledger := &PrivateLedger{}
	token := &Token{
		id:        pkgPath + "." + symbol + "." + id.String(),
		name:      name,
		symbol:    symbol,
		ledger:    ledger,
		origRealm: origRealm,
	}
	ledger.token = token

	chain.Emit(
		NewTokenEvent,
		"token", token.id,
		"name", name,
		"symbol", symbol,
	)

	return token, ledger
}

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

func (tok *Token) GetName() string { return tok.name }

func (tok *Token) GetSymbol() string { return tok.symbol }

func (tok *Token) ID() string { return tok.id }

func (tok *Token) TotalSupply() int64 { return tok.ledger.totalSupply }

func (tok *Token) KnownAccounts() int { return tok.ledger.balances.Size() }

// EIP-721: balanceOf throws for queries about the zero address,
// since NFTs assigned to it are considered invalid.
func (tok *Token) BalanceOf(addr address) (int64, error) {
	if !addr.IsValid() {
		return 0, ErrInvalidAddress
	}

	return tok.ledger.balanceOf(addr), nil
}

func (tok *Token) OwnerOf(tid TokenID) (address, error) { return tok.ledger.ownerOf(tid) }

// EIP-721: getApproved throws for a token that is not a valid NFT, so a missing
// token is ErrInvalidTokenId and a live but unapproved one is ErrTokenIdNotApproved.
func (tok *Token) GetApproved(tid TokenID) (address, error) {
	if !tok.ledger.exists(tid) {
		return zeroAddress, ErrInvalidTokenId
	}

	addr := tok.ledger.tokenApprovals.Get(tid.String())
	if addr == nil {
		return zeroAddress, ErrTokenIdNotApproved
	}

	return addr.(address), nil
}

func (tok *Token) IsApprovedForAll(owner, operator address) bool {
	return tok.ledger.isApprovedForAll(owner, operator)
}

func (tok *Token) RenderHome() string {
	str := ""
	str += ufmt.Sprintf("# %s ($%s)\n\n", tok.name, tok.symbol)
	str += ufmt.Sprintf("* **Total supply**: %d\n", tok.ledger.totalSupply)
	str += ufmt.Sprintf("* **Known accounts**: %d\n", tok.KnownAccounts())

	return str
}

// RegisterExtension attaches an issuer-supplied hook set. Issuer-only (holds the
// PrivateLedger), which keeps the surface from being attacker-attachable — but the
// attachment is still a privilege: per the Extension doc, an attached hook can veto
// transfers (a panic aborts the transaction before the Transfer event is emitted) and
// adds gas to every movement.
// One extension per kind; a second registration of an attached kind panics.
// view is the read-view the extension publishes, or nil when it has none.
func (led *PrivateLedger) RegisterExtension(ext Extension, view ExtensionView) {
	if ext == nil {
		panic("grc721: nil extension")
	}

	kind := ext.ExtensionKind()
	if led.HasExtension(kind) {
		panic("grc721: extension kind already registered: " + kind)
	}

	led.extensions = append(led.extensions, registeredExtension{kind: kind, hook: ext, view: view})
}

func (led *PrivateLedger) HasExtension(kind string) bool {
	_, ok := led.lookupExtension(kind)

	return ok
}

// ExtensionKinds returns the registered kinds, in registration order.
func (led *PrivateLedger) ExtensionKinds() []string {
	kinds := make([]string, 0, len(led.extensions))
	for _, e := range led.extensions {
		kinds = append(kinds, e.kind)
	}

	return kinds
}

// ExtensionView returns the read-view registered for kind, if it has one.
func (led *PrivateLedger) ExtensionView(kind string) (ExtensionView, bool) {
	e, ok := led.lookupExtension(kind)
	if !ok || e.view == nil {
		return nil, false
	}

	return e.view, true
}

func (led *PrivateLedger) lookupExtension(kind string) (registeredExtension, bool) {
	for _, e := range led.extensions {
		if e.kind == kind {
			return e, true
		}
	}

	return registeredExtension{}, false
}

func (led *PrivateLedger) ReadToken() *Token { return led.token }

func (led *PrivateLedger) Mint(to address, tid TokenID) error {
	if !to.IsValid() {
		return ErrInvalidAddress
	}

	if tid == "" {
		return ErrInvalidTokenId
	}

	if led.exists(tid) {
		return ErrTokenIdAlreadyExists
	}

	led.balances.Set(to.String(), overflow.Add64p(led.balanceOf(to), 1))
	led.owners.Set(tid.String(), to)
	led.totalSupply = overflow.Add64p(led.totalSupply, 1)

	for _, e := range led.extensions {
		e.hook.OnMint(to, tid)
	}

	// Mint emits Transfer from the empty address (EIP-721); indexers reconstruct ownership from Transfer alone.
	chain.Emit(
		TransferEvent,
		"token", led.token.id,
		"from", "",
		"to", to.String(),
		"tokenId", tid.String(),
	)

	return nil
}

// Fans OnBurn out to every extension so per-token state cannot resurface if tid is re-minted.
func (led *PrivateLedger) Burn(tid TokenID) error {
	owner, err := led.ownerOf(tid)
	if err != nil {
		return err
	}

	tidStr := tid.String()
	led.tokenApprovals.Remove(tidStr)
	led.setBalance(owner, overflow.Sub64p(led.balanceOf(owner), 1))
	led.owners.Remove(tidStr)
	led.totalSupply = overflow.Sub64p(led.totalSupply, 1)

	for _, e := range led.extensions {
		e.hook.OnBurn(tid)
	}

	// Burn emits Transfer to the empty address (EIP-721).
	chain.Emit(
		TransferEvent,
		"token", led.token.id,
		"from", owner.String(),
		"to", "",
		"tokenId", tidStr,
	)

	return nil
}

// spender must own or be approved for tid; callers derive it from a trusted per-frame identity.
func (led *PrivateLedger) TransferFrom(spender, from, to address, tid TokenID) error {
	if !led.isApprovedOrOwner(spender, tid) {
		return ErrCallerIsNotOwnerOrApproved
	}

	return led.transfer(from, to, tid)
}

// caller must own tid or be an approved operator of the owner.
// EIP-721: approving the zero address clears (revokes) any single-token approval.
func (led *PrivateLedger) Approve(caller, to address, tid TokenID) error {
	revoke := to == zeroAddress
	if !revoke && !to.IsValid() {
		return ErrInvalidAddress
	}

	owner, err := led.ownerOf(tid)
	if err != nil {
		return err
	}

	if owner == to {
		return ErrApprovalToCurrentOwner
	}

	if caller != owner && !led.isApprovedForAll(owner, caller) {
		return ErrCallerIsNotOwnerOrApproved
	}

	if revoke {
		led.tokenApprovals.Remove(tid.String())
	} else {
		led.tokenApprovals.Set(tid.String(), to)
	}

	chain.Emit(
		ApprovalEvent,
		"token", led.token.id,
		"owner", owner.String(),
		"to", to.String(), // empty for a revoke
		"tokenId", tid.String(),
	)

	return nil
}

func (led *PrivateLedger) SetApprovalForAll(owner, operator address, approved bool) error {
	if !owner.IsValid() || !operator.IsValid() {
		return ErrInvalidAddress
	}

	if owner == operator {
		return ErrApprovalToCurrentOwner
	}

	key := operatorKey(owner, operator)
	if approved {
		led.operatorApprovals.Set(key, approved)
	} else {
		led.operatorApprovals.Remove(key)
	}

	chain.Emit(
		ApprovalForAllEvent,
		"token", led.token.id,
		"owner", owner.String(),
		"to", operator.String(),
		"approved", strconv.FormatBool(approved),
	)

	return nil
}

func (led *PrivateLedger) transfer(from, to address, tid TokenID) error {
	if !from.IsValid() || !to.IsValid() {
		return ErrInvalidAddress
	}

	if from == to {
		return ErrCannotTransferToSelf
	}

	owner, err := led.ownerOf(tid)
	if err != nil {
		return err
	}

	if owner != from {
		return ErrTransferFromIncorrectOwner
	}

	tidStr := tid.String()
	led.tokenApprovals.Remove(tidStr)
	led.setBalance(from, overflow.Sub64p(led.balanceOf(from), 1))
	led.setBalance(to, overflow.Add64p(led.balanceOf(to), 1))
	led.owners.Set(tidStr, to)

	for _, e := range led.extensions {
		e.hook.OnTransfer(from, to, tid)
	}

	chain.Emit(
		TransferEvent,
		"token", led.token.id,
		"from", from.String(),
		"to", to.String(),
		"tokenId", tidStr,
	)

	return nil
}

func (led *PrivateLedger) ownerOf(tid TokenID) (address, error) {
	owner := led.owners.Get(tid.String())
	if owner == nil {
		return zeroAddress, ErrInvalidTokenId
	}

	return owner.(address), nil
}

func (led *PrivateLedger) balanceOf(addr address) int64 {
	balance := led.balances.Get(addr.String())
	if balance == nil {
		return 0
	}

	return balance.(int64)
}

func (led *PrivateLedger) setBalance(addr address, balance int64) {
	if balance == 0 {
		led.balances.Remove(addr.String())

		return
	}

	led.balances.Set(addr.String(), balance)
}

func (led *PrivateLedger) isApprovedForAll(owner, operator address) bool {
	approved := led.operatorApprovals.Get(operatorKey(owner, operator))
	if approved == nil {
		return false
	}

	return approved.(bool)
}

func (led *PrivateLedger) isApprovedOrOwner(addr address, tid TokenID) bool {
	owner := led.owners.Get(tid.String())
	if owner == nil {
		return false
	}

	ownerAddr := owner.(address)
	if addr == ownerAddr || led.isApprovedForAll(ownerAddr, addr) {
		return true
	}

	approved := led.tokenApprovals.Get(tid.String())

	return approved != nil && approved.(address) == addr
}

func (led *PrivateLedger) exists(tid TokenID) bool {
	return led.owners.Has(tid.String())
}

func operatorKey(owner, operator address) string {
	return owner.String() + ":" + operator.String()
}
