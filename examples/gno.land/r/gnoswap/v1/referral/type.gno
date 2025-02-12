package referral

import "std"

// zeroAddress represents an empty address used for validation and comparison.
var zeroAddress = std.Address("")

// Event types for each referral actions.
const (
	EventTypeRegister = "RegisterReferral"
	EventTypeUpdate   = "UpdateReferral"
	EventTypeRemove   = "RemoveReferral"
)

// ReferralKeeper defines the interface for managing referral relationships.
type ReferralKeeper interface {
	// register creates a new refferal relationship betwwen address and referral address.
	// returns an error if the addresses are invalid or if the caller is not authorized.
	register(addr, refAddr std.Address) error

	// update updates the referral address for a given address.
	// returns an error if the addresses are invalid or if the caller is not authorized.
	update(addr, newRefAddr std.Address) error

	// remove removes the referral relationship for a given address.
	// returns an error if the address is invalid or if the caller is not authorized.
	remove(addr std.Address) error

	// has checks if a referral relationship exists for a given address.
	has(addr std.Address) bool

	// get retrieves the referral address for a given address.
	// returns an error if the address is invalid or if the referral relationship does not exist.
	get(addr std.Address) (std.Address, error)

	// isEmpty checks if the referral relationship is empty.
	isEmpty() bool
}
