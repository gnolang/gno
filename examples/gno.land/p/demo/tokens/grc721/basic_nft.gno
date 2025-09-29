package grc721

import (
	"math/overflow"
	"std"
	"strconv"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

type basicNFT struct {
	name              string
	symbol            string
	owners            avl.Tree // tokenId -> OwnerAddress
	balances          avl.Tree // OwnerAddress -> TokenCount
	tokenApprovals    avl.Tree // TokenId -> ApprovedAddress
	tokenURIs         avl.Tree // TokenId -> URIs
	operatorApprovals avl.Tree // "OwnerAddress:OperatorAddress" -> bool
	origRealm         string   // Original realm path for ID
}

// Returns new basic NFT
func NewBasicNFT(name string, symbol string) *basicNFT {
	return &basicNFT{
		name:      name,
		symbol:    symbol,
		origRealm: std.CurrentRealm().PkgPath(),

		owners:            avl.Tree{},
		balances:          avl.Tree{},
		tokenApprovals:    avl.Tree{},
		tokenURIs:         avl.Tree{},
		operatorApprovals: avl.Tree{},
	}
}

func (s *basicNFT) Name() string      { return s.name }
func (s *basicNFT) Symbol() string    { return s.symbol }
func (s *basicNFT) TokenCount() int64 { return int64(s.owners.Size()) }

// ID returns the Identifier of the token collection.
// It is composed of the original realm and the provided symbol.
func (s *basicNFT) ID() string {
	return s.origRealm + "." + s.symbol
}

// BalanceOf returns balance of input address
func (s *basicNFT) BalanceOf(addr std.Address) (int64, error) {
	if err := isValidAddress(addr); err != nil {
		return 0, err
	}

	balance, found := s.balances.Get(addr.String())
	if !found {
		return 0, nil
	}

	return balance.(int64), nil
}

// OwnerOf returns owner of input token id
func (s *basicNFT) OwnerOf(tid TokenID) (std.Address, error) {
	owner, found := s.owners.Get(tid.String())
	if !found {
		return "", ErrInvalidTokenId
	}

	return owner.(std.Address), nil
}

// TokenURI returns the URI of input token id
func (s *basicNFT) TokenURI(tid TokenID) (string, error) {
	uri, found := s.tokenURIs.Get(tid.String())
	if !found {
		return "", ErrInvalidTokenId
	}

	return uri.(string), nil
}

func (s *basicNFT) SetTokenURI(tid TokenID, tURI TokenURI) (bool, error) {
	// check for invalid TokenID
	if !s.exists(tid) {
		return false, ErrInvalidTokenId
	}

	// check for the right owner
	owner, err := s.OwnerOf(tid)
	if err != nil {
		return false, err
	}
	caller := std.PreviousRealm().Address()
	if caller != owner {
		return false, ErrCallerIsNotOwner
	}
	s.tokenURIs.Set(tid.String(), tURI.String())
	return true, nil
}

// IsApprovedForAll returns true if operator is approved for all by the owner.
// Otherwise, returns false
func (s *basicNFT) IsApprovedForAll(owner, operator std.Address) bool {
	key := owner.String() + ":" + operator.String()
	approved, found := s.operatorApprovals.Get(key)
	if !found {
		return false
	}

	return approved.(bool)
}

// Approve approves the input address for particular token
func (s *basicNFT) Approve(to std.Address, tid TokenID) error {
	if err := isValidAddress(to); err != nil {
		return err
	}

	owner, err := s.OwnerOf(tid)
	if err != nil {
		return err
	}
	if owner == to {
		return ErrApprovalToCurrentOwner
	}

	caller := std.PreviousRealm().Address()
	if caller != owner && !s.IsApprovedForAll(owner, caller) {
		return ErrCallerIsNotOwnerOrApproved
	}

	tidStr := tid.String()
	toStr := to.String()
	s.tokenApprovals.Set(tidStr, toStr)

	std.Emit(
		ApprovalEvent,
		"token", s.ID(),
		"owner", owner.String(),
		"to", toStr,
		"tokenId", tidStr,
	)

	return nil
}

// GetApproved return the approved address for token
func (s *basicNFT) GetApproved(tid TokenID) (std.Address, error) {
	addr, found := s.tokenApprovals.Get(tid.String())
	if !found {
		return zeroAddress, ErrTokenIdNotHasApproved
	}

	addrStr, ok := addr.(string)
	if !ok {
		return zeroAddress, ErrInvalidAddress
	}

	return std.Address(addrStr), nil
}

// SetApprovalForAll can approve the operator to operate on all tokens
func (s *basicNFT) SetApprovalForAll(operator std.Address, approved bool) error {
	if err := isValidAddress(operator); err != nil {
		return ErrInvalidAddress
	}

	caller := std.PreviousRealm().Address()
	return s.setApprovalForAll(caller, operator, approved)
}

// Safely transfers `tokenId` token from `from` to `to`, checking that
// contract recipients are aware of the GRC721 protocol to prevent
// tokens from being forever locked.
func (s *basicNFT) SafeTransferFrom(from, to std.Address, tid TokenID) error {
	caller := std.PreviousRealm().Address()
	if !s.isApprovedOrOwner(caller, tid) {
		return ErrCallerIsNotOwnerOrApproved
	}

	err := s.transfer(from, to, tid)
	if err != nil {
		return err
	}

	if !s.checkOnGRC721Received(from, to, tid) {
		return ErrTransferToNonGRC721Receiver
	}

	return nil
}

// Transfers `tokenId` token from `from` to `to`.
func (s *basicNFT) TransferFrom(from, to std.Address, tid TokenID) error {
	caller := std.PreviousRealm().Address()
	if !s.isApprovedOrOwner(caller, tid) {
		return ErrCallerIsNotOwnerOrApproved
	}

	err := s.transfer(from, to, tid)
	if err != nil {
		return err
	}

	return nil
}

// Mints `tokenId` and transfers it to `to`.
func (s *basicNFT) Mint(to std.Address, tid TokenID) error {
	return s.mint(to, tid)
}

// Mints `tokenId` and transfers it to `to`. Also checks that
// contract recipients are using GRC721 protocol
func (s *basicNFT) SafeMint(to std.Address, tid TokenID) error {
	err := s.mint(to, tid)
	if err != nil {
		return err
	}

	if !s.checkOnGRC721Received(zeroAddress, to, tid) {
		return ErrTransferToNonGRC721Receiver
	}

	return nil
}

func (s *basicNFT) Burn(tid TokenID) error {
	owner, err := s.OwnerOf(tid)
	if err != nil {
		return err
	}

	s.beforeTokenTransfer(owner, zeroAddress, tid, 1)

	tidStr := tid.String()
	s.tokenApprovals.Remove(tidStr)
	balance, err := s.BalanceOf(owner)
	if err != nil {
		return err
	}
	balance = overflow.Sub64p(balance, 1)

	ownerStr := owner.String()
	s.balances.Set(ownerStr, balance)
	s.owners.Remove(tidStr)

	std.Emit(
		BurnEvent,
		"token", s.ID(),
		"from", ownerStr,
		"tokenId", tidStr,
	)

	s.afterTokenTransfer(owner, zeroAddress, tid, 1)

	return nil
}

/* Helper methods */

// Helper for SetApprovalForAll()
func (s *basicNFT) setApprovalForAll(owner, operator std.Address, approved bool) error {
	if owner == operator {
		return ErrApprovalToCurrentOwner
	}

	key := owner.String() + ":" + operator.String()
	s.operatorApprovals.Set(key, approved)

	std.Emit(
		ApprovalForAllEvent,
		"token", s.ID(),
		"owner", owner.String(),
		"to", operator.String(),
		"approved", strconv.FormatBool(approved),
	)

	return nil
}

// Helper for TransferFrom() and SafeTransferFrom()
func (s *basicNFT) transfer(from, to std.Address, tid TokenID) error {
	if err := isValidAddress(from); err != nil {
		return ErrInvalidAddress
	}
	if err := isValidAddress(to); err != nil {
		return ErrInvalidAddress
	}

	if from == to {
		return ErrCannotTransferToSelf
	}

	owner, err := s.OwnerOf(tid)
	if err != nil {
		return err
	}
	if owner != from {
		return ErrTransferFromIncorrectOwner
	}

	s.beforeTokenTransfer(from, to, tid, 1)

	// Check that tokenId was not transferred by `beforeTokenTransfer`
	owner, err = s.OwnerOf(tid)
	if err != nil {
		return err
	}
	if owner != from {
		return ErrTransferFromIncorrectOwner
	}

	tidStr := tid.String()
	s.tokenApprovals.Remove(tidStr)
	fromBalance, err := s.BalanceOf(from)
	if err != nil {
		return err
	}
	toBalance, err := s.BalanceOf(to)
	if err != nil {
		return err
	}
	fromBalance = overflow.Sub64p(fromBalance, 1)
	toBalance = overflow.Add64p(toBalance, 1)

	fromStr := from.String()
	toStr := to.String()

	s.balances.Set(fromStr, fromBalance)
	s.balances.Set(toStr, toBalance)
	s.owners.Set(tidStr, to)

	std.Emit(
		TransferEvent,
		"token", s.ID(),
		"from", fromStr,
		"to", toStr,
		"tokenId", tidStr,
	)

	s.afterTokenTransfer(from, to, tid, 1)

	return nil
}

// Helper for Mint() and SafeMint()
func (s *basicNFT) mint(to std.Address, tid TokenID) error {
	if err := isValidAddress(to); err != nil {
		return err
	}

	if s.exists(tid) {
		return ErrTokenIdAlreadyExists
	}

	s.beforeTokenTransfer(zeroAddress, to, tid, 1)

	// Check that tokenId was not minted by `beforeTokenTransfer`
	if s.exists(tid) {
		return ErrTokenIdAlreadyExists
	}

	toBalance, err := s.BalanceOf(to)
	if err != nil {
		return err
	}
	toBalance = overflow.Add64p(toBalance, 1)
	toStr := to.String()
	tidStr := tid.String()
	s.balances.Set(toStr, toBalance)
	s.owners.Set(tidStr, to)

	std.Emit(
		MintEvent,
		"token", s.ID(),
		"to", toStr,
		"tokenId", tidStr,
	)

	s.afterTokenTransfer(zeroAddress, to, tid, 1)

	return nil
}

func (s *basicNFT) isApprovedOrOwner(addr std.Address, tid TokenID) bool {
	owner, found := s.owners.Get(tid.String())
	if !found {
		return false
	}

	ownerAddr, ok := owner.(std.Address)
	if !ok {
		return false
	}
	if addr == ownerAddr || s.IsApprovedForAll(ownerAddr, addr) {
		return true
	}

	approved, err := s.GetApproved(tid)
	if err != nil {
		return false
	}

	return approved == addr
}

// Checks if token id already exists
func (s *basicNFT) exists(tid TokenID) bool {
	_, found := s.owners.Get(tid.String())
	return found
}

func (s *basicNFT) beforeTokenTransfer(from, to std.Address, firstTokenId TokenID, batchSize int64) {
	// TODO: Implementation
}

func (s *basicNFT) afterTokenTransfer(from, to std.Address, firstTokenId TokenID, batchSize int64) {
	// TODO: Implementation
}

func (s *basicNFT) checkOnGRC721Received(from, to std.Address, tid TokenID) bool {
	// TODO: Implementation
	return true
}

func (s *basicNFT) RenderHome() (str string) {
	str += ufmt.Sprintf("# %s ($%s)\n\n", s.name, s.symbol)
	str += ufmt.Sprintf("* **Total supply**: %d\n", s.TokenCount())
	str += ufmt.Sprintf("* **Known accounts**: %d\n", s.balances.Size())

	return
}

func (n *basicNFT) Getter() NFTGetter {
	return func() IGRC721 {
		return n
	}
}
