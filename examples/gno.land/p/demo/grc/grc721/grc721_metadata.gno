package grc721

import (
	"std"

	"gno.land/p/demo/avl"
)

// metadataNFT represents an NFT with metadata extensions.
type metadataNFT struct {
	*basicNFT            // Embedded basicNFT struct for basic NFT functionality
	extensions *avl.Tree // AVL tree for storing metadata extensions
}

// Ensure that metadataNFT implements the IGRC721MetadataOnchain interface.
var _ IGRC721MetadataOnchain = (*metadataNFT)(nil)

// NewNFTWithMetadata creates a new basic NFT with metadata extensions.
func NewNFTWithMetadata(name string, symbol string) *metadataNFT {
	// Create a new basic NFT
	nft := NewBasicNFT(name, symbol)

	// Return a metadataNFT with basicNFT embedded and an empty AVL tree for extensions
	return &metadataNFT{
		basicNFT:   nft,
		extensions: avl.NewTree(),
	}
}

// SetTokenMetadata sets metadata for a given token ID.
func (s *metadataNFT) SetTokenMetadata(tid TokenID, metadata Metadata) error {
	// Check if the caller is the owner of the token
	owner, err := s.basicNFT.OwnerOf(tid)
	if err != nil {
		return err
	}
	caller := std.PrevRealm().Addr()
	if caller != owner {
		return ErrCallerIsNotOwner
	}

	// Set the metadata for the token ID in the extensions AVL tree
	s.extensions.Set(string(tid), metadata)
	return nil
}

// TokenMetadata retrieves metadata for a given token ID.
func (s *metadataNFT) TokenMetadata(tid TokenID) (Metadata, error) {
	// Retrieve metadata from the extensions AVL tree
	metadata, found := s.extensions.Get(string(tid))
	if !found {
		return Metadata{}, ErrInvalidTokenId
	}

	return metadata.(Metadata), nil
}

// mint mints a new token and assigns it to the specified address.
func (s *metadataNFT) mint(to std.Address, tid TokenID) error {
	// Check if the address is valid
	if err := isValidAddress(to); err != nil {
		return err
	}

	// Check if the token ID already exists
	if s.basicNFT.exists(tid) {
		return ErrTokenIdAlreadyExists
	}

	s.basicNFT.beforeTokenTransfer(zeroAddress, to, tid, 1)

	// Check if the token ID was minted by beforeTokenTransfer
	if s.basicNFT.exists(tid) {
		return ErrTokenIdAlreadyExists
	}

	// Increment balance of the recipient address
	toBalance, err := s.basicNFT.BalanceOf(to)
	if err != nil {
		return err
	}
	toBalance += 1
	s.basicNFT.balances.Set(to.String(), toBalance)

	// Set owner of the token ID to the recipient address
	s.basicNFT.owners.Set(string(tid), to)

	// Emit transfer event
	event := TransferEvent{zeroAddress, to, tid}
	emit(&event)

	s.basicNFT.afterTokenTransfer(zeroAddress, to, tid, 1)

	return nil
}
