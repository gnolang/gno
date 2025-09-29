package grc721

import "gno.land/p/nt/avl"

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
	// Set the metadata for the token ID in the extensions AVL tree
	s.extensions.Set(tid.String(), metadata)
	return nil
}

// TokenMetadata retrieves metadata for a given token ID.
func (s *metadataNFT) TokenMetadata(tid TokenID) (Metadata, error) {
	// Retrieve metadata from the extensions AVL tree
	metadata, found := s.extensions.Get(tid.String())
	if !found {
		return Metadata{}, ErrInvalidTokenId
	}

	return metadata.(Metadata), nil
}

// Basic NFT methods forwarded to embedded basicNFT

func (s *metadataNFT) Name() string {
	return s.basicNFT.Name()
}

func (s *metadataNFT) Symbol() string {
	return s.basicNFT.Symbol()
}

func (s *metadataNFT) TokenCount() int64 {
	return s.basicNFT.TokenCount()
}

func (s *metadataNFT) BalanceOf(addr address) (int64, error) {
	return s.basicNFT.BalanceOf(addr)
}

func (s *metadataNFT) OwnerOf(tid TokenID) (address, error) {
	return s.basicNFT.OwnerOf(tid)
}

func (s *metadataNFT) TokenURI(tid TokenID) (string, error) {
	return s.basicNFT.TokenURI(tid)
}

func (s *metadataNFT) SetTokenURI(tid TokenID, tURI TokenURI) (bool, error) {
	return s.basicNFT.SetTokenURI(tid, tURI)
}

func (s *metadataNFT) IsApprovedForAll(owner, operator address) bool {
	return s.basicNFT.IsApprovedForAll(owner, operator)
}

func (s *metadataNFT) Approve(to address, tid TokenID) error {
	return s.basicNFT.Approve(to, tid)
}

func (s *metadataNFT) GetApproved(tid TokenID) (address, error) {
	return s.basicNFT.GetApproved(tid)
}

func (s *metadataNFT) SetApprovalForAll(operator address, approved bool) error {
	return s.basicNFT.SetApprovalForAll(operator, approved)
}

func (s *metadataNFT) SafeTransferFrom(from, to address, tid TokenID) error {
	return s.basicNFT.SafeTransferFrom(from, to, tid)
}

func (s *metadataNFT) TransferFrom(from, to address, tid TokenID) error {
	return s.basicNFT.TransferFrom(from, to, tid)
}

func (s *metadataNFT) Mint(to address, tid TokenID) error {
	return s.basicNFT.Mint(to, tid)
}

func (s *metadataNFT) SafeMint(to address, tid TokenID) error {
	return s.basicNFT.SafeMint(to, tid)
}

func (s *metadataNFT) Burn(tid TokenID) error {
	return s.basicNFT.Burn(tid)
}

func (s *metadataNFT) RenderHome() string {
	return s.basicNFT.RenderHome()
}
