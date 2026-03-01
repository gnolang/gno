package grc721remake

import (
	"std"
	"strconv"
)

// NewNFT creates a new NFT.
// It returns a pointer to the NFT and a pointer to the Ledger.
// Expected usage: NFT, ledger := NewNFT("DummyNFT", "DNFT")
func NewNFT(name, symbol string) (*NFT, *PrivateLedger) {
	if name == "" {
		panic("name should not be empty")
	}
	if symbol == "" {
		panic("symbol should not be empty")
	}

	ledger := &PrivateLedger{
		hookRegistry: NewHookRegistry(),
	}
	nft := &NFT{
		name:   name,
		symbol: symbol,
		ledger: ledger,
	}
	ledger.nft = nft
	return nft, ledger
}

// GetName returns the name of the NFT.
func (n NFT) GetName() string { return n.name }

// GetSymbol returns the symbol of the NFT.
func (n NFT) GetSymbol() string { return n.symbol }

// TokenCount returns the total number of tokens.
func (n NFT) TokenCount() uint64 { return uint64(n.ledger.owners.Size()) }

// KnownAccounts returns the number of known accounts.
func (n NFT) KnownAccounts() int { return n.ledger.balances.Size() }

// BalanceOf returns the balance of the specified address.
func (n NFT) BalanceOf(address std.Address) (uint64, error) {
	return n.ledger.balanceOf(address)
}

// OwnerOf returns the owner of the specified token ID.
func (n NFT) OwnerOf(tid TokenID) (std.Address, error) {
	return n.ledger.ownerOf(tid)
}

// TokenURI returns the URI of the specified token ID.
func (n NFT) TokenURI(tid TokenID) (string, error) {
	return n.ledger.tokenURI(tid)
}

// GetApproved returns the approved address for the specified token ID.
func (n NFT) GetApproved(tid TokenID) (std.Address, error) {
	return n.ledger.getApproved(tid)
}

// IsApprovedForAll returns whether the specified operator is approved for all tokens of the owner.
func (n NFT) IsApprovedForAll(owner, operator std.Address) bool {
	return n.ledger.isApprovedForAll(owner, operator)
}

// Getter returns an NFTGetter function that returns this NFT.
func (n *NFT) Getter() NFTGetter {
	return func() *NFT {
		return n
	}
}

// RegisterHook registers a hook with a specific key (admin only)
func (led *PrivateLedger) RegisterHook(key string, hook Hook) error {
	led.hookRegistry.RegisterHook(key, hook)
	return nil
}

// GetHook retrieves a hook by its key
func (led *PrivateLedger) GetHook(key string) (Hook, error) {
	hook, found := led.hookRegistry.GetHook(key)
	if !found {
		return nil, ErrHookNotFound
	}
	return hook, nil
}

// SetTokenURI sets the URI for a token
func (led *PrivateLedger) SetTokenURI(caller std.Address, tid TokenID, tURI TokenURI) error {
	if !led.exists(tid) {
		return ErrInvalidTokenId
	}

	owner, err := led.ownerOf(tid)
	if err != nil {
		return err
	}

	if caller != owner {
		return ErrCallerIsNotOwner
	}

	led.tokenURIs.Set(string(tid), string(tURI))
	return nil
}

// Approve approves an address to transfer a specific token
func (led *PrivateLedger) Approve(caller std.Address, to std.Address, tid TokenID) error {
	if err := isValidAddress(to); err != nil {
		return err
	}

	owner, err := led.ownerOf(tid)
	if err != nil {
		return err
	}

	if caller != owner && !led.isApprovedForAll(owner, caller) {
		return ErrCallerIsNotOwnerOrApproved
	}

	if err := led.hookRegistry.TriggerApprovalHooks(led.nft, HookBefore, owner, to, tid); err != nil {
		return err
	}

	led.tokenApprovals.Set(string(tid), to.String())
	std.Emit(
		ApprovalEvent,
		"owner", string(owner),
		"to", string(to),
		"tokenId", string(tid),
	)

	return led.hookRegistry.TriggerApprovalHooks(led.nft, HookAfter, owner, to, tid)
}

// SetApprovalForAll approves or removes an operator to transfer all tokens
func (led *PrivateLedger) SetApprovalForAll(caller std.Address, operator std.Address, approved bool) error {
	if err := isValidAddress(operator); err != nil {
		return ErrInvalidAddress
	}

	if caller == operator {
		return ErrApprovalToCurrentOwner
	}

	if err := led.hookRegistry.TriggerApprovalForAllHooks(led.nft, HookBefore, caller, operator, approved); err != nil {
		return err
	}

	key := caller.String() + ":" + operator.String()
	led.operatorApprovals.Set(key, approved)

	std.Emit(
		ApprovalForAllEvent,
		"owner", string(caller),
		"operator", string(operator),
		"approved", strconv.FormatBool(approved),
	)

	return led.hookRegistry.TriggerApprovalForAllHooks(led.nft, HookAfter, caller, operator, approved)
}

// Transfer transfers a nft directly from the caller to another address
func (led *PrivateLedger) Transfer(from, to std.Address, tid TokenID) error {
	if err := isValidAddress(from); err != nil {
		return err
	}
	if err := isValidAddress(to); err != nil {
		return err
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

	if err := led.hookRegistry.TriggerTransferHooks(led.nft, HookBefore, from, to, tid); err != nil {
		return err
	}

	err = led.transferToken(from, to, tid)
	if err != nil {
		return err
	}

	return led.hookRegistry.TriggerTransferHooks(led.nft, HookAfter, from, to, tid)
}

// TransferFrom transfers a nft from one address to another
func (led *PrivateLedger) TransferFrom(caller, from, to std.Address, tid TokenID) error {
	if err := isValidAddress(from); err != nil {
		return err
	}
	if err := isValidAddress(to); err != nil {
		return err
	}

	if from == to {
		return ErrCannotTransferToSelf
	}

	if !led.isApprovedOrOwner(caller, tid) {
		return ErrCallerIsNotOwnerOrApproved
	}

	owner, err := led.ownerOf(tid)
	if err != nil {
		return err
	}
	if owner != from {
		return ErrTransferFromIncorrectOwner
	}

	if err := led.hookRegistry.TriggerTransferHooks(led.nft, HookBefore, from, to, tid); err != nil {
		return err
	}

	err = led.transferToken(from, to, tid)
	if err != nil {
		return err
	}

	return led.hookRegistry.TriggerTransferHooks(led.nft, HookAfter, from, to, tid)
}

// Mint creates a new token
func (led *PrivateLedger) Mint(to std.Address, tid TokenID) error {
	if err := isValidAddress(to); err != nil {
		return err
	}

	if led.exists(tid) {
		return ErrTokenIdAlreadyExists
	}

	if err := led.hookRegistry.TriggerMintHooks(led.nft, HookBefore, to, tid); err != nil {
		return err
	}

	err := led.mintToken(to, tid)
	if err != nil {
		return err
	}

	return led.hookRegistry.TriggerMintHooks(led.nft, HookAfter, to, tid)
}

func (led *PrivateLedger) Burn(tid TokenID) error {
	if err := led.hookRegistry.TriggerBurnHooks(led.nft, HookBefore, tid); err != nil {
		return err
	}

	_, err := led.burnToken(tid)
	if err != nil {
		return err
	}

	return led.hookRegistry.TriggerBurnHooks(led.nft, HookAfter, tid)
}

// BatchTransferFrom transfers multiple tokens from one address to another
func (led *PrivateLedger) BatchTransferFrom(caller, from, to std.Address, tids []TokenID) error {
	if len(tids) == 0 {
		return ErrEmptyTokenIDList
	}

	var err error
	for _, tid := range tids {
		err = led.TransferFrom(caller, from, to, tid)
		if err != nil {
			return err
		}
	}

	return nil
}

// BatchMint creates multiple new tokens
func (led *PrivateLedger) BatchMint(to std.Address, tids []TokenID) error {
	if len(tids) == 0 {
		return ErrEmptyTokenIDList
	}

	var err error
	for _, tid := range tids {
		err = led.Mint(to, tid)
		if err != nil {
			return err
		}
	}

	return nil
}

// BatchBurn destroys multiple tokens
func (led *PrivateLedger) BatchBurn(tids []TokenID) error {
	if len(tids) == 0 {
		return ErrEmptyTokenIDList
	}

	var err error
	for _, tid := range tids {
		err = led.Burn(tid)
		if err != nil {
			return err
		}
	}

	return nil
}

// BatchTransfer transfers multiple tokens to another address
func (led *PrivateLedger) BatchTransfer(caller std.Address, to std.Address, tids []TokenID) error {
	if len(tids) == 0 {
		return ErrEmptyTokenIDList
	}

	var err error

	for _, tid := range tids {
		err = led.Transfer(caller, to, tid)
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper functions

// balanceOf returns the balance of the specified address.
func (led PrivateLedger) balanceOf(address std.Address) (uint64, error) {
	if err := isValidAddress(address); err != nil {
		return 0, err
	}

	balance, found := led.balances.Get(address.String())
	if !found {
		return 0, nil
	}

	return balance.(uint64), nil
}

// ownerOf returns the owner of the specified token ID.
func (led PrivateLedger) ownerOf(tid TokenID) (std.Address, error) {
	owner, found := led.owners.Get(string(tid))
	if !found {
		return "", ErrInvalidTokenId
	}

	return owner.(std.Address), nil
}

// tokenURI returns the URI of the specified token ID.
func (led PrivateLedger) tokenURI(tid TokenID) (string, error) {
	uri, found := led.tokenURIs.Get(string(tid))
	if !found {
		return "", ErrInvalidTokenId
	}

	return uri.(string), nil
}

// getApproved returns the approved address for the specified token ID.
func (led PrivateLedger) getApproved(tid TokenID) (std.Address, error) {
	addr, found := led.tokenApprovals.Get(string(tid))
	if !found {
		return zeroAddress, ErrTokenIdNotHasApproved
	}

	return std.Address(addr.(string)), nil
}

// isApprovedForAll returns whether the specified operator is approved for all tokens of the owner.
func (led PrivateLedger) isApprovedForAll(owner, operator std.Address) bool {
	key := owner.String() + ":" + operator.String()
	_, found := led.operatorApprovals.Get(key)
	return found
}

// Private helper function to transfer a single token
func (led *PrivateLedger) transferToken(from, to std.Address, tid TokenID) error {
	led.tokenApprovals.Remove(string(tid))

	led.owners.Set(string(tid), to)

	fromBalance, _ := led.balanceOf(from)
	toBalance, _ := led.balanceOf(to)
	fromBalance -= 1
	toBalance += 1
	led.balances.Set(from.String(), fromBalance)
	led.balances.Set(to.String(), toBalance)

	std.Emit(
		TransferEvent,
		"from", string(from),
		"to", string(to),
		"tokenId", string(tid),
	)

	return nil
}

// Private helper function to mint a single token
func (led *PrivateLedger) mintToken(to std.Address, tid TokenID) error {
	led.owners.Set(string(tid), to)

	toBalance, _ := led.balanceOf(to)
	toBalance += 1
	led.balances.Set(to.String(), toBalance)

	std.Emit(
		MintEvent,
		"to", string(to),
		"tokenId", string(tid),
	)

	return nil
}

// Private helper function to burn a single token
func (led *PrivateLedger) burnToken(tid TokenID) (std.Address, error) {
	owner, err := led.ownerOf(tid)
	if err != nil {
		return "", err
	}

	led.tokenApprovals.Remove(string(tid))

	led.owners.Remove(string(tid))

	led.tokenURIs.Remove(string(tid))

	balance, _ := led.balanceOf(owner)
	balance -= 1
	led.balances.Set(owner.String(), balance)

	std.Emit(
		BurnEvent,
		"from", string(owner),
		"tokenId", string(tid),
	)

	return owner, nil
}

// exists checks if a nft exists
func (led *PrivateLedger) exists(tid TokenID) bool {
	_, found := led.owners.Get(string(tid))
	return found
}

// isApprovedOrOwner checks if an address is the owner or approved for a token
func (led *PrivateLedger) isApprovedOrOwner(addr std.Address, tid TokenID) bool {
	owner, err := led.ownerOf(tid)
	if err != nil {
		return false
	}

	if addr == owner || led.isApprovedForAll(owner, addr) {
		return true
	}

	approved, err := led.getApproved(tid)
	if err != nil {
		return false
	}

	return addr == approved
}

// Helper function to check if an address is valid
func isValidAddress(addr std.Address) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	return nil
}
