package grc721

import (
	"std"

	"gno.land/p/demo/avl"
)

// metadataNFT represents an NFT with metadata extensions.
type metadataNFT struct {
	*basicNFT

	extensions avl.Tree
}

// Ensure that metadataNFT implements the IGRC721Metadata interface.
var _ IGRC721Metadata = (*metadataNFT)(nil)

// NewNFTWithMetadata creates a new basic NFT with metadata extensions.
func NewNFTWithMetadata(name string, symbol string) *metadataNFT {

	nft := NewBasicNFT(name, symbol)

	return &metadataNFT{
		basicNFT:   nft,
		extensions: avl.Tree{},
	}
}

// SetExtension sets the metadata extension for the given token ID.
func (s *metadataNFT) SetExtension(tid TokenID, extension Metadata) (bool, error) {
	// Check for the right owner
	owner, err := s.basicNFT.OwnerOf(tid)
	if err != nil {
		return false, err
	}
	caller := std.PrevRealm().Addr()
	if caller != owner {
		return false, ErrCallerIsNotOwner
	}
	s.extensions.Set(string(tid), extension)
	return true, nil
}

// Extension retrieves the metadata extension for the given token ID.
func (s *metadataNFT) Extension(tid TokenID) (Metadata, error) {
	extension, found := s.extensions.Get(string(tid))
	if !found {
		return Metadata{}, ErrInvalidTokenId
	}

	return extension.(Metadata), nil
}

func (s *metadataNFT) mint(to std.Address, tid TokenID) error {
	if err := isValidAddress(to); err != nil {
		return err
	}

	if s.basicNFT.exists(tid) {
		return ErrTokenIdAlreadyExists
	}

	s.basicNFT.beforeTokenTransfer(zeroAddress, to, tid, 1)

	// Check that tokenId was not minted by `beforeTokenTransfer`
	if s.basicNFT.exists(tid) {
		return ErrTokenIdAlreadyExists
	}

	toBalance, err := s.basicNFT.BalanceOf(to)
	if err != nil {
		return err
	}
	toBalance += 1
	s.basicNFT.balances.Set(to.String(), toBalance)
	s.basicNFT.owners.Set(string(tid), to)

	event := TransferEvent{zeroAddress, to, tid}
	emit(&event)

	s.basicNFT.afterTokenTransfer(zeroAddress, to, tid, 1)

	return nil
}