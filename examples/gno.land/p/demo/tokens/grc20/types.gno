package grc20

import (
	"errors"

	"gno.land/p/nt/avl/v0"
)

// Teller interface defines the methods that a GRC20 token must implement. It
// extends the TokenMetadata interface to include methods for managing token
// transfers, allowances, and querying balances.
//
// The Teller interface is designed to ensure that any token adhering to this
// standard provides a consistent API for interacting with fungible tokens.
//
// SECURITY: Transfer/Approve/TransferFrom take (_ int, rlm realm, ...), so
// handing a Teller value to untrusted code yields a capability token to
// whatever Transfer/Approve/TransferFrom impl that code dispatches into.
// Any /p/ or /r/ function that accepts a Teller as a parameter from external
// callers MUST type-assert against the canonical concrete type (*fnTeller)
// via IsCanonicalTeller and reject otherwise. An unexported-marker "seal"
// does NOT defend against this — see
// p/test/seal/filetests/z_seal_iface_embedding_filetest.gno
// for the realistic embedding-bypass attack. Reference impl for the
// canonical-allowlist pattern: p/jaekwon/allowancesender.
type Teller interface {
	// Returns the name of the token.
	GetName() string

	// Returns the symbol of the token, usually a shorter version of the
	// name.
	GetSymbol() string

	// Returns the decimals places of the token.
	GetDecimals() int

	// Returns the amount of tokens in existence.
	TotalSupply() int64

	// Returns the amount of tokens owned by `account`.
	BalanceOf(account address) int64

	// Moves `amount` tokens from the caller's account to `to`. rlm must
	// be the caller's own captured cur — verified via rlm.IsCurrent().
	//
	// Returns an error if the operation failed.
	Transfer(_ int, rlm realm, to address, amount int64) error

	// Returns the remaining number of tokens that `spender` will be
	// allowed to spend on behalf of `owner` through {transferFrom}. This is
	// zero by default.
	//
	// This value changes when {approve} or {transferFrom} are called.
	Allowance(owner, spender address) int64

	// Sets `amount` as the allowance of `spender` over the caller's tokens.
	//
	// Returns an error if the operation failed.
	//
	// IMPORTANT: Beware that changing an allowance with this method brings
	// the risk that someone may use both the old and the new allowance by
	// unfortunate transaction ordering. One possible solution to mitigate
	// this race condition is to first reduce the spender's allowance to 0
	// and set the desired value afterwards:
	// https://github.com/ethereum/EIPs/issues/20#issuecomment-263524729
	Approve(_ int, rlm realm, spender address, amount int64) error

	// Moves `amount` tokens from `from` to `to` using the
	// allowance mechanism. `amount` is then deducted from the caller's
	// allowance.
	//
	// Returns an error if the operation failed.
	TransferFrom(_ int, rlm realm, from, to address, amount int64) error
}

// Token represents a fungible token with a name, symbol, and a certain number
// of decimal places. It maintains a ledger for tracking balances and allowances
// of addresses.
//
// The Token struct provides methods for retrieving token metadata, such as the
// name, symbol, and decimals, as well as methods for interacting with the
// ledger, including checking balances and allowances.
type Token struct {
	// Name of the token (e.g., "Dummy Token").
	name string
	// Symbol of the token (e.g., "DUMMY").
	symbol string
	// Number of decimal places used for the token's precision.
	decimals int
	// Original realm of the token (e.g., "gno.land/r/demo/foo20").
	origRealm string
	// Pointer to the PrivateLedger that manages balances and allowances.
	ledger *PrivateLedger
}

// PrivateLedger is a struct that holds the balances and allowances for the
// token. It provides administrative functions for minting, burning,
// transferring tokens, and managing allowances.
//
// The PrivateLedger is not safe to expose publicly, as it contains sensitive
// information regarding token balances and allowances, and allows direct,
// unrestricted access to all administrative functions.
type PrivateLedger struct {
	// Total supply of the token managed by this ledger.
	totalSupply int64
	// chain.Address -> int64
	balances avl.Tree
	// owner.(chain.Address)+":"+spender.(chain.Address)) -> int64
	allowances avl.Tree
	// Pointer to the associated Token struct
	token *Token
}

var (
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrInsufficientAllowance = errors.New("insufficient allowance")
	ErrInvalidAddress        = errors.New("invalid address")
	ErrCannotTransferToSelf  = errors.New("cannot send transfer to self")
	ErrReadonly              = errors.New("banker is readonly")
	ErrRestrictedTokenOwner  = errors.New("restricted to bank owner")
	ErrMintOverflow          = errors.New("mint overflow")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrSpoofedRealm          = errors.New("rlm does not match the current crossing frame")
	ErrNotRealm              = errors.New("rlm must be a realm (got EOA/origin)")
	ErrInvalidName           = errors.New("invalid token name (empty, too long, or contains control chars)")
	ErrInvalidSymbol         = errors.New("invalid token symbol (empty, too long, or contains chars outside [A-Za-z0-9_-])")
	ErrInvalidDecimals       = errors.New("invalid decimals (must be 0..18)")
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
	MaxDecimals  = 18
)

const (
	MintEvent     = "Mint"
	BurnEvent     = "Burn"
	TransferEvent = "Transfer"
	ApprovalEvent = "Approval"
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
// `type Evil struct { grc20.Teller }` — are rejected because type
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
