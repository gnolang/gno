package grc721

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type BasicNFT struct {
	name              string
	symbol            string
	owners            avl.Tree // tokenId -> OwnerAddress
	balances          avl.Tree // OwnerAddress -> TokenCount
	tokenApprovals    avl.Tree // TokenId -> ApprovedAddress
	tokenURIs         avl.Tree // TokenId -> URIs
	operatorApprovals avl.Tree // "OwnerAddress:OperatorAddress" -> bool
}

// Returns new basic NFT
func NewBasicNFT(name string, symbol string) *BasicNFT {
	return &BasicNFT{
		name:   name,
		symbol: symbol,

		owners:            avl.Tree{},
		balances:          avl.Tree{},
		tokenApprovals:    avl.Tree{},
		tokenURIs:         avl.Tree{},
		operatorApprovals: avl.Tree{},
	}
}

func (s *BasicNFT) Name() string       { return s.name }
func (s *BasicNFT) Symbol() string     { return s.symbol }
func (s *BasicNFT) TokenCount() uint64 { return uint64(s.owners.Size()) }

// BalanceOf returns balance of input address
func (s *BasicNFT) BalanceOf(addr std.Address) (uint64, error) {
	if err := isValidAddress(addr); err != nil {
		return 0, err
	}

	balance, found := s.balances.Get(addr.String())
	if !found {
		return 0, nil
	}

	return balance.(uint64), nil
}

// OwnerOf returns owner of input token id
func (s *BasicNFT) OwnerOf(tid TokenID) (std.Address, error) {
	owner, found := s.owners.Get(string(tid))
	if !found {
		return "", ErrInvalidTokenId
	}

	return owner.(std.Address), nil
}

// TokenURI returns the URI of input token id
func (s *BasicNFT) TokenURI(tid TokenID) (string, error) {
	uri, found := s.tokenURIs.Get(string(tid))
	if !found {
		return "", ErrInvalidTokenId
	}

	return uri.(string), nil
}

// IsApprovedForAll returns true if operator is approved for all by the owner.
// Otherwise, returns false
func (s *BasicNFT) IsApprovedForAll(owner, operator std.Address) bool {
	key := owner.String() + ":" + operator.String()
	_, found := s.operatorApprovals.Get(key)
	if !found {
		return false
	}

	return true
}

// Approve approves the input address for particular token
func (s *BasicNFT) Approve(to std.Address, tid TokenID) error {
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

	caller := std.GetOrigCaller()
	if caller != owner && !s.IsApprovedForAll(owner, caller) {
		return ErrCallerIsNotOwnerOrApproved
	}

	s.tokenApprovals.Set(string(tid), to.String())
	event := ApprovalEvent{owner, to, tid}
	emit(&event)

	return nil
}

// GetApproved return the approved address for token
func (s *BasicNFT) GetApproved(tid TokenID) (std.Address, error) {
	addr, found := s.tokenApprovals.Get(string(tid))
	if !found {
		return zeroAddress, ErrTokenIdNotHasApproved
	}

	return std.Address(addr.(string)), nil
}

// SetApprovalForAll can approve the operator to operate on all tokens
func (s *BasicNFT) SetApprovalForAll(operator std.Address, approved bool) error {
	if err := isValidAddress(operator); err != nil {
		return ErrInvalidAddress
	}

	caller := std.GetOrigCaller()
	return s.setApprovalForAll(caller, operator, approved)
}

// Safely transfers `tokenId` token from `from` to `to`, checking that
// contract recipients are aware of the GRC721 protocol to prevent
// tokens from being forever locked.
func (s *BasicNFT) SafeTransferFrom(from, to std.Address, tid TokenID) error {
	caller := std.GetOrigCaller()
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
func (s *BasicNFT) TransferFrom(from, to std.Address, tid TokenID) error {
	caller := std.GetOrigCaller()
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
func (s *BasicNFT) Mint(to std.Address, tid TokenID) error {
	return s.mint(to, tid)
}

// Mints `tokenId` and transfers it to `to`. Also checks that
// contract recipients are using GRC721 protocol
func (s *BasicNFT) SafeMint(to std.Address, tid TokenID) error {
	err := s.mint(to, tid)
	if err != nil {
		return err
	}

	if !s.checkOnGRC721Received(zeroAddress, to, tid) {
		return ErrTransferToNonGRC721Receiver
	}

	return nil
}

func (s *BasicNFT) Burn(tid TokenID) error {
	owner, err := s.OwnerOf(tid)
	if err != nil {
		return err
	}

	s.beforeTokenTransfer(owner, zeroAddress, tid, 1)

	s.tokenApprovals.Remove(string(tid))
	balance, err := s.BalanceOf(owner)
	if err != nil {
		return err
	}
	balance -= 1
	s.balances.Set(owner.String(), balance)
	s.owners.Remove(string(tid))

	event := TransferEvent{owner, zeroAddress, tid}
	emit(&event)

	s.afterTokenTransfer(owner, zeroAddress, tid, 1)

	return nil
}

/* Helper methods */

// Helper for SetApprovalForAll()
func (s *BasicNFT) setApprovalForAll(owner, operator std.Address, approved bool) error {
	if owner == operator {
		return ErrApprovalToCurrentOwner
	}

	key := owner.String() + ":" + operator.String()
	s.operatorApprovals.Set(key, approved)

	event := ApprovalForAllEvent{owner, operator, approved}
	emit(&event)

	return nil
}

// Helper for TransferFrom() and SafeTransferFrom()
func (s *BasicNFT) transfer(from, to std.Address, tid TokenID) error {
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

	s.tokenApprovals.Remove(string(tid))
	fromBalance, err := s.BalanceOf(from)
	if err != nil {
		return err
	}
	toBalance, err := s.BalanceOf(to)
	if err != nil {
		return err
	}
	fromBalance -= 1
	toBalance += 1
	s.balances.Set(from.String(), fromBalance)
	s.balances.Set(to.String(), toBalance)
	s.owners.Set(string(tid), to)

	event := TransferEvent{from, to, tid}
	emit(&event)

	s.afterTokenTransfer(from, to, tid, 1)

	return nil
}

// Helper for Mint() and SafeMint()
func (s *BasicNFT) mint(to std.Address, tid TokenID) error {
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
	toBalance += 1
	s.balances.Set(to.String(), toBalance)
	s.owners.Set(string(tid), to)

	event := TransferEvent{zeroAddress, to, tid}
	emit(&event)

	s.afterTokenTransfer(zeroAddress, to, tid, 1)

	return nil
}

func (s *BasicNFT) isApprovedOrOwner(addr std.Address, tid TokenID) bool {
	owner, found := s.owners.Get(string(tid))
	if !found {
		return false
	}

	if addr == owner.(std.Address) || s.IsApprovedForAll(owner.(std.Address), addr) {
		return true
	}

	_, err := s.GetApproved(tid)
	if err != nil {
		return false
	}

	return true
}

// Checks if token id already exists
func (s *BasicNFT) exists(tid TokenID) bool {
	_, found := s.owners.Get(string(tid))
	return found
}

func (s *BasicNFT) beforeTokenTransfer(from, to std.Address, firstTokenId TokenID, batchSize uint64) {
	// TODO: Implementation
}

func (s *BasicNFT) afterTokenTransfer(from, to std.Address, firstTokenId TokenID, batchSize uint64) {
	// TODO: Implementation
}

func (s *BasicNFT) checkOnGRC721Received(from, to std.Address, tid TokenID) bool {
	// TODO: Implementation
	return true
}

func (s *BasicNFT) RenderHome() (str string) {
	str += ufmt.Sprintf("# %s ($%s)\n\n", s.name, s.symbol)
	str += ufmt.Sprintf("* **Total supply**: %d\n", s.TokenCount())
	str += ufmt.Sprintf("* **Known accounts**: %d\n", s.balances.Size())

	return
}
