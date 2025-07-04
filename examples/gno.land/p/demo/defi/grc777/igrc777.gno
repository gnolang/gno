package grc777

import (
	"std"
)

// TODO: use big.Int or a custom int64 instead of int64

type IGRC777 interface {
	// Returns the name of the token.
	GetName() string

	// Returns the symbol of the token, usually a shorter version of the
	// name.
	GetSymbol() string

	// Returns the decimals places of the token.
	GetDecimals() int

	// Returns the smallest part of the token that is not divisible. This
	// means all token operations (creation, movement and destruction) must
	// have amounts that are a multiple of this number.
	//
	// For most token contracts, this value will equal 1.
	Granularity() (granularity int64)

	// Returns the amount of tokens in existence.
	TotalSupply() (supply int64)

	// Returns the amount of tokens owned by an account (`owner`).
	BalanceOf(address_XXX std.Address) int64

	// Moves `amount` tokens from the caller's account to `recipient`.
	//
	// If send or receive hooks are registered for the caller and `recipient`,
	// the corresponding functions will be called with `data` and empty
	// `operatorData`. See {IERC777Sender} and {IERC777Recipient}.
	//
	// Emits a {Sent} event.
	//
	// Requirements
	//
	// - the caller must have at least `amount` tokens.
	// - `recipient` cannot be the zero address.
	// - if `recipient` is a contract, it must implement the {IERC777Recipient}
	// interface.
	Send(recipient std.Address, amount int64, data []byte)

	// Destroys `amount` tokens from the caller's account, reducing the
	// total supply.
	//
	// If a send hook is registered for the caller, the corresponding function
	// will be called with `data` and empty `operatorData`. See {IERC777Sender}.
	//
	// Emits a {Burned} event.
	//
	// Requirements
	//
	// - the caller must have at least `amount` tokens.
	Burn(amount int64, data []byte)

	// Returns true if an account is an operator of `tokenHolder`.
	// Operators can send and burn tokens on behalf of their owners. All
	// accounts are their own operator.
	//
	// See {operatorSend} and {operatorBurn}.
	IsOperatorFor(operator, tokenHolder std.Address) bool

	// Make an account an operator of the caller.
	//
	// See {isOperatorFor}.
	//
	// Emits an {AuthorizedOperator} event.
	//
	// Requirements
	//
	// - `operator` cannot be calling address.
	AuthorizeOperator(operator std.Address)

	// Revoke an account's operator status for the caller.
	//
	// See {isOperatorFor} and {defaultOperators}.
	//
	// Emits a {RevokedOperator} event.
	//
	// Requirements
	//
	// - `operator` cannot be calling address.
	RevokeOperator(operators std.Address)

	// Returns the list of default operators. These accounts are operators
	// for all token holders, even if {authorizeOperator} was never called on
	// them.
	//
	// This list is immutable, but individual holders may revoke these via
	// {revokeOperator}, in which case {isOperatorFor} will return false.
	DefaultOperators() []std.Address

	// Moves `amount` tokens from `sender` to `recipient`. The caller must
	// be an operator of `sender`.
	//
	// If send or receive hooks are registered for `sender` and `recipient`,
	// the corresponding functions will be called with `data` and
	// `operatorData`. See {IERC777Sender} and {IERC777Recipient}.
	//
	// Emits a {Sent} event.
	//
	// Requirements
	//
	// - `sender` cannot be the zero address.
	// - `sender` must have at least `amount` tokens.
	// - the caller must be an operator for `sender`.
	// - `recipient` cannot be the zero address.
	// - if `recipient` is a contract, it must implement the {IERC777Recipient}
	// interface.
	OperatorSend(sender, recipient std.Address, amount int64, data, operatorData []byte)

	// Destroys `amount` tokens from `account`, reducing the total supply.
	// The caller must be an operator of `account`.
	//
	// If a send hook is registered for `account`, the corresponding function
	// will be called with `data` and `operatorData`. See {IERC777Sender}.
	//
	// Emits a {Burned} event.
	//
	// Requirements
	//
	// - `account` cannot be the zero address.
	// - `account` must have at least `amount` tokens.
	// - the caller must be an operator for `account`.
	OperatorBurn(account std.Address, amount int64, data, operatorData []byte)
}

// Emitted when `amount` tokens are created by `operator` and assigned to `to`.
//
// Note that some additional user `data` and `operatorData` can be logged in the event.
type MintedEvent struct {
	Operator     std.Address
	To           std.Address
	Amount       int64
	Data         []byte
	OperatorData []byte
}

// Emitted when `operator` destroys `amount` tokens from `account`.
//
// Note that some additional user `data` and `operatorData` can be logged in the event.
type BurnedEvent struct {
	Operator     std.Address
	From         std.Address
	Amount       int64
	Data         []byte
	OperatorData []byte
}

// Emitted when `operator` is made operator for `tokenHolder`
type AuthorizedOperatorEvent struct {
	Operator    std.Address
	TokenHolder std.Address
}

// Emitted when `operator` is revoked its operator status for `tokenHolder`.
type RevokedOperatorEvent struct {
	Operator    std.Address
	TokenHolder std.Address
}

type SentEvent struct {
	Operator     std.Address
	From         std.Address
	To           std.Address
	Amount       int64
	Data         []byte
	OperatorData []byte
}
