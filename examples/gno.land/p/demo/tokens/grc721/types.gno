package grc721

import (
	"errors"

	"gno.land/p/nt/avl/v0"
)

type TokenID string

func (t TokenID) String() string { return string(t) }

// A Teller is a capability that acts as some account whenever it writes.
// IsCanonicalTeller confirms a Teller was minted by this package — an embedding
// forgery fails the check — but it does NOT reveal which account the Teller acts
// as: a caller-scoped CallerTeller and the admin-grade ImpersonateTeller (which
// acts as an arbitrary address) are both canonical. So it is an authenticity
// check on the implementation, not an authorization guard on the acting account;
// a caller accepting a Teller from outside must still establish, out of band, what
// account that Teller is entitled to act as.
// safeTransferFrom is omitted: EIP-721's receiver check needs a registry, and aliasing it would imply false safety.
type Teller interface {
	GetName() string
	GetSymbol() string
	ID() string
	TotalSupply() int64
	BalanceOf(owner address) int64
	OwnerOf(tid TokenID) (address, error)
	GetApproved(tid TokenID) (address, error)
	IsApprovedForAll(owner, operator address) bool

	Approve(_ int, rlm realm, to address, tid TokenID) error
	SetApprovalForAll(_ int, rlm realm, operator address, approved bool) error
	TransferFrom(_ int, rlm realm, from, to address, tid TokenID) error
}

// Hooks fire on mint/transfer/burn; no rlm params (cannot capture cur). Attach only via RegisterExtension.
type Extension interface {
	ExtensionKind() string // unique; duplicates rejected at register
	OnMint(to address, tid TokenID)
	OnTransfer(from, to address, tid TokenID)
	OnBurn(tid TokenID)
}

type ExtensionView interface {
	ExtensionKind() string
	TokenID() string // core Token.ID
}

type Token struct {
	id     string // origRealm + "." + symbol + "." + id
	name   string
	symbol string
	ledger *PrivateLedger
}

type PrivateLedger struct {
	token             *Token
	totalSupply       int64
	owners            avl.Tree // TokenID -> owner address
	balances          avl.Tree // owner address -> int64
	tokenApprovals    avl.Tree // TokenID -> approved address
	operatorApprovals avl.Tree // "owner:operator" -> bool
	extensions        []Extension
}

var (
	ErrInvalidTokenId             = errors.New("invalid token id")
	ErrInvalidAddress             = errors.New("invalid address")
	ErrTokenIdNotApproved         = errors.New("token id not approved for anyone")
	ErrApprovalToCurrentOwner     = errors.New("approval to current owner")
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

// Symbol charset matches grc20reg slug (embedded in Token.ID / events).
const (
	MaxNameLen   = 64
	MaxSymbolLen = 11
)

const (
	// NewToken announces every token creation; see NewToken for why it is a
	// complete provenance signal.
	NewTokenEvent = "NewToken"
	// Mint emits from empty addr; burn emits to empty addr (EIP-721).
	TransferEvent       = "Transfer"
	ApprovalEvent       = "Approval"
	ApprovalForAllEvent = "ApprovalForAll"
)

var zeroAddress = address("")

type fnTeller struct {
	accountFn func(_ int, rlm realm) address
	*Token
}

var _ Teller = (*fnTeller)(nil)

// IsCanonicalTeller reports whether t was minted by this package, rejecting Tellers
// forged by embedding *fnTeller in a wrapper. It does NOT discriminate by capability
// grade: CallerTeller, ReadonlyTeller, RealmTeller, RealmSubTeller and the admin-grade
// ImpersonateTeller all return true. Do not treat a true result as proof that a Teller
// is a caller-scoped capability — establish the acting account separately.
func IsCanonicalTeller(t Teller) bool {
	_, ok := t.(*fnTeller)
	return ok
}
