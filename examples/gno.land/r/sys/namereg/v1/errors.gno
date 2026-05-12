package namereg

import (
	"errors"

	"gno.land/p/nt/ufmt/v0"
)

var (
	ErrNonUserCall     = errors.New("r/gnoland/users: non-user call")
	ErrPaused          = errors.New("r/gnoland/users: paused")
	ErrInvalidUsername = errors.New("r/gnoland/users: invalid username")

	// ErrInvalidPayment is the sentinel for "OriginSend amount didn't
	// match registerPrice." It deliberately omits the price from its
	// message — the price is read at panic time via errInvalidPayment()
	// below, so users see the CURRENT price even after governance has
	// changed it. Tests that match against this error use it as a
	// substring check via uassert.AbortsWithMessage. (audit finding #13)
	ErrInvalidPayment = errors.New("r/gnoland/users: invalid payment amount")
)

// errInvalidPayment returns the panic value for an OriginSend mismatch.
// Constructed lazily so the formatted price reflects the current value
// of registerPrice rather than the value frozen at package init time.
// Replaces the old `ErrInvalidPayment = ufmt.Errorf(...)` package var
// (audit finding #13) which captured registerPrice once and silently
// drifted out of date after every ProposeNewRegisterPrice execution.
func errInvalidPayment() error {
	return ufmt.Errorf("%s: must send exactly %d ugnot",
		ErrInvalidPayment.Error(), registerPrice)
}
