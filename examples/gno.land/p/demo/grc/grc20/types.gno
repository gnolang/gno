package grc20

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/exts"
)

// Teller interface defines the methods that a GRC20 token must implement. It
// extends the TokenMetadata interface to include methods for managing token
// transfers, allowances, and querying balances.
//
// The Teller interface is designed to ensure that any token adhering to this
// standard provides a consistent API for interacting with fungible tokens.
type Teller interface {
	exts.TokenMetadata

	// Returns the amount of tokens in existence.
	TotalSupply() uint64

	// Returns the amount of tokens owned by `account`.
	BalanceOf(account std.Address) uint64

	// Moves `amount` tokens from the caller's account to `to`.
	//
	// Returns an error if the operation failed.
	Transfer(to std.Address, amount uint64) error

	// Returns the remaining number of tokens that `spender` will be
	// allowed to spend on behalf of `owner` through {transferFrom}. This is
	// zero by default.
	//
	// This value changes when {approve} or {transferFrom} are called.
	Allowance(owner, spender std.Address) uint64

	// Sets `amount` as the allowance of `spender` over the caller's tokens.
	//
	// Returns an error if the operation failed.
	//
	// IMPORTANT: Beware that changing an allowance with this method brings the risk
	// that someone may use both the old and the new allowance by unfortunate
	// transaction ordering. One possible solution to mitigate this race
	// condition is to first reduce the spender's allowance to 0 and set the
	// desired value afterwards:
	// https://github.com/ethereum/EIPs/issues/20#issuecomment-263524729
	Approve(spender std.Address, amount uint64) error

	// Moves `amount` tokens from `from` to `to` using the
	// allowance mechanism. `amount` is then deducted from the caller's
	// allowance.
	//
	// Returns an error if the operation failed.
	TransferFrom(from, to std.Address, amount uint64) error
}

// Token represents a fungible token with a name, symbol, and a certain number
// of decimal places. It maintains a ledger for tracking balances and allowances
// of addresses.
//
// The Token struct provides methods for retrieving token metadata, such as the
// name, symbol, and decimals, as well as methods for interacting with the
// ledger, including checking balances and allowances.
type Token struct {
	name     string         // Name of the token (e.g., "Dummy Token").
	symbol   string         // Symbol of the token (e.g., "DUMMY").
	decimals uint           // Number of decimal places used for the token's precision.
	ledger   *PrivateLedger // Pointer to the PrivateLedger that manages balances and allowances.
}

// PrivateLedger is a struct that holds the balances and allowances for the
// token. It provides administrative functions for minting, burning,
// transferring tokens, and managing allowances.
//
// The PrivateLedger is not safe to expose publicly, as it contains sensitive
// information regarding token balances and allowances, and allows direct,
// unrestricted access to all administrative functions.
type PrivateLedger struct {
	totalSupply uint64   // Total supply of the token managed by this ledger.
	balances    avl.Tree // std.Address -> uint64
	allowances  avl.Tree // owner.(std.Address)+":"+spender.(std.Address)) -> uint64
	token       *Token   // Pointer to the associated Token struct
}

var (
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrInsufficientAllowance = errors.New("insufficient allowance")
	ErrInvalidAddress        = errors.New("invalid address")
	ErrCannotTransferToSelf  = errors.New("cannot send transfer to self")
	ErrReadonly              = errors.New("banker is readonly")
	ErrRestrictedTokenOwner  = errors.New("restricted to bank owner")
	ErrOverflow              = errors.New("Mint overflow")
)

const (
	MintEvent     = "Mint"
	BurnEvent     = "Burn"
	TransferEvent = "Transfer"
	ApprovalEvent = "Approval"
)

type fnTeller struct {
	accountFn func() std.Address
	*Token
}

var _ Teller = (*fnTeller)(nil)
