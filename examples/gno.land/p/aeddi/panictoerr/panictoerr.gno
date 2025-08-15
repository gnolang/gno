package panictoerr

import (
	"errors"

	"gno.land/p/nt/ufmt"
)

// PanicToError executes a function that might panic and, if it does,
// recovers the panic and converts it to an error.
func PanicToError(mightPanic func()) (err error) {
	// Catch any panic that might occur and convert it to an error.
	defer func() {
		if r := recover(); r != nil {
			err = anyToError(r)
		}
	}()

	// Execute the function that might panic.
	mightPanic()

	return nil
}

// AbortToError executes a function that might abort, and if it does,
// revives the abort and converts it to an error.
func AbortToError(mightAbort func()) error {
	// Catch any abort that might occur and convert it to an error.
	if r := revive(mightAbort); r != nil {
		return anyToError(r)
	}

	return nil
}

// PanicAbortToError executes a function that might either panic or abort,
// and if it does, it recovers the panic or revives the abort and converts
// it to an error.
func PanicAbortToError(mightPanicOrAbort func()) error {
	var panicErr error

	// Catch any panic or abort that might occur and convert it to an error.
	if abortErr := AbortToError(func() {
		panicErr = PanicToError(mightPanicOrAbort)
	}); abortErr != nil {
		return abortErr
	}

	return panicErr
}

// anyToError converts any value to an error.
func anyToError(v any) error {
	switch v := v.(type) {
	case string:
		return errors.New(v)
	case error:
		return v
	default:
		return errors.New(ufmt.Sprint(v))
	}
}
