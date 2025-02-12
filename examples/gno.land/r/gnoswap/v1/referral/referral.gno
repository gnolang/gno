package referral

import (
	"std"
)

// Referral provides the main interface for managing referral relationships.
// It encapsulates a `ReferralKeeper` instance which handles the actual storage
// and action of referral functionality.
type Referral struct {
	keeper ReferralKeeper
}

// NewReferral creates and returns a new instance of `Referral`.
// it initializes the underlying keeper with a new storage instance.
func NewReferral() *Referral {
	return &Referral{
		keeper: NewKeeper(),
	}
}

// Register creates a new referral relationship between addr and refAddr.
// It validates both addresses and ensures the caller has the required permissions.
// Returns an error if the addresses are invalid or if the caller is unauthorized.
func (r *Referral) Register(addr, refAddr std.Address) error {
	return r.keeper.register(addr, refAddr)
}

// Update modifies an existing referral relationship for the given address.
// The new referral address will replace any existing referral.
// Returns an error if the addresses are invalid or if the caller is unauthorized.
func (r *Referral) Update(addr, newAddr std.Address) error {
	return r.keeper.update(addr, newAddr)
}

// Remove deletes the referral relationship for the given address.
// If no referral exists for the address, the operation is a no-op.
// Returns an error if the address is invalid or if the caller is unauthorized.
func (r *Referral) Remove(addr std.Address) error {
	return r.keeper.remove(addr)
}

// Has checks if a referral relationship exists for the given address.
// Returns true if a referral exists, false otherwise.
func (r *Referral) Has(addr std.Address) bool {
	return r.keeper.has(addr)
}

// Get retrieves the referral address for the given address.
// Returns the referral address and nil error if found.
// Returns zeroAddress and an error if the address is invalid or no referral exists.
func (r *Referral) Get(addr std.Address) (std.Address, error) {
	return r.keeper.get(addr)
}
