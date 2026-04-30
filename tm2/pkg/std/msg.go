package std

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// Transactions messages must fulfill the Msg.
type Msg interface {
	// Return the message type.
	// Must be alphanumeric or empty.
	Route() string

	// Returns a human-readable string for the message, intended for utilization
	// within tags
	Type() string

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error

	// Get the canonical byte representation of the Msg.
	GetSignBytes() []byte

	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []crypto.Address
}

// SpendEstimator is an optional interface a Msg can implement to declare
// the coin outflow it expects to cause for a given signer. The auth ante's
// session pre-check aggregates these estimates across all msgs in a tx
// and rejects session-signed txs whose gas fee plus declared outflow
// would exceed the session's remaining SpendLimit — before gas is charged.
//
// Returning zero or nil means "this msg does not declare outflow for that
// signer." Msgs that don't implement this interface are skipped in the
// pre-check; the bank.Keeper.SendCoins session hook still catches the
// actual outflow at execution time. So implementing SpendEstimator is a
// gas-efficiency optimization, not a correctness requirement.
type SpendEstimator interface {
	SpendForSigner(signer crypto.Address) Coins
}
