package grc721

import (
	"chain"

	"gno.land/p/nt/avl/v0"
)

// metadataNFT represents an NFT with metadata extensions.
type metadataNFT struct {
	*BasicNFT
	extensions *avl.Tree // AVL tree for storing metadata extensions
}

// Ensure that metadataNFT implements the IGRC721MetadataOnchain interface.
var _ IGRC721MetadataOnchain = (*metadataNFT)(nil)

// NewNFTWithMetadata creates a new basic NFT with metadata extensions.
func NewNFTWithMetadata(_ int, rlm realm, name, symbol string) *metadataNFT {
	return &metadataNFT{
		BasicNFT:   NewBasicNFT(0, rlm, name, symbol),
		extensions: avl.NewTree(),
	}
}

// SetTokenMetadata sets metadata for a given token ID. The token must exist and
// caller must equal its owner; the owning realm's wrapper derives caller from
// rlm.Previous().Address() under IsCurrent (mirrors SetTokenRoyalty).
func (s *metadataNFT) SetTokenMetadata(caller address, tid TokenID, metadata Metadata) error {
	// Check that the token exists and the caller is its owner.
	owner, err := s.OwnerOf(tid)
	if err != nil {
		return err
	}
	if caller != owner {
		return ErrCallerIsNotOwner
	}

	// Set the metadata for the token ID in the extensions AVL tree
	s.extensions.Set(tid.String(), metadata)

	chain.Emit(
		MetadataUpdateEvent,
		"token", s.ID(),
		"tokenId", tid.String(),
	)

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

// Basic NFT methods forwarded to embedded BasicNFT

func (s *metadataNFT) Name() string {
	return s.BasicNFT.Name()
}

func (s *metadataNFT) Symbol() string {
	return s.BasicNFT.Symbol()
}

func (s *metadataNFT) TokenCount() int64 {
	return s.BasicNFT.TokenCount()
}

func (s *metadataNFT) BalanceOf(addr address) (int64, error) {
	return s.BasicNFT.BalanceOf(addr)
}

func (s *metadataNFT) OwnerOf(tid TokenID) (address, error) {
	return s.BasicNFT.OwnerOf(tid)
}

func (s *metadataNFT) TokenURI(tid TokenID) (string, error) {
	return s.BasicNFT.TokenURI(tid)
}

func (s *metadataNFT) SetTokenURI(caller address, tid TokenID, tURI TokenURI) (bool, error) {
	return s.BasicNFT.SetTokenURI(caller, tid, tURI)
}

func (s *metadataNFT) IsApprovedForAll(owner, operator address) bool {
	return s.BasicNFT.IsApprovedForAll(owner, operator)
}

func (s *metadataNFT) Approve(caller, to address, tid TokenID) error {
	return s.BasicNFT.Approve(caller, to, tid)
}

func (s *metadataNFT) GetApproved(tid TokenID) (address, error) {
	return s.BasicNFT.GetApproved(tid)
}

func (s *metadataNFT) SetApprovalForAll(caller, operator address, approved bool) error {
	return s.BasicNFT.SetApprovalForAll(caller, operator, approved)
}

func (s *metadataNFT) SafeTransferFrom(caller, from, to address, tid TokenID) error {
	return s.BasicNFT.SafeTransferFrom(caller, from, to, tid)
}

func (s *metadataNFT) TransferFrom(caller, from, to address, tid TokenID) error {
	return s.BasicNFT.TransferFrom(caller, from, to, tid)
}

func (s *metadataNFT) Mint(to address, tid TokenID) error {
	return s.BasicNFT.Mint(to, tid)
}

func (s *metadataNFT) SafeMint(to address, tid TokenID) error {
	return s.BasicNFT.SafeMint(to, tid)
}

func (s *metadataNFT) Burn(tid TokenID) error {
	return s.BasicNFT.Burn(tid)
}

func (s *metadataNFT) RenderHome() string {
	return s.BasicNFT.RenderHome()
}
