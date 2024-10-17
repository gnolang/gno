package grc20

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/exts"
)

var (
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrInsufficientAllowance = errors.New("insufficient allowance")
	ErrInvalidAddress        = errors.New("invalid address")
	ErrCannotTransferToSelf  = errors.New("cannot send transfer to self")
	ErrReadonly              = errors.New("banker is readonly")
	ErrRestrictedBankOwner   = errors.New("restricted to bank owner")
)

const (
	MintEvent     = "Mint"
	BurnEvent     = "Burn"
	TransferEvent = "Transfer"
	ApprovalEvent = "Approval"
)

// Bank represents a token's bank, including its name, symbol,
// decimals, and administrative functions.
type Bank struct {
	name     string
	symbol   string
	decimals uint
	adm      *AdminBanker
}

// AdminBanker is a struct that holds administrative functions for the token.
type AdminBanker struct {
	totalSupply uint64
	balances    avl.Tree // std.Address -> uint64
	allowances  avl.Tree // owner.(std.Address)+":"+spender.(std.Address)) -> uint64
}

// GRC20 interface defines the methods that a GRC20 token must implement.
type GRC20 interface {
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

type fnBanker struct {
	accountFn func() std.Address
	*Bank
}

var _ GRC20 = (*fnBanker)(nil)

// adminBanker is a struct that holds administrative functions for the token.
type adminBanker struct {
	bank *Bank
}

// accountSlugAddr returns the address derived from the specified address and slug.
func accountSlugAddr(addr std.Address, slug string) std.Address {
	// XXX: use a new `std.XXX` call for this.
	if slug == "" {
		return addr
	}
	key := addr.String() + "/" + slug
	return std.DerivePkgAddr(key) // temporarily using this helper
}
