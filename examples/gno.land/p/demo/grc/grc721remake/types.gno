package grc721remake

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
)

// NFTTeller interface defines the methods that a GRC721 token must implement.
// It extends the TokenMetadata interface to include methods for managing NFT
// ownership, transfers, approvals, and querying token information.
//
// The NFTTeller interface ensures that any NFT adhering to this standard
// provides a consistent API for interacting with non-fungible tokens.
type NFTTeller interface {
	// Returns the name of the NFT.
	GetName() string

	// Returns the symbol of the NFT, usually a shorter version of the
	// name.
	GetSymbol() string

	// Returns the total number of tokens in existence.
	TokenCount() uint64

	// Returns the number of tokens owned by the specified address.
	BalanceOf(address std.Address) (uint64, error)

	// Returns the owner of the specified token ID.
	OwnerOf(tid TokenID) (std.Address, error)

	// Returns the URI for a given token ID.
	TokenURI(tid TokenID) (string, error)

	// Sets the URI for a token
	SetTokenURI(tid TokenID, tURI TokenURI) error

	// Returns the approved address for a token ID, or an error if no approval exists.
	GetApproved(tid TokenID) (std.Address, error)

	// Returns whether an operator is approved for all tokens of an owner.
	IsApprovedForAll(owner, operator std.Address) bool

	// Approves another address to transfer a specific token.
	Approve(to std.Address, tid TokenID) error

	// Sets or unsets the approval of a given operator for all tokens owned by the caller.
	SetApprovalForAll(operator std.Address, approved bool) error

	// Transfers a token from the caller to another address.
	Transfer(to std.Address, tid TokenID) error

	// Transfers a token from one address to another.
	TransferFrom(from, to std.Address, tid TokenID) error

	// BatchTransferFrom transfers multiple tokens from one address to another.
	BatchTransferFrom(from, to std.Address, tids []TokenID) error

	// BatchTransfer transfers multiple tokens owned by the caller to another address.
	BatchTransfer(to std.Address, tids []TokenID) error
}

// TokenID represents a unique identifier for an NFT
type TokenID string

// TokenURI represents the URI for an NFT
type TokenURI string

// NFT represents a non-fungible token with a name and symbol.
// It maintains a ledger for tracking ownership, approvals, and token URIs.
//
// The NFT struct provides methods for retrieving token metadata and
// interacting with the ledger, including checking ownership and approvals.
type NFT struct {
	// Name of the NFT collection (e.g., "CryptoKitties").
	name string
	// Symbol of the NFT collection (e.g., "CK").
	symbol string
	// Pointer to the PrivateLedger that manages ownership and approvals.
	ledger *PrivateLedger
}

// NFTGetter is a function type that returns an NFT pointer. This type allows
// bypassing a limitation where we cannot directly pass NFT pointers between
// realms. Instead, we pass this function which can then be called to get the
// NFT pointer.
type NFTGetter func() *NFT

// PrivateLedger manages the state of the NFT collection, including ownership,
// balances, approvals, and token URIs. It provides administrative functions
// for minting, burning, transferring tokens, and managing approvals.
//
// The PrivateLedger is not safe to expose publicly, as it contains sensitive
// information and allows direct access to all administrative functions.
type PrivateLedger struct {
	// Pointer to the associated NFT struct
	nft *NFT
	// tokenId -> OwnerAddress
	owners avl.Tree
	// OwnerAddress -> TokenCount
	balances avl.Tree
	// TokenId -> ApprovedAddress
	tokenApprovals avl.Tree
	// TokenId -> URIs
	tokenURIs avl.Tree
	// "OwnerAddress:OperatorAddress" -> bool
	operatorApprovals avl.Tree
	// Hook registry for extensibility
	hookRegistry *HookRegistry
}

const (
	TransferEvent       = "Transfer"
	ApprovalEvent       = "Approval"
	ApprovalForAllEvent = "ApprovalForAll"
	MintEvent           = "Mint"
	BurnEvent           = "Burn"
)

var (
	ErrInvalidAddress              = errors.New("invalid address")
	ErrInvalidTokenId              = errors.New("invalid token id")
	ErrCallerIsNotOwner            = errors.New("caller is not owner")
	ErrApprovalToCurrentOwner      = errors.New("approval to current owner")
	ErrCallerIsNotOwnerOrApproved  = errors.New("caller is not owner or approved")
	ErrTokenIdNotHasApproved       = errors.New("token id not has approved")
	ErrTransferToNonGRC721Receiver = errors.New("transfer to non GRC721 receiver")
	ErrCannotTransferToSelf        = errors.New("cannot transfer to self")
	ErrTransferFromIncorrectOwner  = errors.New("transfer from incorrect owner")
	ErrTokenIdAlreadyExists        = errors.New("token id already exists")
	ErrEmptyTokenIDList            = errors.New("empty token ID list")
	ErrReadOnly                    = errors.New("teller is readonly")
	ErrHookNotFound                = errors.New("hook not found")
)

var zeroAddress std.Address

type fnNFTTeller struct {
	accountFn func() std.Address
	*NFT
}

// Ensure fnNFTTeller implements NFTTeller
var _ NFTTeller = (*fnNFTTeller)(nil)
