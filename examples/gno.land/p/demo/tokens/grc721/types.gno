package grc721

import (
	"errors"

	"gno.land/p/nt/bptree/v0"
)

type (
	TokenID  string
	TokenURI string
)

func (t TokenID) String() string  { return string(t) }
func (t TokenURI) String() string { return string(t) }

// Teller interface defines the methods that a GRC721 token must implement.
// It exposes the read surface mirrored on *Token together with caller-resolved
// write methods.
//
// The Teller interface is designed to ensure that any token adhering to this
// standard provides a consistent API for interacting with non-fungible tokens.
//
// Per EIP-721, name/symbol/tokenURI all belong to the optional
// ERC721Metadata extension; tokenURI in particular is opted out of the
// core surface here and lives on the metadata subpackage. Core only
// exposes ownership, approvals, and transfers.
//
// SECURITY: Approve/SetApprovalForAll/TransferFrom take (_ int, rlm realm,
// ...), so handing a Teller value to untrusted code yields a capability
// token to whatever impl that code dispatches into. Any /p/ or /r/ function
// that accepts a Teller as a parameter from external callers MUST
// type-assert against the canonical concrete type (*fnTeller) via
// IsCanonicalTeller and reject otherwise. An unexported-marker "seal" does
// NOT defend against this — see
// p/test/seal/filetests/z_seal_iface_embedding_filetest.gno
// for the realistic embedding-bypass attack.
type Teller interface {
	// Returns the name of the token.
	GetName() string

	// Returns the symbol of the token.
	GetSymbol() string

	// Returns the total number of tokens in existence.
	TotalSupply() int64

	// Returns the number of tokens owned by `account`.
	BalanceOf(account address) (int64, error)

	// Returns the owner of the token identified by `tid`.
	OwnerOf(tid TokenID) (address, error)

	// Returns the address approved to act on the token identified by `tid`.
	GetApproved(tid TokenID) (address, error)

	// Reports whether `op` is an approved operator for all of `owner`'s
	// tokens.
	IsApprovedForAll(owner, op address) bool

	// Moves token `tid` from `from` to `to`. The caller must be the owner
	// of `tid`, an approved operator for the owner, or the address
	// approved for `tid`.
	//
	// rlm must be the caller's own captured cur — verified via
	// rlm.IsCurrent().
	//
	// Returns an error if the operation failed.
	//
	// NOTE: EIP-721's safeTransferFrom is not exposed: it would require a
	// registry-based ERC721TokenReceiver check (see
	// https://github.com/gnolang/gno/issues/5546). Until that lands,
	// implementing it as a TransferFrom alias would have leaked a false
	// safety guarantee, so it is intentionally absent.
	TransferFrom(_ int, rlm realm, from, to address, tid TokenID) error

	// Sets `to` as the approved address for token `tid`.
	//
	// Returns an error if the operation failed.
	Approve(_ int, rlm realm, to address, tid TokenID) error

	// Grants or revokes `op` as an operator over all of the caller's tokens.
	//
	// Returns an error if the operation failed.
	SetApprovalForAll(_ int, rlm realm, op address, approved bool) error
}

// Token represents a non-fungible token collection with a name and symbol.
// It maintains a ledger for tracking owners, balances, approvals, and token
// URIs.
//
// The Token struct provides methods for retrieving collection metadata, such
// as the name and symbol, as well as methods for interacting with the ledger,
// including querying ownership, balances, and approvals.
type Token struct {
	// Name of the token (e.g., "Dummy NFT").
	name string
	// Symbol of the token (e.g., "DUMMY").
	symbol string
	// Original realm of the token (e.g., "gno.land/r/demo/foo721").
	origRealm string
	// Pointer to the PrivateLedger that manages ownership and approvals.
	ledger *PrivateLedger
}

// PrivateLedger is a struct that holds the ownership and approvals for the
// collection. It provides administrative functions for minting, burning,
// transferring tokens, and managing approvals.
//
// The PrivateLedger is not safe to expose publicly, as it contains sensitive
// information regarding token ownership and approvals, and allows direct,
// unrestricted access to all administrative functions.
type PrivateLedger struct {
	// TokenID -> address
	owners bptree.BPTree
	// address -> int64
	balances bptree.BPTree
	// TokenID -> address
	tokenApprovals bptree.BPTree
	// owner.(address)+":"+op.(address) -> bool
	operatorApprovals bptree.BPTree
	// Pointer to the associated Token struct.
	token *Token
}

var (
	ErrInvalidTokenId             = errors.New("invalid token id")
	ErrInvalidAddress             = errors.New("invalid address")
	ErrTokenIdNotHasApproved      = errors.New("token id not approved for anyone")
	ErrApprovalToCurrentOwner     = errors.New("approval to current owner")
	ErrSelfApproval               = errors.New("cannot approve self as operator")
	ErrCallerIsNotOwner           = errors.New("caller is not token owner")
	ErrCannotTransferToSelf       = errors.New("cannot send transfer to self")
	ErrTransferFromIncorrectOwner = errors.New("transfer from incorrect owner")
	ErrCallerIsNotOwnerOrApproved = errors.New("caller is not token owner or approved")
	ErrTokenIdAlreadyExists       = errors.New("token id already exists")
	ErrReadonly                   = errors.New("teller is readonly")
	ErrSpoofedRealm               = errors.New("rlm does not match the current crossing frame")
	ErrNotRealm                   = errors.New("rlm must be a realm (got EOA/origin)")
	ErrInvalidName                = errors.New("invalid token name (empty, too long, or contains control chars)")
	ErrInvalidSymbol              = errors.New("invalid token symbol (empty, too long, or contains chars outside [A-Za-z0-9_-])")
)

// Construction limits. Symbol is restricted to the same charset as
// grc20reg.validateSlug because Token.ID() = origRealm + "." + symbol is
// emitted in events and frequently used as a registry slug; banning `.`
// `/` and whitespace here prevents downstream parsers from being fooled
// by ambiguous IDs (e.g. an attacker emitting tokens whose ID parses as
// belonging to a victim realm). Name is for display only and allows
// any valid UTF-8 except control characters.
const (
	MaxNameLen   = 64
	MaxSymbolLen = 11
)

const (
	MintEvent           = "Mint"
	BurnEvent           = "Burn"
	TransferEvent       = "Transfer"
	ApprovalEvent       = "Approval"
	ApprovalForAllEvent = "ApprovalForAll"
)

type fnTeller struct {
	accountFn func(_ int, rlm realm) address
	*Token
}

var _ Teller = (*fnTeller)(nil)

// IsCanonicalTeller reports whether t is the canonical *fnTeller produced by
// Token.CallerTeller / RealmTeller / RealmSubTeller / ReadonlyTeller /
// ImpersonateTeller. Use this at any public entry point that accepts a
// Teller from an external caller before invoking its methods.
//
// Foreign types — including embedding-based wrappers like
// `type Evil struct { grc721.Teller }` — are rejected because type
// assertions are nominal: *Evil is not *fnTeller, regardless of method
// promotion. This is the reliable defense; the unexported-marker "seal"
// pattern is bypassable via embedding (see
// p/test/seal/filetests/z_seal_iface_embedding_filetest.gno).
//
// Mirrors the precedent of chain/banker.IsCanonical and
// p/jaekwon/allowancesender's canonical-impl check.
func IsCanonicalTeller(t Teller) bool {
	_, ok := t.(*fnTeller)
	return ok
}
