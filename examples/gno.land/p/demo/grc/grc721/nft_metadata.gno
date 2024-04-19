package grc721

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
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