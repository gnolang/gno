package grc20

import (
	"std"

	"gno.land/p/demo/grc/exts"
)

type IGRC20 interface {
	exts.TokenMetadata

	// Returns the amount of tokens in existence.
	TotalSupply() uint64

	// Returns the amount of tokens owned by `account`.
	BalanceOf(account std.Address) (uint64, error)

	// Moves `amount` tokens from the caller's account to `to`.
	//
	// Returns an error if the operation failed.
	Transfer(to std.Address, amount uint64) error

	// Returns the remaining number of tokens that `spender` will be
	// allowed to spend on behalf of `owner` through {transferFrom}. This is
	// zero by default.
	//
	// This value changes when {approve} or {transferFrom} are called.
	Allowance(owner, spender std.Address) (uint64, error)

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

// Emitted when `value` tokens are moved from one account (`from`) to another (`to`).
//
// Note that `value` may be zero.
type TransferEvent struct {
	From  std.Address
	To    std.Address
	Value uint64
}

// Emitted when the allowance of a `spender` for an `owner` is set by
// a call to {approve}. `value` is the new allowance.
type ApprovalEvent struct {
	Owner   std.Address
	Spender std.Address
	Value   uint64
}
